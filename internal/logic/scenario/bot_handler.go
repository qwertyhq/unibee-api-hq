package scenario

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/gogf/gf/v2/frame/g"

	"unibee/internal/logic/telegram"
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

		// Handle built-in /start
		if command == "/start" {
			handleStartCommand(ctx, b, merchantId, msg)
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
		// Let the scenario engine handle it
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
