package telegram

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Setup — configure Telegram bot token and chat ID
type SetupReq struct {
	g.Meta   `path:"/setup" tags:"Telegram" method:"post" summary:"Setup Telegram Bot"`
	BotToken string `json:"botToken" dc:"Telegram Bot Token from @BotFather" v:"required"`
	ChatID   string `json:"chatId" dc:"Telegram Chat ID or Channel Username" v:"required"`
	Enabled  bool   `json:"enabled" dc:"Enable Telegram notifications"`
}
type SetupRes struct {
}

// GetSetup — retrieve current Telegram configuration
type GetSetupReq struct {
	g.Meta `path:"/get_setup" tags:"Telegram" method:"get" summary:"Get Telegram Bot Setup"`
}
type GetSetupRes struct {
	BotToken string `json:"botToken" dc:"Bot Token (masked)"`
	ChatID   string `json:"chatId" dc:"Chat ID"`
	Enabled  bool   `json:"enabled" dc:"Is Enabled"`
}

// SendTest — send a test message
type SendTestReq struct {
	g.Meta `path:"/send_test" tags:"Telegram" method:"post" summary:"Send Test Telegram Message"`
}
type SendTestRes struct {
	Success   bool   `json:"success" dc:"Whether test message was sent"`
	MessageId int    `json:"messageId" dc:"Telegram Message ID"`
	Error     string `json:"error" dc:"Error message if failed"`
}

// TemplateList — get all templates (defaults + custom overrides)
type TemplateListReq struct {
	g.Meta `path:"/template_list" tags:"Telegram" method:"get" summary:"Get Telegram Message Templates"`
}
type TemplateListRes struct {
	Templates          []*TemplateItem `json:"templates" dc:"Template list"`
	AvailableVariables []string        `json:"availableVariables" dc:"Available template variables"`
}
type TemplateItem struct {
	Event          string `json:"event" dc:"Webhook event name"`
	Template       string `json:"template" dc:"Current template text"`
	IsCustom       bool   `json:"isCustom" dc:"Whether this is a custom override"`
	DefaultTemplate string `json:"defaultTemplate" dc:"Default template for reference"`
}

// TemplateUpdate — update a template for a specific event
type TemplateUpdateReq struct {
	g.Meta   `path:"/template_update" tags:"Telegram" method:"post" summary:"Update Telegram Message Template"`
	Event    string `json:"event" dc:"Webhook event name" v:"required"`
	Template string `json:"template" dc:"Template text with {{variables}}. Empty string resets to default." `
}
type TemplateUpdateRes struct {
}

// TemplatePreview — preview a rendered template with sample data
type TemplatePreviewReq struct {
	g.Meta   `path:"/template_preview" tags:"Telegram" method:"post" summary:"Preview Telegram Message Template"`
	Event    string `json:"event" dc:"Webhook event name" v:"required"`
	Template string `json:"template" dc:"Template text to preview" v:"required"`
}
type TemplatePreviewRes struct {
	RenderedMessage string   `json:"renderedMessage" dc:"Rendered message preview"`
	UsedVariables   []string `json:"usedVariables" dc:"Variables used in template"`
}
