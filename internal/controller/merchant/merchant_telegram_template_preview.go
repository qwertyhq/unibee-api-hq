package merchant

import (
	"context"

	"unibee/internal/logic/telegram"

	_telegram "unibee/api/merchant/telegram"
)

// TemplatePreview renders a template with sample data for preview.
func (c *ControllerTelegram) TemplatePreview(ctx context.Context, req *_telegram.TemplatePreviewReq) (res *_telegram.TemplatePreviewRes, err error) {
	// Build sample data for preview
	sampleData := map[string]interface{}{
		"subscriptionId":         "sub_12345",
		"planId":                 "plan_67890",
		"planName":               "Pro Monthly",
		"userId":                 "usr_11111",
		"userEmail":              "user@example.com",
		"userName":               "John Doe",
		"firstName":              "John",
		"lastName":               "Doe",
		"amount":                 int64(2999),
		"currency":               "USD",
		"status":                 "active",
		"paymentId":              "pay_22222",
		"invoiceId":              "inv_33333",
		"refundId":               "ref_44444",
		"gatewayId":              int64(1),
		"quantity":               int64(1),
		"description":            "Monthly subscription",
		"reason":                 "Customer request",
		"periodStart":            "2025-01-01",
		"periodEnd":              "2025-02-01",
		"trialEnd":               "2025-01-15",
		"nextBillingDate":        "2025-02-01",
		"externalSubscriptionId": "ext_sub_55555",
	}

	vars := telegram.BuildVariableMap(req.Event, sampleData)
	rendered := telegram.RenderTemplate(req.Template, vars)
	usedVars := telegram.ExtractVariables(req.Template)

	return &_telegram.TemplatePreviewRes{
		RenderedMessage: rendered,
		UsedVariables:   usedVars,
	}, nil
}
