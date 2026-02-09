package telegram

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"

	"unibee/internal/logic/merchant_config"
	"unibee/internal/logic/merchant_config/update"
)

const (
	ConfigKeyBotToken  = "telegram_bot_token"
	ConfigKeyChatID    = "telegram_chat_id"
	ConfigKeyEnabled   = "telegram_enabled"
	ConfigKeyTemplates = "telegram_templates"
)

var (
	botCache   = make(map[uint64]*bot.Bot)
	botCacheMu sync.RWMutex
)

// TelegramConfig holds telegram bot configuration for a merchant.
type TelegramConfig struct {
	BotToken string `json:"botToken"`
	ChatID   string `json:"chatID"`
	Enabled  bool   `json:"enabled"`
}

// TemplateConfig holds per-event template overrides for a merchant.
type TemplateConfig struct {
	Templates map[string]string `json:"templates"`
}

// GetConfig retrieves telegram configuration for a merchant.
func GetConfig(ctx context.Context, merchantId uint64) *TelegramConfig {
	cfg := &TelegramConfig{}

	if tokenCfg := merchant_config.GetMerchantConfig(ctx, merchantId, ConfigKeyBotToken); tokenCfg != nil {
		cfg.BotToken = tokenCfg.ConfigValue
	}
	if chatCfg := merchant_config.GetMerchantConfig(ctx, merchantId, ConfigKeyChatID); chatCfg != nil {
		cfg.ChatID = chatCfg.ConfigValue
	}
	if enabledCfg := merchant_config.GetMerchantConfig(ctx, merchantId, ConfigKeyEnabled); enabledCfg != nil {
		cfg.Enabled = enabledCfg.ConfigValue == "true"
	}

	return cfg
}

// SaveConfig persists telegram configuration for a merchant.
func SaveConfig(ctx context.Context, merchantId uint64, botToken string, chatID string, enabled bool) error {
	if err := update.SetMerchantConfig(ctx, merchantId, ConfigKeyBotToken, botToken); err != nil {
		return err
	}
	if err := update.SetMerchantConfig(ctx, merchantId, ConfigKeyChatID, chatID); err != nil {
		return err
	}
	enabledStr := "false"
	if enabled {
		enabledStr = "true"
	}
	if err := update.SetMerchantConfig(ctx, merchantId, ConfigKeyEnabled, enabledStr); err != nil {
		return err
	}

	// Invalidate bot cache
	botCacheMu.Lock()
	delete(botCache, merchantId)
	botCacheMu.Unlock()

	return nil
}

// GetTemplates retrieves custom template overrides for a merchant.
func GetTemplates(ctx context.Context, merchantId uint64) map[string]string {
	cfg := merchant_config.GetMerchantConfig(ctx, merchantId, ConfigKeyTemplates)
	if cfg == nil || cfg.ConfigValue == "" {
		return make(map[string]string)
	}

	j, err := gjson.DecodeToJson([]byte(cfg.ConfigValue))
	if err != nil {
		return make(map[string]string)
	}

	result := make(map[string]string)
	m := j.Map()
	for k, v := range m {
		if s, ok := v.(string); ok {
			result[k] = s
		}
	}
	return result
}

// SaveTemplates persists custom template overrides for a merchant.
func SaveTemplates(ctx context.Context, merchantId uint64, templates map[string]string) error {
	j := gjson.New(templates)
	return update.SetMerchantConfig(ctx, merchantId, ConfigKeyTemplates, j.MustToJsonString())
}

// GetTemplateForEvent returns the template for a specific event, checking merchant overrides first.
func GetTemplateForEvent(ctx context.Context, merchantId uint64, event string) string {
	// Check merchant custom templates
	customs := GetTemplates(ctx, merchantId)
	if tmpl, ok := customs[event]; ok && tmpl != "" {
		return tmpl
	}

	// Fall back to defaults
	if tmpl, ok := DefaultTemplates[event]; ok {
		return tmpl
	}

	// Generic fallback
	return "📌 {{event}}\nUser: {{userEmail}}"
}

// getOrCreateBot returns a cached bot instance or creates a new one.
func getOrCreateBot(merchantId uint64, token string) (*bot.Bot, error) {
	botCacheMu.RLock()
	if b, ok := botCache[merchantId]; ok {
		botCacheMu.RUnlock()
		return b, nil
	}
	botCacheMu.RUnlock()

	botCacheMu.Lock()
	defer botCacheMu.Unlock()

	// Double-check after lock
	if b, ok := botCache[merchantId]; ok {
		return b, nil
	}

	b, err := bot.New(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	botCache[merchantId] = b
	return b, nil
}

// SendNotification sends a rendered template message to the merchant's Telegram chat.
func SendNotification(ctx context.Context, merchantId uint64, event string, dataJson map[string]interface{}) error {
	cfg := GetConfig(ctx, merchantId)
	if !cfg.Enabled || cfg.BotToken == "" || cfg.ChatID == "" {
		return nil
	}

	tmpl := GetTemplateForEvent(ctx, merchantId, event)
	vars := BuildVariableMap(event, dataJson)
	message := RenderTemplate(tmpl, vars)

	if message == "" {
		return nil
	}

	b, err := getOrCreateBot(merchantId, cfg.BotToken)
	if err != nil {
		g.Log().Errorf(ctx, "telegram bot init error for merchant %d: %v", merchantId, err)
		return err
	}

	chatID, err := strconv.ParseInt(cfg.ChatID, 10, 64)
	if err != nil {
		// Try as string (channel username)
		_, sendErr := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: cfg.ChatID,
			Text:   message,
		})
		if sendErr != nil {
			g.Log().Errorf(ctx, "telegram send error for merchant %d: %v", merchantId, sendErr)
			return sendErr
		}
		return nil
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   message,
	})
	if err != nil {
		g.Log().Errorf(ctx, "telegram send error for merchant %d: %v", merchantId, err)
		return err
	}

	return nil
}

// SendTestMessage sends a test message to verify bot configuration.
func SendTestMessage(ctx context.Context, merchantId uint64) (*models.Message, error) {
	cfg := GetConfig(ctx, merchantId)
	if cfg.BotToken == "" || cfg.ChatID == "" {
		return nil, fmt.Errorf("telegram bot not configured: bot token or chat ID is empty")
	}

	b, err := getOrCreateBot(merchantId, cfg.BotToken)
	if err != nil {
		return nil, err
	}

	testMessage := "✅ UniBee Telegram Integration Test\n\nBot is connected and working correctly."

	chatID, err := strconv.ParseInt(cfg.ChatID, 10, 64)
	if err != nil {
		// Try as string (channel username)
		msg, sendErr := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: cfg.ChatID,
			Text:   testMessage,
		})
		return msg, sendErr
	}

	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   testMessage,
	})
	return msg, err
}
