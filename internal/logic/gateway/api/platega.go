package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"unibee/internal/consts"
	_interface "unibee/internal/interface"
	"unibee/internal/logic/gateway/api/log"
	"unibee/internal/logic/gateway/gateway_bean"
	entity "unibee/internal/model/entity/default"
	"unibee/utility"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
)

const plategaBaseURL = "https://app.platega.io"

// Platega implements GatewayInterface for the Platega.io payment gateway.
type Platega struct{}

// Platega API request/response structures

type PlategaCreateTransactionRequest struct {
	PaymentMethod  int                   `json:"paymentMethod"`
	PaymentDetails PlategaPaymentDetails `json:"paymentDetails"`
	Description    string                `json:"description"`
	Return         string                `json:"return"`
	FailedURL      string                `json:"failedUrl"`
	Payload        string                `json:"payload,omitempty"`
}

type PlategaPaymentDetails struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type PlategaCreateTransactionResponse struct {
	PaymentMethod  string  `json:"paymentMethod"`
	TransactionID  string  `json:"transactionId"`
	Redirect       string  `json:"redirect"`
	Return         string  `json:"return"`
	PaymentDetails any     `json:"paymentDetails"`
	Status         string  `json:"status"`
	ExpiresIn      string  `json:"expiresIn"`
	MerchantID     string  `json:"merchantId"`
	USDTRate       float64 `json:"usdtRate"`
}

type PlategaTransactionStatusResponse struct {
	ID                string                 `json:"id"`
	Status            string                 `json:"status"`
	PaymentDetails    *PlategaPaymentDetails `json:"paymentDetails"`
	MerchantName      string                 `json:"merchantName"`
	MerchantID        string                 `json:"mechantId"`
	Commission        float64                `json:"comission"`
	PaymentMethod     string                 `json:"paymentMethod"`
	ExpiresIn         string                 `json:"expiresIn"`
	Return            string                 `json:"return"`
	CommissionUSDT    float64                `json:"comissionUsdt"`
	AmountUSDT        float64                `json:"amountUsdt"`
	QR                string                 `json:"qr"`
	PayformSuccessURL string                 `json:"payformSuccessUrl"`
	Payload           string                 `json:"payload"`
	CommissionType    int                    `json:"comissionType"`
	ExternalID        string                 `json:"externalId"`
	Description       string                 `json:"description"`
}

func (p Platega) GatewayInfo(ctx context.Context) *_interface.GatewayInfo {
	return &_interface.GatewayInfo{
		Name:                          "Platega",
		Description:                   "Platega payment gateway — SBP, cards, acquiring, international, crypto",
		DisplayName:                   "Platega",
		GatewayWebsiteLink:            "https://platega.io/",
		GatewayWebhookIntegrationLink: "https://docs.platega.io/",
		GatewayLogo:                   "",
		GatewayIcons:                  []string{},
		GatewayType:                   consts.GatewayTypeCard,
		Sort:                          230,
		AutoChargeEnabled:             false,
		PublicKeyName:                 "MerchantId",
		PrivateSecretName:             "Secret",
		Host:                          plategaBaseURL,
	}
}

func (p Platega) GatewayTest(ctx context.Context, req *_interface.GatewayTestReq) (icon string, gatewayType int64, err error) {
	utility.Assert(len(req.Key) > 0, "Platega MerchantId is required")
	utility.Assert(len(req.Secret) > 0, "Platega Secret is required")
	return "", consts.GatewayTypeCard, nil
}

func (p Platega) GatewayUserCreate(ctx context.Context, gateway *entity.MerchantGateway, user *entity.UserAccount) (res *gateway_bean.GatewayUserCreateResp, err error) {
	return &gateway_bean.GatewayUserCreateResp{
		GatewayUserId: fmt.Sprintf("platega_user_%d", user.Id),
	}, nil
}

func (p Platega) GatewayUserDetailQuery(ctx context.Context, gateway *entity.MerchantGateway, gatewayUserId string) (res *gateway_bean.GatewayUserDetailQueryResp, err error) {
	return &gateway_bean.GatewayUserDetailQueryResp{
		GatewayUserId: gatewayUserId,
	}, nil
}

func (p Platega) GatewayMerchantBalancesQuery(ctx context.Context, gateway *entity.MerchantGateway) (res *gateway_bean.GatewayMerchantBalanceQueryResp, err error) {
	balances := []*gateway_bean.GatewayBalance{
		{Amount: 0, Currency: "RUB"},
	}
	return &gateway_bean.GatewayMerchantBalanceQueryResp{
		AvailableBalance: balances,
		PendingBalance:   balances,
	}, nil
}

func (p Platega) GatewayNewPayment(ctx context.Context, gateway *entity.MerchantGateway, createPayContext *gateway_bean.GatewayNewPaymentReq) (res *gateway_bean.GatewayNewPaymentResp, err error) {
	_, description := createPayContext.GetInvoiceSingleProductNameAndDescription()
	if description == "" {
		description = "Payment " + createPayContext.Pay.PaymentId
	}

	returnURL := createPayContext.Pay.ReturnUrl
	if returnURL == "" {
		returnURL = "https://example.com/success"
	}
	failedURL := returnURL

	// Amount in Platega is float (e.g. 500.00 RUB). UniBee stores amounts in cents.
	amountFloat := float64(createPayContext.Pay.PaymentAmount) / 100.0

	reqBody := PlategaCreateTransactionRequest{
		PaymentMethod: 10, // Default: Cards RUB. Override via gateway config if needed.
		PaymentDetails: PlategaPaymentDetails{
			Amount:   amountFloat,
			Currency: strings.ToUpper(createPayContext.Pay.Currency),
		},
		Description: description,
		Return:      returnURL,
		FailedURL:   failedURL,
		Payload:     createPayContext.Pay.PaymentId,
	}

	respBody, err := p.doRequest(ctx, gateway, "POST", "/transaction/process", reqBody)
	if err != nil {
		return nil, gerror.Newf("Platega create payment failed: %v", err)
	}

	var platResp PlategaCreateTransactionResponse
	if err = json.Unmarshal(respBody, &platResp); err != nil {
		return nil, gerror.Newf("Platega parse response failed: %v", err)
	}

	log.SaveChannelHttpLog("GatewayNewPayment", reqBody, string(respBody), nil,
		fmt.Sprintf("%s-%d", gateway.GatewayName, gateway.Id), nil, gateway)

	return &gateway_bean.GatewayNewPaymentResp{
		Payment:          createPayContext.Pay,
		Status:           p.mapStatus(platResp.Status),
		GatewayPaymentId: platResp.TransactionID,
		Link:             platResp.Redirect,
	}, nil
}

func (p Platega) GatewayPaymentDetail(ctx context.Context, gateway *entity.MerchantGateway, gatewayPaymentId string, payment *entity.Payment) (res *gateway_bean.GatewayPaymentRo, err error) {
	respBody, err := p.doRequest(ctx, gateway, "GET", fmt.Sprintf("/transaction/%s", gatewayPaymentId), nil)
	if err != nil {
		return nil, gerror.Newf("Platega get payment detail failed: %v", err)
	}

	var platResp PlategaTransactionStatusResponse
	if err = json.Unmarshal(respBody, &platResp); err != nil {
		return nil, gerror.Newf("Platega parse response failed: %v", err)
	}

	log.SaveChannelHttpLog("GatewayPaymentDetail",
		map[string]interface{}{"transaction_id": gatewayPaymentId}, string(respBody), nil,
		fmt.Sprintf("%s-%d", gateway.GatewayName, gateway.Id), nil, gateway)

	var amount int64
	var currency string
	if platResp.PaymentDetails != nil {
		amount = int64(platResp.PaymentDetails.Amount * 100) // convert to cents
		currency = platResp.PaymentDetails.Currency
	}

	return &gateway_bean.GatewayPaymentRo{
		GatewayPaymentId: gatewayPaymentId,
		Status:           int(p.mapStatus(platResp.Status)),
		PaymentAmount:    amount,
		Currency:         currency,
		CreateTime:       gtime.Now(),
	}, nil
}

func (p Platega) GatewayPaymentList(ctx context.Context, gateway *entity.MerchantGateway, listReq *gateway_bean.GatewayPaymentListReq) (res []*gateway_bean.GatewayPaymentRo, err error) {
	return []*gateway_bean.GatewayPaymentRo{}, nil
}

func (p Platega) GatewayCapture(ctx context.Context, gateway *entity.MerchantGateway, payment *entity.Payment) (res *gateway_bean.GatewayPaymentCaptureResp, err error) {
	return &gateway_bean.GatewayPaymentCaptureResp{}, nil
}

func (p Platega) GatewayCancel(ctx context.Context, gateway *entity.MerchantGateway, payment *entity.Payment) (res *gateway_bean.GatewayPaymentCancelResp, err error) {
	return &gateway_bean.GatewayPaymentCancelResp{
		Status: consts.PaymentCancelled,
	}, nil
}

func (p Platega) GatewayRefund(ctx context.Context, gateway *entity.MerchantGateway, req *gateway_bean.GatewayNewPaymentRefundReq) (res *gateway_bean.GatewayPaymentRefundResp, err error) {
	// Platega does not expose a refund API — refunds are handled via the merchant dashboard
	return &gateway_bean.GatewayPaymentRefundResp{
		GatewayRefundId: req.Payment.GatewayPaymentId,
		Status:          consts.RefundCreated,
		RefundAmount:    req.Refund.RefundAmount,
		Currency:        req.Refund.Currency,
		Type:            consts.RefundTypeGateway,
		Reason:          "Platega refunds are processed via merchant dashboard",
	}, nil
}

func (p Platega) GatewayRefundDetail(ctx context.Context, gateway *entity.MerchantGateway, gatewayRefundId string, refund *entity.Refund) (res *gateway_bean.GatewayPaymentRefundResp, err error) {
	return &gateway_bean.GatewayPaymentRefundResp{
		GatewayRefundId: gatewayRefundId,
		Status:          consts.RefundCreated,
		Type:            consts.RefundTypeGateway,
	}, nil
}

func (p Platega) GatewayRefundList(ctx context.Context, gateway *entity.MerchantGateway, gatewayPaymentId string) (res []*gateway_bean.GatewayPaymentRefundResp, err error) {
	return []*gateway_bean.GatewayPaymentRefundResp{}, nil
}

func (p Platega) GatewayRefundCancel(ctx context.Context, gateway *entity.MerchantGateway, payment *entity.Payment, refund *entity.Refund) (res *gateway_bean.GatewayPaymentRefundResp, err error) {
	return &gateway_bean.GatewayPaymentRefundResp{
		GatewayRefundId: refund.GatewayRefundId,
		Status:          consts.RefundCancelled,
		RefundAmount:    refund.RefundAmount,
		Currency:        refund.Currency,
		Type:            consts.RefundTypeGateway,
	}, nil
}

func (p Platega) GatewayCryptoFiatTrans(ctx context.Context, from *gateway_bean.GatewayCryptoFromCurrencyAmountDetailReq) (to *gateway_bean.GatewayCryptoToCurrencyAmountDetailRes, err error) {
	return nil, fmt.Errorf("Platega does not support crypto fiat conversion")
}

func (p Platega) GatewayUserAttachPaymentMethodQuery(ctx context.Context, gateway *entity.MerchantGateway, userId uint64, gatewayPaymentMethod string) (res *gateway_bean.GatewayUserAttachPaymentMethodResp, err error) {
	return nil, fmt.Errorf("Platega does not support payment method management")
}

func (p Platega) GatewayUserDeAttachPaymentMethodQuery(ctx context.Context, gateway *entity.MerchantGateway, userId uint64, gatewayPaymentMethod string) (res *gateway_bean.GatewayUserDeAttachPaymentMethodResp, err error) {
	return nil, fmt.Errorf("Platega does not support payment method management")
}

func (p Platega) GatewayUserPaymentMethodListQuery(ctx context.Context, gateway *entity.MerchantGateway, req *gateway_bean.GatewayUserPaymentMethodReq) (res *gateway_bean.GatewayUserPaymentMethodListResp, err error) {
	return nil, fmt.Errorf("Platega does not support payment method management")
}

func (p Platega) GatewayUserCreateAndBindPaymentMethod(ctx context.Context, gateway *entity.MerchantGateway, userId uint64, currency string, metadata map[string]interface{}) (res *gateway_bean.GatewayUserPaymentMethodCreateAndBindResp, err error) {
	return nil, fmt.Errorf("Platega does not support payment method management")
}

// doRequest sends an authenticated HTTP request to the Platega API.
func (p Platega) doRequest(ctx context.Context, gateway *entity.MerchantGateway, method, path string, payload interface{}) ([]byte, error) {
	url := plategaBaseURL + path
	var req *http.Request
	var err error

	if method == "GET" || payload == nil {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	} else {
		jsonData, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			return nil, marshalErr
		}
		req, err = http.NewRequestWithContext(ctx, method, url, strings.NewReader(string(jsonData)))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
		}
	}
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-MerchantId", gateway.GatewayKey)
	req.Header.Set("X-Secret", gateway.GatewaySecret)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, gerror.Newf("failed to read Platega response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, gerror.Newf("Platega API error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// mapStatus converts Platega transaction status to UniBee PaymentStatusEnum.
func (p Platega) mapStatus(status string) consts.PaymentStatusEnum {
	switch strings.ToUpper(status) {
	case "PENDING":
		return consts.PaymentCreated
	case "CONFIRMED":
		return consts.PaymentSuccess
	case "CANCELED":
		return consts.PaymentFailed
	case "CHARGEBACKED":
		return consts.PaymentCancelled
	default:
		return consts.PaymentCreated
	}
}
