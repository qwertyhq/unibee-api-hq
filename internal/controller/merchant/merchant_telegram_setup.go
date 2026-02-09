package merchant

import (
	"context"

	_interface "unibee/internal/interface/context"
	"unibee/internal/logic/telegram"

	_telegram "unibee/api/merchant/telegram"
)

// Setup configures the Telegram bot token and chat ID for the merchant.
func (c *ControllerTelegram) Setup(ctx context.Context, req *_telegram.SetupReq) (res *_telegram.SetupRes, err error) {
	merchantId := _interface.GetMerchantId(ctx)
	err = telegram.SaveConfig(ctx, merchantId, req.BotToken, req.ChatID, req.Enabled)
	if err != nil {
		return nil, err
	}
	return &_telegram.SetupRes{}, nil
}
