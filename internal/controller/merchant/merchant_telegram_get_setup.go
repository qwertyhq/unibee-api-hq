package merchant

import (
	"context"
	"strings"

	_interface "unibee/internal/interface/context"
	"unibee/internal/logic/telegram"

	_telegram "unibee/api/merchant/telegram"
)

// GetSetup retrieves the current Telegram bot configuration for the merchant.
func (c *ControllerTelegram) GetSetup(ctx context.Context, req *_telegram.GetSetupReq) (res *_telegram.GetSetupRes, err error) {
	merchantId := _interface.GetMerchantId(ctx)
	cfg := telegram.GetConfig(ctx, merchantId)

	// Mask the bot token for security
	maskedToken := ""
	if cfg.BotToken != "" {
		parts := strings.SplitN(cfg.BotToken, ":", 2)
		if len(parts) == 2 && len(parts[1]) > 4 {
			maskedToken = parts[0] + ":****" + parts[1][len(parts[1])-4:]
		} else {
			maskedToken = "****"
		}
	}

	return &_telegram.GetSetupRes{
		BotToken: maskedToken,
		ChatID:   cfg.ChatID,
		Enabled:  cfg.Enabled,
	}, nil
}
