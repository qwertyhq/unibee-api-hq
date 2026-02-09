package merchant

import (
	"context"

	_interface "unibee/internal/interface/context"
	"unibee/internal/logic/telegram"

	_telegram "unibee/api/merchant/telegram"
)

// TemplateList returns all event templates (both defaults and custom overrides).
func (c *ControllerTelegram) TemplateList(ctx context.Context, req *_telegram.TemplateListReq) (res *_telegram.TemplateListRes, err error) {
	merchantId := _interface.GetMerchantId(ctx)
	customs := telegram.GetTemplates(ctx, merchantId)

	var items []*_telegram.TemplateItem
	for event, defaultTmpl := range telegram.DefaultTemplates {
		item := &_telegram.TemplateItem{
			Event:           event,
			Template:        defaultTmpl,
			IsCustom:        false,
			DefaultTemplate: defaultTmpl,
		}
		if customTmpl, ok := customs[event]; ok && customTmpl != "" {
			item.Template = customTmpl
			item.IsCustom = true
		}
		items = append(items, item)
	}

	// Add custom templates for events not in defaults
	for event, customTmpl := range customs {
		if _, exists := telegram.DefaultTemplates[event]; !exists {
			items = append(items, &_telegram.TemplateItem{
				Event:           event,
				Template:        customTmpl,
				IsCustom:        true,
				DefaultTemplate: "",
			})
		}
	}

	return &_telegram.TemplateListRes{
		Templates:          items,
		AvailableVariables: telegram.AvailableVariables,
	}, nil
}
