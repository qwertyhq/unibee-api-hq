package merchant

import (
	"context"

	_scenario "unibee/api/merchant/scenario"
)

func (c *ControllerScenario) TriggerList(ctx context.Context, req *_scenario.TriggerListReq) (res *_scenario.TriggerListRes, err error) {
	triggers := []*_scenario.TriggerType{
		{
			Type:        "webhook_event",
			Name:        "Webhook Event",
			Description: "Triggered by a billing system event (payment.success, subscription.cancelled, etc.)",
		},
		{
			Type:        "bot_command",
			Name:        "Bot Command",
			Description: "Triggered by a Telegram bot command (/start, /status, /help, or custom)",
		},
		{
			Type:        "button_click",
			Name:        "Button Click",
			Description: "Triggered when a user clicks an inline button in Telegram",
		},
		{
			Type:        "schedule",
			Name:        "Schedule (Cron)",
			Description: "Triggered on a cron schedule (e.g. every day at 9:00)",
		},
		{
			Type:        "manual",
			Name:        "Manual",
			Description: "Triggered manually from the admin panel",
		},
	}

	return &_scenario.TriggerListRes{Triggers: triggers}, nil
}
