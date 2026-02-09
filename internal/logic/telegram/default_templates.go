package telegram

// DefaultTemplates provides default message templates per webhook event.
// Merchants can customize these via the API.
var DefaultTemplates = map[string]string{
	// Subscription lifecycle
	"subscription.created":   "🆕 New subscription created\nPlan: {{planName}}\nUser: {{userEmail}}\nAmount: {{amountFormatted}}",
	"subscription.activated": "✅ Subscription activated\nPlan: {{planName}}\nUser: {{userEmail}}",
	"subscription.updated":   "📝 Subscription updated\nPlan: {{planName}}\nUser: {{userEmail}}",
	"subscription.cancelled": "❌ Subscription cancelled\nPlan: {{planName}}\nUser: {{userEmail}}",
	"subscription.expired":   "⏰ Subscription expired\nPlan: {{planName}}\nUser: {{userEmail}}",
	"subscription.failed":    "🚫 Subscription failed\nPlan: {{planName}}\nUser: {{userEmail}}",

	// Auto-renewal
	"subscription.auto_renew.success": "🔄 Auto-renewal successful\nPlan: {{planName}}\nUser: {{userEmail}}\nAmount: {{amountFormatted}}",
	"subscription.auto_renew.failure": "⚠️ Auto-renewal failed\nPlan: {{planName}}\nUser: {{userEmail}}",

	// Payments
	"payment.created":   "💳 Payment created\nAmount: {{amountFormatted}}\nUser: {{userEmail}}",
	"payment.success":   "✅ Payment successful\nAmount: {{amountFormatted}}\nUser: {{userEmail}}",
	"payment.failure":   "❌ Payment failed\nAmount: {{amountFormatted}}\nUser: {{userEmail}}",
	"payment.cancelled": "🚫 Payment cancelled\nAmount: {{amountFormatted}}\nUser: {{userEmail}}",

	// Invoices
	"invoice.created":   "📄 Invoice created\nAmount: {{amountFormatted}}\nUser: {{userEmail}}",
	"invoice.paid":      "✅ Invoice paid\nAmount: {{amountFormatted}}\nUser: {{userEmail}}",
	"invoice.cancelled": "❌ Invoice cancelled\nUser: {{userEmail}}",
	"invoice.failed":    "🚫 Invoice failed\nUser: {{userEmail}}",

	// Refunds
	"refund.created":   "💰 Refund initiated\nAmount: {{amountFormatted}}\nUser: {{userEmail}}",
	"refund.success":   "✅ Refund completed\nAmount: {{amountFormatted}}\nUser: {{userEmail}}",
	"refund.failure":   "❌ Refund failed\nAmount: {{amountFormatted}}\nUser: {{userEmail}}",
	"refund.cancelled": "🚫 Refund cancelled\nUser: {{userEmail}}",
	"refund.reversed":  "🔙 Refund reversed\nAmount: {{amountFormatted}}\nUser: {{userEmail}}",

	// Users
	"user.created": "👤 New user registered\nEmail: {{userEmail}}\nName: {{firstName}} {{lastName}}",
	"user.updated": "📝 User updated\nEmail: {{userEmail}}",

	// One-time addons
	"subscription.onetime_addon.created":   "➕ One-time addon created\nUser: {{userEmail}}\nAmount: {{amountFormatted}}",
	"subscription.onetime_addon.success":   "✅ One-time addon paid\nUser: {{userEmail}}\nAmount: {{amountFormatted}}",
	"subscription.onetime_addon.cancelled": "❌ One-time addon cancelled\nUser: {{userEmail}}",
	"subscription.onetime_addon.expired":   "⏰ One-time addon expired\nUser: {{userEmail}}",
}

// AvailableVariables lists all template variables that can be used.
var AvailableVariables = []string{
	"event",
	"subscriptionId",
	"planId",
	"planName",
	"userId",
	"userEmail",
	"userName",
	"firstName",
	"lastName",
	"amount",
	"amountFormatted",
	"currency",
	"status",
	"paymentId",
	"invoiceId",
	"refundId",
	"gatewayId",
	"quantity",
	"description",
	"reason",
	"periodStart",
	"periodEnd",
	"trialEnd",
	"nextBillingDate",
	"externalSubscriptionId",
}
