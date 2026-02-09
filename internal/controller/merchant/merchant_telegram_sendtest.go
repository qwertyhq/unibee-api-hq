package merchant

import (
	"context"

	_interface "unibee/internal/interface/context"
	"unibee/internal/logic/telegram"

	_telegram "unibee/api/merchant/telegram"
)

// SendTest sends a test message to verify the Telegram bot configuration.
func (c *ControllerTelegram) SendTest(ctx context.Context, req *_telegram.SendTestReq) (res *_telegram.SendTestRes, err error) {
	merchantId := _interface.GetMerchantId(ctx)

	msg, sendErr := telegram.SendTestMessage(ctx, merchantId)
	if sendErr != nil {
		return &_telegram.SendTestRes{
			Success: false,
			Error:   sendErr.Error(),
		}, nil
	}

	messageId := 0
	if msg != nil {
		messageId = msg.ID
	}

	return &_telegram.SendTestRes{
		Success:   true,
		MessageId: messageId,
	}, nil
}
