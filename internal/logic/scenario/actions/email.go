package actions

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	"unibee/internal/logic/email"
	"unibee/internal/logic/scenario"
)

func init() {
	scenario.RegisterAction(scenario.StepSendEmail, &EmailAction{})
}

// EmailAction sends an email through the merchant's configured email gateway.
type EmailAction struct{}

func (a *EmailAction) Execute(ctx context.Context, execCtx *scenario.ExecutionContext, step *scenario.StepDSL) (map[string]interface{}, error) {
	to := getStringParam(step.Params, "to", execCtx.Variables)
	if to == "" {
		return nil, fmt.Errorf("send_email: 'to' is required")
	}

	subject := getStringParam(step.Params, "subject", execCtx.Variables)
	if subject == "" {
		subject = "Notification"
	}

	body := getStringParam(step.Params, "body", execCtx.Variables)
	if body == "" {
		return nil, fmt.Errorf("send_email: 'body' is required")
	}

	req := &email.SendgridEmailReq{
		MerchantId: execCtx.MerchantID,
		MailTo:     to,
		Subject:    subject,
		Content:    body,
	}

	err := email.Send(ctx, req)
	if err != nil {
		g.Log().Errorf(ctx, "scenario: send_email to %s failed: %v", to, err)
		return nil, fmt.Errorf("send_email failed: %w", err)
	}

	g.Log().Infof(ctx, "scenario: email sent to %s, subject: %s", to, subject)
	return map[string]interface{}{
		"email_sent": "true",
		"email_to":   to,
	}, nil
}
