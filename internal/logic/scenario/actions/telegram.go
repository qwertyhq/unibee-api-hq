package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/gogf/gf/v2/frame/g"

	"unibee/internal/logic/scenario"
	"unibee/internal/logic/telegram"
)

func init() {
	scenario.RegisterAction(scenario.StepSendTelegram, &TelegramAction{})
}

// TelegramAction sends a Telegram message, optionally with inline buttons.
type TelegramAction struct{}

func (a *TelegramAction) Execute(ctx context.Context, execCtx *scenario.ExecutionContext, step *scenario.StepDSL) (map[string]interface{}, error) {
	message, _ := step.Params["message"].(string)
	if message == "" {
		return nil, fmt.Errorf("send_telegram: message is required")
	}

	// Determine chat ID: from params, trigger data, or merchant config
	chatIdStr, _ := step.Params["chatId"].(string)
	if chatIdStr == "" {
		// Try trigger data (e.g., from bot_command / button_click)
		if cid, ok := execCtx.Variables["chat_id"]; ok {
			chatIdStr = cid
		}
	}

	// Get bot config
	cfg := telegram.GetConfig(ctx, execCtx.MerchantID)
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("send_telegram: bot not configured for merchant %d", execCtx.MerchantID)
	}

	// If still no chat ID, fall back to merchant's default
	if chatIdStr == "" {
		chatIdStr = cfg.ChatID
	}
	if chatIdStr == "" {
		return nil, fmt.Errorf("send_telegram: no chat ID available")
	}

	// Parse buttons
	var replyMarkup *models.InlineKeyboardMarkup
	if buttonsRaw, ok := step.Params["buttons"]; ok {
		buttons := parseButtons(buttonsRaw)
		if len(buttons) > 0 {
			var rows [][]models.InlineKeyboardButton
			for _, btn := range buttons {
				rows = append(rows, []models.InlineKeyboardButton{
					{
						Text:         btn.Text,
						CallbackData: fmt.Sprintf("sc_%d_%s", execCtx.MerchantID, btn.Action),
					},
				})
			}
			replyMarkup = &models.InlineKeyboardMarkup{InlineKeyboard: rows}
		}
	}

	// Create or get bot
	b, err := bot.New(cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("send_telegram: failed to create bot: %w", err)
	}

	params := &bot.SendMessageParams{
		Text:      message,
		ParseMode: models.ParseModeHTML,
	}

	if replyMarkup != nil {
		params.ReplyMarkup = replyMarkup
	}

	// Try numeric chat ID first
	chatID, err := strconv.ParseInt(chatIdStr, 10, 64)
	if err != nil {
		params.ChatID = chatIdStr
	} else {
		params.ChatID = chatID
	}

	sentMsg, err := b.SendMessage(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("send_telegram: %w", err)
	}

	output := map[string]interface{}{
		"message_id": sentMsg.ID,
		"chat_id":    chatIdStr,
	}

	g.Log().Infof(ctx, "scenario exec %d: sent telegram message to %s", execCtx.ExecutionID, chatIdStr)
	return output, nil
}

func parseButtons(raw interface{}) []scenario.ButtonDSL {
	var buttons []scenario.ButtonDSL

	switch v := raw.(type) {
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				btn := scenario.ButtonDSL{}
				if t, ok := m["text"].(string); ok {
					btn.Text = t
				}
				if a, ok := m["action"].(string); ok {
					btn.Action = a
				}
				if btn.Text != "" {
					buttons = append(buttons, btn)
				}
			}
		}
	case string:
		// Try JSON decode
		var arr []scenario.ButtonDSL
		if err := json.Unmarshal([]byte(v), &arr); err == nil {
			buttons = arr
		}
	}

	return buttons
}
