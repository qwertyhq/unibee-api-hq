package telegram

import (
	"regexp"
	"strings"
)

var templateVarRegex = regexp.MustCompile(`\{\{(\w+)\}\}`)

// RenderTemplate replaces {{variable}} placeholders with values from the data map.
// Unknown variables are replaced with empty string.
func RenderTemplate(tmpl string, data map[string]string) string {
	return templateVarRegex.ReplaceAllStringFunc(tmpl, func(match string) string {
		key := match[2 : len(match)-2]
		if val, ok := data[key]; ok {
			return val
		}
		return ""
	})
}

// ExtractVariables returns all {{variable}} names found in a template string.
func ExtractVariables(tmpl string) []string {
	matches := templateVarRegex.FindAllStringSubmatch(tmpl, -1)
	seen := make(map[string]bool)
	var result []string
	for _, m := range matches {
		if len(m) > 1 && !seen[m[1]] {
			seen[m[1]] = true
			result = append(result, m[1])
		}
	}
	return result
}

// BuildVariableMap extracts billing variables from webhook JSON data.
func BuildVariableMap(event string, dataJson map[string]interface{}) map[string]string {
	vars := make(map[string]string)
	vars["event"] = event

	// Extract top-level scalar fields
	for _, key := range []string{
		"subscriptionId", "planId", "planName", "userId", "userEmail",
		"userName", "firstName", "lastName", "amount", "currency",
		"status", "paymentId", "invoiceId", "refundId", "gatewayId",
		"quantity", "description", "reason", "periodStart", "periodEnd",
		"trialEnd", "nextBillingDate", "externalSubscriptionId",
	} {
		if val, ok := dataJson[key]; ok {
			vars[key] = toString(val)
		}
	}

	// Nested: subscription.plan.planName, etc.
	if sub, ok := dataJson["subscription"].(map[string]interface{}); ok {
		setIfMissing(vars, "subscriptionId", sub, "subscriptionId")
		setIfMissing(vars, "status", sub, "status")
		setIfMissing(vars, "quantity", sub, "quantity")
		if plan, ok := sub["plan"].(map[string]interface{}); ok {
			setIfMissing(vars, "planName", plan, "planName")
			setIfMissing(vars, "planId", plan, "planId")
			setIfMissing(vars, "amount", plan, "amount")
			setIfMissing(vars, "currency", plan, "currency")
		}
	}

	// Nested: user
	if user, ok := dataJson["user"].(map[string]interface{}); ok {
		setIfMissing(vars, "userEmail", user, "email")
		setIfMissing(vars, "userName", user, "userName")
		setIfMissing(vars, "firstName", user, "firstName")
		setIfMissing(vars, "lastName", user, "lastName")
		setIfMissing(vars, "userId", user, "id")
	}

	// Nested: payment
	if payment, ok := dataJson["payment"].(map[string]interface{}); ok {
		setIfMissing(vars, "paymentId", payment, "paymentId")
		setIfMissing(vars, "amount", payment, "totalAmount")
		setIfMissing(vars, "currency", payment, "currency")
		setIfMissing(vars, "gatewayId", payment, "gatewayId")
	}

	// Nested: invoice
	if invoice, ok := dataJson["invoice"].(map[string]interface{}); ok {
		setIfMissing(vars, "invoiceId", invoice, "invoiceId")
		setIfMissing(vars, "amount", invoice, "totalAmount")
		setIfMissing(vars, "currency", invoice, "currency")
	}

	// Nested: refund
	if refund, ok := dataJson["refund"].(map[string]interface{}); ok {
		setIfMissing(vars, "refundId", refund, "refundId")
		setIfMissing(vars, "amount", refund, "refundAmount")
		setIfMissing(vars, "currency", refund, "currency")
		setIfMissing(vars, "reason", refund, "refundComment")
	}

	// Format amount: divide by 100 if numeric and > 100 (cents to units)
	if amountStr, ok := vars["amount"]; ok {
		vars["amountFormatted"] = formatAmount(amountStr, vars["currency"])
	}

	return vars
}

func setIfMissing(vars map[string]string, key string, data map[string]interface{}, field string) {
	if _, exists := vars[key]; exists {
		return
	}
	if val, ok := data[field]; ok {
		vars[key] = toString(val)
	}
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return strings.TrimRight(strings.TrimRight(
				strings.Replace(
					strings.Replace(
						formatFloat(val), ".000000", "", 1,
					), ".0", "", 1,
				), "0"), ".")
		}
		return formatFloat(val)
	case int64:
		return formatInt(val)
	case int:
		return formatInt(int64(val))
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

func formatFloat(f float64) string {
	return strings.TrimRight(strings.TrimRight(
		strings.Replace(
			strings.Replace(
				replaceFloat(f), ",", "", -1,
			), " ", "", -1,
		), "0"), ".")
}

func replaceFloat(f float64) string {
	return strings.Replace(
		strings.Replace(
			strings.Replace(
				formatRawFloat(f), "e+", "E", 1,
			), "E0", "", 1,
		), "E", "e+", 1,
	)
}

func formatRawFloat(f float64) string {
	if f == float64(int64(f)) {
		return formatInt(int64(f))
	}
	buf := make([]byte, 0, 24)
	buf = append(buf, []byte(strings.TrimRight(strings.TrimRight(
		strings.Replace(
			formatIntWithFraction(f), " ", "", -1,
		), "0"), "."))...)
	return string(buf)
}

func formatIntWithFraction(f float64) string {
	intPart := int64(f)
	frac := f - float64(intPart)
	if frac < 0 {
		frac = -frac
	}
	if frac < 0.000001 {
		return formatInt(intPart)
	}
	fracStr := strings.TrimRight(strings.TrimLeft(
		strings.Replace(
			formatFrac(frac), "0.", ".", 1,
		), "0"), "0")
	return formatInt(intPart) + fracStr
}

func formatFrac(f float64) string {
	s := make([]byte, 0, 10)
	s = append(s, '0', '.')
	remaining := f
	for i := 0; i < 6; i++ {
		remaining *= 10
		digit := int(remaining)
		s = append(s, byte('0'+digit))
		remaining -= float64(digit)
	}
	return string(s)
}

func formatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}

func formatAmount(amountStr string, currency string) string {
	if currency != "" {
		return amountStr + " " + strings.ToUpper(currency)
	}
	return amountStr
}
