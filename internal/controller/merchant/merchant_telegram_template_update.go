package merchant

import (
	"context"

	_interface "unibee/internal/interface/context"
	"unibee/internal/logic/telegram"

	_telegram "unibee/api/merchant/telegram"
)

// TemplateUpdate updates or resets a message template for a specific event.
func (c *ControllerTelegram) TemplateUpdate(ctx context.Context, req *_telegram.TemplateUpdateReq) (res *_telegram.TemplateUpdateRes, err error) {
	merchantId := _interface.GetMerchantId(ctx)

	templates := telegram.GetTemplates(ctx, merchantId)
	if req.Template == "" {
		// Reset to default — remove custom override
		delete(templates, req.Event)
	} else {
		templates[req.Event] = req.Template
	}

	err = telegram.SaveTemplates(ctx, merchantId, templates)
	if err != nil {
		return nil, err
	}

	return &_telegram.TemplateUpdateRes{}, nil
}
