package webhook

import (
	"context"
	"fmt"
	"io"
	"strings"

	"unibee/internal/logic/gateway/gateway_bean"
	entity "unibee/internal/model/entity/default"
	"unibee/internal/query"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// PlategaWebhook implements GatewayWebhookInterface for Platega.io callbacks.
type PlategaWebhook struct{}

func (w PlategaWebhook) GatewayCheckAndSetupWebhook(ctx context.Context, gateway *entity.MerchantGateway) (err error) {
	// Platega webhook URL is configured in the merchant dashboard (Settings → Callback URLs).
	// No programmatic setup needed.
	return nil
}

// GatewayWebhook handles Platega callback notifications.
// Platega sends: X-MerchantId, X-Secret headers + JSON body with id, amount, currency, status, paymentMethod, payload.
func (w PlategaWebhook) GatewayWebhook(r *ghttp.Request, gateway *entity.MerchantGateway) {
	ctx := r.Context()

	// 1. Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		g.Log().Errorf(ctx, "Platega webhook: failed to read body: %v", err)
		r.Response.WriteJsonExit(g.Map{"error": "failed to read body"})
		return
	}
	if len(body) == 0 {
		g.Log().Errorf(ctx, "Platega webhook: empty body")
		r.Response.WriteJsonExit(g.Map{"error": "empty body"})
		return
	}

	g.Log().Infof(ctx, "Platega webhook raw body: %s", string(body))

	// 2. Verify credentials from headers
	headerMerchantID := r.Header.Get("X-MerchantId")
	headerSecret := r.Header.Get("X-Secret")

	if headerMerchantID != gateway.GatewayKey || headerSecret != gateway.GatewaySecret {
		g.Log().Errorf(ctx, "Platega webhook: invalid credentials, merchantId=%s", headerMerchantID)
		r.Response.WriteJsonExit(g.Map{"error": "unauthorized"})
		return
	}

	// 3. Parse callback payload
	data := gjson.New(body)
	transactionID := data.Get("id").String()
	status := data.Get("status").String()
	payload := data.Get("payload").String() // This is our PaymentId

	g.Log().Infof(ctx, "Platega webhook: transactionId=%s status=%s payload=%s", transactionID, status, payload)

	// 4. Route by status
	switch strings.ToUpper(status) {
	case "CONFIRMED":
		w.handlePaymentConfirmed(ctx, transactionID, payload, gateway)
	case "CANCELED":
		w.handlePaymentCanceled(ctx, transactionID, payload, gateway)
	case "CHARGEBACKED":
		w.handlePaymentChargebacked(ctx, transactionID, payload, gateway)
	default:
		g.Log().Infof(ctx, "Platega webhook: unhandled status: %s", status)
	}

	// 5. Return 200 OK
	r.Response.WriteJsonExit(g.Map{"status": "ok"})
}

func (w PlategaWebhook) GatewayRedirect(r *ghttp.Request, gateway *entity.MerchantGateway) (res *gateway_bean.GatewayRedirectResp, err error) {
	ctx := r.Context()

	// Platega redirects to the return/failedUrl set during payment creation.
	// We look up the payment by gateway payment ID from query params.
	transactionID := r.Get("transaction_id").String()
	if transactionID == "" {
		transactionID = r.Get("transactionId").String()
	}

	g.Log().Infof(ctx, "Platega redirect: transactionId=%s", transactionID)

	payment := query.GetPaymentByGatewayPaymentId(ctx, transactionID)
	if payment == nil {
		return nil, gerror.Newf("payment not found for transactionId: %s", transactionID)
	}

	return &gateway_bean.GatewayRedirectResp{
		Payment:   payment,
		Status:    true,
		Success:   true,
		Message:   "Payment redirect",
		ReturnUrl: payment.ReturnUrl,
	}, nil
}

func (w PlategaWebhook) GatewayNewPaymentMethodRedirect(r *ghttp.Request, gateway *entity.MerchantGateway) (err error) {
	return gerror.New("Platega does not support payment method management")
}

// handlePaymentConfirmed processes a successful payment callback.
func (w PlategaWebhook) handlePaymentConfirmed(ctx context.Context, transactionID, paymentID string, gateway *entity.MerchantGateway) {
	if paymentID == "" {
		// Fallback: look up by gateway payment ID
		payment := query.GetPaymentByGatewayPaymentId(ctx, transactionID)
		if payment != nil {
			paymentID = payment.PaymentId
		}
	}

	if paymentID == "" {
		g.Log().Errorf(ctx, "Platega webhook CONFIRMED: cannot resolve paymentId for transactionId=%s", transactionID)
		return
	}

	err := ProcessPaymentWebhook(ctx, paymentID, transactionID, gateway)
	if err != nil {
		g.Log().Errorf(ctx, "Platega webhook CONFIRMED: processing failed: %v", err)
	}
}

// handlePaymentCanceled processes a canceled payment callback.
func (w PlategaWebhook) handlePaymentCanceled(ctx context.Context, transactionID, paymentID string, gateway *entity.MerchantGateway) {
	if paymentID == "" {
		payment := query.GetPaymentByGatewayPaymentId(ctx, transactionID)
		if payment != nil {
			paymentID = payment.PaymentId
		}
	}

	if paymentID == "" {
		g.Log().Errorf(ctx, "Platega webhook CANCELED: cannot resolve paymentId for transactionId=%s", transactionID)
		return
	}

	err := ProcessPaymentWebhook(ctx, paymentID, transactionID, gateway)
	if err != nil {
		g.Log().Errorf(ctx, "Platega webhook CANCELED: processing failed: %v", err)
	}
}

// handlePaymentChargebacked processes a chargeback callback.
func (w PlategaWebhook) handlePaymentChargebacked(ctx context.Context, transactionID, paymentID string, gateway *entity.MerchantGateway) {
	g.Log().Warningf(ctx, "Platega webhook CHARGEBACK: transactionId=%s paymentId=%s — manual intervention required", transactionID, paymentID)

	if paymentID == "" {
		payment := query.GetPaymentByGatewayPaymentId(ctx, transactionID)
		if payment != nil {
			paymentID = payment.PaymentId
		}
	}

	if paymentID == "" {
		g.Log().Errorf(ctx, "Platega webhook CHARGEBACK: cannot resolve paymentId for transactionId=%s", transactionID)
		return
	}

	// Process as regular webhook — GatewayPaymentDetail will fetch the current status from Platega
	err := ProcessPaymentWebhook(ctx, paymentID, transactionID, gateway)
	if err != nil {
		g.Log().Errorf(ctx, "Platega webhook CHARGEBACK: processing failed: %v", err)
	}

	_ = fmt.Sprintf("chargeback processed for %s", transactionID)
}
