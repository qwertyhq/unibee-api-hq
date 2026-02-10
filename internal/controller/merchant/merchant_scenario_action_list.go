package merchant

import (
	"context"

	_scenario "unibee/api/merchant/scenario"
)

func (c *ControllerScenario) ActionList(ctx context.Context, req *_scenario.ActionListReq) (res *_scenario.ActionListRes, err error) {
	actions := []*_scenario.ActionType{
		{
			Type:        "send_telegram",
			Name:        "Send Telegram Message",
			Description: "Send a message to Telegram chat with optional inline buttons",
			Params:      []string{"message", "chatId (optional)", "buttons (optional)"},
		},
		{
			Type:        "http_request",
			Name:        "HTTP Request",
			Description: "Send HTTP request to any external API",
			Params:      []string{"method", "url", "headers (optional)", "body (optional)"},
		},
		{
			Type:        "delay",
			Name:        "Delay",
			Description: "Wait for a specified duration before continuing",
			Params:      []string{"duration (e.g. 30s, 5m, 1h, 1d)"},
		},
		{
			Type:        "condition",
			Name:        "Condition",
			Description: "If/then/else branching based on variable values",
			Params:      []string{"if", "then (step_id)", "else (step_id)"},
		},
		{
			Type:        "set_variable",
			Name:        "Set Variable",
			Description: "Set a variable value for use in subsequent steps",
			Params:      []string{"name", "value"},
		},
		{
			Type:        "unibee_api",
			Name:        "UniBee API Call",
			Description: "Call internal UniBee API actions (create invoice, cancel subscription, etc.)",
			Params:      []string{"action", "params"},
		},
		{
			Type:        "send_email",
			Name:        "Send Email",
			Description: "Send an email notification",
			Params:      []string{"to", "subject", "body"},
		},
		{
			Type:        "log",
			Name:        "Log",
			Description: "Write a log message for debugging",
			Params:      []string{"message", "level (info/warning/error)"},
		},
	}

	return &_scenario.ActionListRes{Actions: actions}, nil
}
