package scenario

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/gogf/gf/v2/frame/g"

	"unibee/internal/consts"
	"unibee/internal/logic/telegram"
	"unibee/internal/query"
	"unibee/utility"
)

var (
	// activeBots tracks running bot polling goroutines per merchant
	activeBots   = make(map[uint64]context.CancelFunc)
	activeBotsMu sync.Mutex
)

// StartBotPolling starts long polling for a merchant's Telegram bot.
// Non-blocking — runs in a goroutine.
func StartBotPolling(ctx context.Context, merchantId uint64) error {
	cfg := telegram.GetConfig(ctx, merchantId)
	if cfg.BotToken == "" {
		return fmt.Errorf("bot token not configured for merchant %d", merchantId)
	}

	activeBotsMu.Lock()
	// Stop existing polling if any
	if cancel, ok := activeBots[merchantId]; ok {
		cancel()
	}
	activeBotsMu.Unlock()

	pollCtx, cancel := context.WithCancel(ctx)

	activeBotsMu.Lock()
	activeBots[merchantId] = cancel
	activeBotsMu.Unlock()

	opts := []bot.Option{
		bot.WithDefaultHandler(func(bCtx context.Context, b *bot.Bot, update *models.Update) {
			handleUpdate(bCtx, b, merchantId, update)
		}),
	}

	b, err := bot.New(cfg.BotToken, opts...)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create bot for merchant %d: %w", merchantId, err)
	}

	go func() {
		g.Log().Infof(ctx, "scenario: starting bot polling for merchant %d", merchantId)
		b.Start(pollCtx)
		g.Log().Infof(ctx, "scenario: bot polling stopped for merchant %d", merchantId)
	}()

	return nil
}

// StopBotPolling stops long polling for a merchant's Telegram bot.
func StopBotPolling(merchantId uint64) {
	activeBotsMu.Lock()
	defer activeBotsMu.Unlock()

	if cancel, ok := activeBots[merchantId]; ok {
		cancel()
		delete(activeBots, merchantId)
	}
}

// StopAllBots stops all running bot polling goroutines.
func StopAllBots() {
	activeBotsMu.Lock()
	defer activeBotsMu.Unlock()

	for mid, cancel := range activeBots {
		cancel()
		delete(activeBots, mid)
	}
}

// handleUpdate processes a single Telegram update.
func handleUpdate(ctx context.Context, b *bot.Bot, merchantId uint64, update *models.Update) {
	if update.Message != nil && update.Message.Text != "" {
		handleMessage(ctx, b, merchantId, update.Message)
	}
	if update.CallbackQuery != nil {
		handleCallback(ctx, b, merchantId, update.CallbackQuery)
	}
}

// handleMessage processes a text message (bot command or text).
func handleMessage(ctx context.Context, b *bot.Bot, merchantId uint64, msg *models.Message) {
	text := strings.TrimSpace(msg.Text)

	// Track user
	username := ""
	firstName := ""
	lastName := ""
	if msg.From != nil {
		username = msg.From.Username
		firstName = msg.From.FirstName
		lastName = msg.From.LastName
	}
	_ = UpsertTelegramUser(ctx, merchantId, msg.Chat.ID, username, firstName, lastName)

	// Check if it's a command
	if strings.HasPrefix(text, "/") {
		parts := strings.SplitN(text, " ", 2)
		command := strings.ToLower(parts[0])

		// Handle built-in commands first
		switch command {
		case "/start":
			handleStartCommand(ctx, b, merchantId, msg)
			return
		case "/help":
			handleHelpCommand(ctx, b, merchantId, msg)
			return
		case "/status":
			handleStatusCommand(ctx, b, merchantId, msg)
			return
		case "/invoices":
			handleInvoicesCommand(ctx, b, merchantId, msg)
			return
		case "/plans":
			handlePlansCommand(ctx, b, merchantId, msg)
			return
		}

		// Match against scenario triggers
		MatchAndRunBotCommand(ctx, merchantId, command, msg.Chat.ID, username)
		return
	}

	// Non-command text — could be a freeform trigger in the future
}

// handleCallback processes an inline button callback query.
func handleCallback(ctx context.Context, b *bot.Bot, merchantId uint64, cq *models.CallbackQuery) {
	data := cq.Data

	// Answer callback to dismiss the loading indicator
	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: cq.ID,
	})

	// Parse callback data — format: "sc_{merchantId}_{action}"
	if strings.HasPrefix(data, "sc_") {
		parts := strings.SplitN(data, "_", 3)
		if len(parts) == 3 {
			action := parts[2]
			chatId := int64(0)
			username := cq.From.Username
			if cq.Message.Message != nil {
				chatId = cq.Message.Message.Chat.ID
			}

			MatchAndRunButtonClick(ctx, merchantId, action, chatId, username)
		}
	}
}

// handleStartCommand handles the built-in /start command.
func handleStartCommand(ctx context.Context, b *bot.Bot, merchantId uint64, msg *models.Message) {
	// Check if there are custom scenarios for /start
	scenarios, err := GetScenariosByTrigger(ctx, merchantId, TriggerBotCommand, "/start")
	if err == nil && len(scenarios) > 0 {
		username := ""
		if msg.From != nil {
			username = msg.From.Username
		}
		MatchAndRunBotCommand(ctx, merchantId, "/start", msg.Chat.ID, username)
		return
	}

	// Default /start response
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    msg.Chat.ID,
		Text:      "👋 Welcome! This bot is powered by UniBee.\n\nUse /help to see available commands.",
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		g.Log().Errorf(ctx, "scenario: failed to send /start response: %v", err)
	}
}

// handleHelpCommand sends the list of available commands.
func handleHelpCommand(ctx context.Context, b *bot.Bot, merchantId uint64, msg *models.Message) {
	text := "📖 <b>Available commands:</b>\n\n" +
		"/status — Your subscription status\n" +
		"/invoices — Recent invoices\n" +
		"/plans — Available plans\n" +
		"/help — Show this message"

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    msg.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		g.Log().Errorf(ctx, "scenario: failed to send /help: %v", err)
	}
}

// handleStatusCommand shows the user's current subscription status.
func handleStatusCommand(ctx context.Context, b *bot.Bot, merchantId uint64, msg *models.Message) {
	userId, err := resolveUserId(ctx, merchantId, msg.Chat.ID)
	if err != nil || userId == 0 {
		sendText(ctx, b, msg.Chat.ID, "⚠️ Your Telegram account is not linked to a UniBee user.\nPlease contact support to link your account.")
		return
	}

	subs := query.GetUserAllActiveOrIncompleteSubscriptions(ctx, userId, merchantId)
	if len(subs) == 0 {
		sendText(ctx, b, msg.Chat.ID, "ℹ️ You have no active subscriptions.")
		return
	}

	var lines []string
	lines = append(lines, "📊 <b>Your subscriptions:</b>\n")
	for i, sub := range subs {
		plan := query.GetPlanById(ctx, sub.PlanId)
		planName := "Unknown plan"
		if plan != nil {
			planName = plan.PlanName
		}

		statusEmoji := "🟢"
		statusText := subscriptionStatus(sub.Status)
		if sub.Status != consts.SubStatusActive {
			statusEmoji = "🟡"
		}
		if sub.Status == consts.SubStatusCancelled || sub.Status == consts.SubStatusExpired {
			statusEmoji = "🔴"
		}

		amount := utility.ConvertCentToDollarStr(sub.Amount, sub.Currency)

		line := fmt.Sprintf("%d. %s <b>%s</b>\n   %s %s — %s %s",
			i+1, statusEmoji, planName,
			amount, strings.ToUpper(sub.Currency),
			statusText, formatPeriod(sub.CurrentPeriodEnd))
		lines = append(lines, line)
	}

	sendText(ctx, b, msg.Chat.ID, strings.Join(lines, "\n"))
}

// handleInvoicesCommand shows the user's recent invoices.
func handleInvoicesCommand(ctx context.Context, b *bot.Bot, merchantId uint64, msg *models.Message) {
	userId, err := resolveUserId(ctx, merchantId, msg.Chat.ID)
	if err != nil || userId == 0 {
		sendText(ctx, b, msg.Chat.ID, "⚠️ Your Telegram account is not linked to a UniBee user.\nPlease contact support to link your account.")
		return
	}

	type invoiceRow struct {
		InvoiceId   string `json:"invoiceId"`
		TotalAmount int64  `json:"totalAmount"`
		Currency    string `json:"currency"`
		Status      int    `json:"status"`
		CreateTime  int64  `json:"createTime"`
	}

	var rows []*invoiceRow
	err = g.DB().Ctx(ctx).Raw(
		`SELECT invoice_id, total_amount, currency, status, create_time 
		 FROM invoice 
		 WHERE user_id = ? AND merchant_id = ? AND is_deleted = 0 
		 ORDER BY gmt_create DESC LIMIT 10`,
		userId, merchantId,
	).Scan(&rows)
	if err != nil || len(rows) == 0 {
		sendText(ctx, b, msg.Chat.ID, "ℹ️ No invoices found.")
		return
	}

	var lines []string
	lines = append(lines, "🧾 <b>Recent invoices:</b>\n")
	for i, inv := range rows {
		statusEmoji := "⏳"
		statusText := invoiceStatus(inv.Status)
		if inv.Status == 3 {
			statusEmoji = "✅"
		} else if inv.Status == 4 || inv.Status == 5 {
			statusEmoji = "❌"
		}

		amount := utility.ConvertCentToDollarStr(inv.TotalAmount, inv.Currency)
		line := fmt.Sprintf("%d. %s %s %s — %s",
			i+1, statusEmoji, amount, strings.ToUpper(inv.Currency), statusText)
		lines = append(lines, line)
	}

	sendText(ctx, b, msg.Chat.ID, strings.Join(lines, "\n"))
}

// handlePlansCommand shows available billing plans.
func handlePlansCommand(ctx context.Context, b *bot.Bot, merchantId uint64, msg *models.Message) {
	type planRow struct {
		PlanName      string `json:"planName"`
		Amount        int64  `json:"amount"`
		Currency      string `json:"currency"`
		IntervalUnit  string `json:"intervalUnit"`
		IntervalCount int    `json:"intervalCount"`
		Description   string `json:"description"`
	}

	var rows []*planRow
	err := g.DB().Ctx(ctx).Raw(
		`SELECT plan_name, amount, currency, interval_unit, interval_count, description
		 FROM plan 
		 WHERE merchant_id = ? AND status = 2 AND is_deleted = 0 AND publish_status = 1
		 ORDER BY amount ASC`,
		merchantId,
	).Scan(&rows)
	if err != nil || len(rows) == 0 {
		sendText(ctx, b, msg.Chat.ID, "ℹ️ No plans available at the moment.")
		return
	}

	var lines []string
	lines = append(lines, "📋 <b>Available plans:</b>\n")
	for i, p := range rows {
		amount := utility.ConvertCentToDollarStr(p.Amount, p.Currency)
		interval := fmt.Sprintf("%d %s", p.IntervalCount, p.IntervalUnit)
		if p.IntervalCount == 1 {
			interval = p.IntervalUnit
		}

		line := fmt.Sprintf("%d. <b>%s</b> — %s %s / %s",
			i+1, p.PlanName, amount, strings.ToUpper(p.Currency), interval)
		if p.Description != "" {
			line += "\n   " + p.Description
		}
		lines = append(lines, line)
	}

	sendText(ctx, b, msg.Chat.ID, strings.Join(lines, "\n"))
}

// ──── Bot Command Helpers ────

// resolveUserId finds the linked UniBee userId for a Telegram chat.
func resolveUserId(ctx context.Context, merchantId uint64, chatId int64) (uint64, error) {
	tgUser, err := GetTelegramUserByChatId(ctx, merchantId, chatId)
	if err != nil {
		return 0, err
	}
	if tgUser == nil || tgUser.UserId == 0 {
		return 0, nil
	}
	return tgUser.UserId, nil
}

// sendText is a convenience function to send an HTML-formatted message.
func sendText(ctx context.Context, b *bot.Bot, chatId int64, text string) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatId,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		g.Log().Errorf(ctx, "scenario: failed to send message to chat %d: %v", chatId, err)
	}
}

// subscriptionStatus returns a human-readable status string.
func subscriptionStatus(status int) string {
	switch status {
	case consts.SubStatusActive:
		return "Active"
	case consts.SubStatusCancelled:
		return "Cancelled"
	case consts.SubStatusExpired:
		return "Expired"
	case consts.SubStatusSuspended:
		return "Suspended"
	case consts.SubStatusIncomplete:
		return "Incomplete"
	default:
		return "Unknown"
	}
}

// invoiceStatus returns a human-readable invoice status.
func invoiceStatus(status int) string {
	switch status {
	case 1:
		return "Pending"
	case 2:
		return "Processing"
	case 3:
		return "Paid"
	case 4:
		return "Failed"
	case 5:
		return "Cancelled"
	case 6:
		return "Reversed"
	default:
		return "Unknown"
	}
}

// formatPeriod formats a Unix timestamp to a readable expiry string.
func formatPeriod(ts int64) string {
	if ts <= 0 {
		return ""
	}
	return fmt.Sprintf("(until %s)", strconv.FormatInt(ts, 10))
}

// InitAllBotPolling starts polling for all merchants that have enabled bots.
// Call this on application startup.
func InitAllBotPolling(ctx context.Context) {
	// We look up merchants that have telegram_bot_token configured.
	// This queries the merchant_config table.
	// For simplicity, we get all merchants from the scenario table that have bot_command triggers.
	merchantIds, err := getDistinctBotMerchantIds(ctx)
	if err != nil {
		g.Log().Errorf(ctx, "scenario: failed to load bot merchants on init: %v", err)
		return
	}

	for _, mid := range merchantIds {
		if err := StartBotPolling(ctx, mid); err != nil {
			g.Log().Errorf(ctx, "scenario: failed to start bot polling for merchant %d: %v", mid, err)
		}
	}

	g.Log().Infof(ctx, "scenario: initialized bot polling for %d merchants", len(merchantIds))
}

// getDistinctBotMerchantIds returns merchant IDs that have enabled bot_command or button_click scenarios.
func getDistinctBotMerchantIds(ctx context.Context) ([]uint64, error) {
	type row struct {
		MerchantId uint64 `json:"merchant_id"`
	}

	var list []*row
	err := g.DB().Ctx(ctx).Raw(`
		SELECT DISTINCT merchant_id 
		FROM merchant_scenario 
		WHERE enabled = 1 
		  AND is_deleted = 0 
		  AND trigger_type IN ('bot_command', 'button_click')
	`).Scan(&list)
	if err != nil {
		return nil, err
	}

	ids := make([]uint64, 0, len(list))
	for _, r := range list {
		ids = append(ids, r.MerchantId)
	}
	return ids, nil
}
