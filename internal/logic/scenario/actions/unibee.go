package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"unibee/internal/consts"
	"unibee/internal/logic/discount"
	"unibee/internal/logic/scenario"
	subdetail "unibee/internal/logic/subscription/service/detail"
	"unibee/internal/query"
	"unibee/utility"
)

func init() {
	scenario.RegisterAction(scenario.StepUniBeeAPI, &UniBeeAPIAction{})
}

// UniBeeAPIAction calls internal UniBee billing API.
type UniBeeAPIAction struct{}

func (a *UniBeeAPIAction) Execute(ctx context.Context, execCtx *scenario.ExecutionContext, step *scenario.StepDSL) (map[string]interface{}, error) {
	action, _ := step.Params["action"].(string)
	if action == "" {
		return nil, fmt.Errorf("unibee_api: action is required")
	}

	params, _ := step.Params["params"].(map[string]interface{})
	if params == nil {
		params = make(map[string]interface{})
	}

	// Render template variables in params
	rendered := renderParamVars(params, execCtx.Variables)

	switch action {
	case "get_subscription":
		return a.getSubscription(ctx, execCtx, rendered)
	case "get_user":
		return a.getUser(ctx, execCtx, rendered)
	case "get_invoice_list":
		return a.getInvoiceList(ctx, execCtx, rendered)
	case "cancel_subscription":
		return a.cancelSubscription(ctx, execCtx, rendered)
	case "create_discount":
		return a.createDiscount(ctx, execCtx, rendered)
	case "get_plan":
		return a.getPlan(ctx, execCtx, rendered)
	default:
		return nil, fmt.Errorf("unibee_api: unknown action %q", action)
	}
}

// getSubscription returns subscription details by subscriptionId or userId.
func (a *UniBeeAPIAction) getSubscription(ctx context.Context, execCtx *scenario.ExecutionContext, params map[string]interface{}) (map[string]interface{}, error) {
	// Try subscriptionId first
	if subId := getStringParam(params, "subscriptionId", execCtx.Variables); subId != "" {
		detail, err := subdetail.SubscriptionDetail(ctx, subId)
		if err != nil {
			return nil, fmt.Errorf("unibee_api: get_subscription failed: %w", err)
		}
		if detail == nil || detail.Subscription == nil {
			return map[string]interface{}{"found": "false"}, nil
		}
		return subscriptionToVars(detail), nil
	}

	// Try userId
	userId := getUint64Param(params, "userId", execCtx.Variables)
	if userId == 0 {
		return nil, fmt.Errorf("unibee_api: get_subscription requires subscriptionId or userId")
	}

	subs := query.GetUserAllActiveOrIncompleteSubscriptions(ctx, userId, execCtx.MerchantID)
	if len(subs) == 0 {
		return map[string]interface{}{"found": "false"}, nil
	}

	// Get detail for the first active subscription
	detail, err := subdetail.SubscriptionDetail(ctx, subs[0].SubscriptionId)
	if err != nil {
		return nil, fmt.Errorf("unibee_api: get_subscription detail failed: %w", err)
	}
	return subscriptionToVars(detail), nil
}

// getUser returns user details.
func (a *UniBeeAPIAction) getUser(ctx context.Context, execCtx *scenario.ExecutionContext, params map[string]interface{}) (map[string]interface{}, error) {
	userId := getUint64Param(params, "userId", execCtx.Variables)
	if userId == 0 {
		// Try to get from email
		email := getStringParam(params, "email", execCtx.Variables)
		if email != "" {
			user := query.GetUserAccountByEmail(ctx, execCtx.MerchantID, email)
			if user == nil {
				return map[string]interface{}{"found": "false"}, nil
			}
			return userToVars(user), nil
		}
		return nil, fmt.Errorf("unibee_api: get_user requires userId or email")
	}

	user := query.GetUserAccountById(ctx, userId)
	if user == nil {
		return map[string]interface{}{"found": "false"}, nil
	}
	return userToVars(user), nil
}

// getInvoiceList returns recent invoices for a user.
func (a *UniBeeAPIAction) getInvoiceList(ctx context.Context, execCtx *scenario.ExecutionContext, params map[string]interface{}) (map[string]interface{}, error) {
	userId := getUint64Param(params, "userId", execCtx.Variables)
	if userId == 0 {
		return nil, fmt.Errorf("unibee_api: get_invoice_list requires userId")
	}

	limit := getIntParam(params, "limit")
	if limit <= 0 || limit > 20 {
		limit = 5
	}

	var list []map[string]interface{}
	// Query invoices via raw DB query for simplicity and to avoid circular import issues
	type invoiceRow struct {
		InvoiceId      string `json:"invoiceId"`
		TotalAmount    int64  `json:"totalAmount"`
		Currency       string `json:"currency"`
		Status         int    `json:"status"`
		SubscriptionId string `json:"subscriptionId"`
		PeriodStart    int64  `json:"periodStart"`
		PeriodEnd      int64  `json:"periodEnd"`
	}

	var rows []*invoiceRow
	err := g.DB().Ctx(ctx).Raw(
		`SELECT invoice_id, total_amount, currency, status, subscription_id, period_start, period_end 
		 FROM invoice 
		 WHERE user_id = ? AND merchant_id = ? AND is_deleted = 0 
		 ORDER BY gmt_create DESC LIMIT ?`,
		userId, execCtx.MerchantID, limit,
	).Scan(&rows)
	if err != nil {
		return nil, fmt.Errorf("unibee_api: get_invoice_list failed: %w", err)
	}

	for _, row := range rows {
		list = append(list, map[string]interface{}{
			"invoiceId":      row.InvoiceId,
			"amount":         utility.ConvertCentToDollarStr(row.TotalAmount, row.Currency),
			"currency":       row.Currency,
			"status":         invoiceStatusText(row.Status),
			"subscriptionId": row.SubscriptionId,
		})
	}

	result := map[string]interface{}{
		"found":         "true",
		"invoice_count": strconv.Itoa(len(rows)),
	}

	// Serialize summary into variable-friendly format
	if len(list) > 0 {
		data, _ := json.Marshal(list)
		result["invoices_json"] = string(data)

		// Build a human-readable text list
		var lines []string
		for i, inv := range list {
			lines = append(lines, fmt.Sprintf("%d. %s %s — %s (%s)",
				i+1,
				inv["amount"], inv["currency"],
				inv["status"], inv["invoiceId"]))
		}
		result["invoices_text"] = strings.Join(lines, "\n")
	}

	return result, nil
}

// cancelSubscription cancels a subscription.
func (a *UniBeeAPIAction) cancelSubscription(ctx context.Context, execCtx *scenario.ExecutionContext, params map[string]interface{}) (map[string]interface{}, error) {
	subId := getStringParam(params, "subscriptionId", execCtx.Variables)
	if subId == "" {
		return nil, fmt.Errorf("unibee_api: cancel_subscription requires subscriptionId")
	}

	sub := query.GetSubscriptionBySubscriptionId(ctx, subId)
	if sub == nil {
		return nil, fmt.Errorf("unibee_api: subscription %s not found", subId)
	}

	// Use cancel at period end by default (safer)
	// Direct import would create circular deps, so use raw DB update
	_, err := g.DB().Ctx(ctx).Exec(ctx,
		`UPDATE subscription SET cancel_at_period_end = 1, gmt_modify = NOW() WHERE subscription_id = ? AND merchant_id = ?`,
		subId, execCtx.MerchantID,
	)
	if err != nil {
		return nil, fmt.Errorf("unibee_api: cancel_subscription failed: %w", err)
	}

	g.Log().Infof(ctx, "scenario: cancel_subscription %s for merchant %d", subId, execCtx.MerchantID)
	return map[string]interface{}{"cancelled": "true", "subscriptionId": subId}, nil
}

// createDiscount creates a discount code.
func (a *UniBeeAPIAction) createDiscount(ctx context.Context, execCtx *scenario.ExecutionContext, params map[string]interface{}) (map[string]interface{}, error) {
	code := getStringParam(params, "code", execCtx.Variables)
	if code == "" {
		return nil, fmt.Errorf("unibee_api: create_discount requires code")
	}

	// Check if code already exists
	existing := query.GetDiscountByCode(ctx, execCtx.MerchantID, code)
	if existing != nil {
		return map[string]interface{}{
			"discount_id":    strconv.FormatUint(existing.Id, 10),
			"code":           existing.Code,
			"already_exists": "true",
		}, nil
	}

	discountType := getIntParam(params, "discountType")
	if discountType == 0 {
		discountType = 1 // default: percentage
	}

	billingType := getIntParam(params, "billingType")
	if billingType == 0 {
		billingType = 2 // default: recurring
	}

	req := &discount.CreateDiscountCodeInternalReq{
		MerchantId:         execCtx.MerchantID,
		Code:               code,
		DiscountType:       discountType,
		BillingType:        billingType,
		DiscountPercentage: int64(getIntParam(params, "discountPercentage")),
		DiscountAmount:     int64(getIntParam(params, "discountAmount")),
		Currency:           getStringParam(params, "currency", execCtx.Variables),
	}

	// Set name if provided
	if name := getStringParam(params, "name", execCtx.Variables); name != "" {
		req.Name = &name
	}

	dc, err := discount.NewMerchantDiscountCode(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("unibee_api: create_discount failed: %w", err)
	}

	// Auto-activate
	_ = discount.ActivateMerchantDiscountCode(ctx, execCtx.MerchantID, dc.Id)

	return map[string]interface{}{
		"discount_id": strconv.FormatUint(dc.Id, 10),
		"code":        dc.Code,
		"created":     "true",
	}, nil
}

// getPlan returns plan details.
func (a *UniBeeAPIAction) getPlan(ctx context.Context, execCtx *scenario.ExecutionContext, params map[string]interface{}) (map[string]interface{}, error) {
	planId := getUint64Param(params, "planId", execCtx.Variables)
	if planId == 0 {
		return nil, fmt.Errorf("unibee_api: get_plan requires planId")
	}

	plan := query.GetPlanById(ctx, planId)
	if plan == nil {
		return map[string]interface{}{"found": "false"}, nil
	}

	return map[string]interface{}{
		"found":       "true",
		"plan_id":     strconv.FormatUint(plan.Id, 10),
		"plan_name":   plan.PlanName,
		"amount":      utility.ConvertCentToDollarStr(plan.Amount, plan.Currency),
		"currency":    plan.Currency,
		"interval":    fmt.Sprintf("%d %s", plan.IntervalCount, plan.IntervalUnit),
		"description": plan.Description,
		"plan_status": planStatusText(plan.Status),
	}, nil
}

// ──── helpers ────

func subscriptionToVars(d interface{}) map[string]interface{} {
	// Type switch to handle the detail struct
	type subDetail interface {
		GetSubscription() interface{}
	}
	// Use JSON round-trip for simplicity and type safety
	data, _ := json.Marshal(d)
	var raw map[string]interface{}
	_ = json.Unmarshal(data, &raw)

	result := map[string]interface{}{"found": "true"}

	// Extract key fields from the subscription detail
	if sub, ok := raw["subscription"].(map[string]interface{}); ok {
		if v, ok := sub["subscriptionId"].(string); ok {
			result["subscription_id"] = v
		}
		if v, ok := sub["status"].(float64); ok {
			result["subscription_status"] = subscriptionStatusText(int(v))
		}
		if v, ok := sub["amount"].(float64); ok {
			currency := ""
			if c, ok := sub["currency"].(string); ok {
				currency = c
			}
			result["subscription_amount"] = utility.ConvertCentToDollarStr(int64(v), currency)
			result["subscription_currency"] = currency
		}
		if v, ok := sub["currentPeriodEnd"].(float64); ok {
			result["period_end"] = strconv.FormatInt(int64(v), 10)
		}
	}
	if plan, ok := raw["plan"].(map[string]interface{}); ok {
		if v, ok := plan["planName"].(string); ok {
			result["plan_name"] = v
		}
		if v, ok := plan["intervalUnit"].(string); ok {
			result["interval_unit"] = v
		}
	}
	if user, ok := raw["user"].(map[string]interface{}); ok {
		if v, ok := user["email"].(string); ok {
			result["user_email"] = v
		}
	}

	return result
}

func userToVars(user interface{}) map[string]interface{} {
	data, _ := json.Marshal(user)
	var raw map[string]interface{}
	_ = json.Unmarshal(data, &raw)

	result := map[string]interface{}{"found": "true"}
	for _, key := range []string{"email", "userName", "firstName", "lastName", "language", "status", "subscriptionId"} {
		if v, ok := raw[key]; ok && v != nil {
			result["user_"+key] = fmt.Sprintf("%v", v)
		}
	}
	if v, ok := raw["id"].(float64); ok {
		result["user_id"] = strconv.FormatUint(uint64(v), 10)
	}

	return result
}

func renderParamVars(params map[string]interface{}, vars map[string]string) map[string]interface{} {
	rendered := make(map[string]interface{}, len(params))
	for k, v := range params {
		if s, ok := v.(string); ok {
			rendered[k] = scenario.RenderVars(s, vars)
		} else {
			rendered[k] = v
		}
	}
	return rendered
}

func getStringParam(params map[string]interface{}, key string, vars map[string]string) string {
	if v, ok := params[key].(string); ok {
		return scenario.RenderVars(v, vars)
	}
	// Fallback: check variables
	if v, ok := vars[key]; ok {
		return v
	}
	return ""
}

func getUint64Param(params map[string]interface{}, key string, vars map[string]string) uint64 {
	raw := getStringParam(params, key, vars)
	if raw == "" {
		// Try float (JSON numbers)
		if v, ok := params[key].(float64); ok {
			return uint64(v)
		}
		return 0
	}
	val, _ := strconv.ParseUint(raw, 10, 64)
	return val
}

func getIntParam(params map[string]interface{}, key string) int {
	if v, ok := params[key].(float64); ok {
		return int(v)
	}
	if v, ok := params[key].(string); ok {
		i, _ := strconv.Atoi(v)
		return i
	}
	return 0
}

func subscriptionStatusText(status int) string {
	switch status {
	case consts.SubStatusActive:
		return "active"
	case consts.SubStatusCancelled:
		return "cancelled"
	case consts.SubStatusExpired:
		return "expired"
	case consts.SubStatusSuspended:
		return "suspended"
	case consts.SubStatusIncomplete:
		return "incomplete"
	default:
		return fmt.Sprintf("status_%d", status)
	}
}

func invoiceStatusText(status int) string {
	switch status {
	case 1:
		return "pending"
	case 2:
		return "processing"
	case 3:
		return "paid"
	case 4:
		return "failed"
	case 5:
		return "cancelled"
	case 6:
		return "reversed"
	default:
		return fmt.Sprintf("status_%d", status)
	}
}

func planStatusText(status int) string {
	switch status {
	case 1:
		return "editing"
	case 2:
		return "active"
	case 3:
		return "inactive"
	case 4:
		return "expired"
	default:
		return fmt.Sprintf("status_%d", status)
	}
}
