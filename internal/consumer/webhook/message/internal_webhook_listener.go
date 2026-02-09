package message

import (
	"context"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	redismq "github.com/jackyang-hk/go-redismq"

	redismq2 "unibee/internal/cmd/redismq"
	"unibee/internal/logic/telegram"
	"unibee/utility"
)

type InternalWebhookListener struct {
}

func (t InternalWebhookListener) GetTopic() string {
	return redismq2.TopicInternalWebhook.Topic
}

func (t InternalWebhookListener) GetTag() string {
	return redismq2.TopicInternalWebhook.Tag
}

func (t InternalWebhookListener) Consume(ctx context.Context, message *redismq.Message) redismq.Action {
	utility.Assert(len(message.Body) > 0, "body is nil")
	utility.Assert(len(message.Body) != 0, "body length is 0")

	// Parse webhook message and send Telegram notification
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(ctx, "telegram notification panic: %v", r)
			}
		}()
		t.sendTelegramNotification(ctx, message.Body)
	}()

	return redismq.CommitMessage
}

// sendTelegramNotification extracts event data and dispatches to Telegram notifier.
func (t InternalWebhookListener) sendTelegramNotification(ctx context.Context, body string) {
	j, err := gjson.DecodeToJson([]byte(body))
	if err != nil {
		g.Log().Debugf(ctx, "telegram: failed to parse webhook body: %v", err)
		return
	}

	merchantId := j.Get("MerchantId").Uint64()
	event := j.Get("Event").String()
	if merchantId == 0 || event == "" {
		return
	}

	// Extract nested Data json
	dataJson := make(map[string]interface{})
	dataRaw := j.Get("Data")
	if dataRaw != nil {
		if m := dataRaw.Map(); m != nil {
			dataJson = m
		}
	}

	if err := telegram.SendNotification(ctx, merchantId, event, dataJson); err != nil {
		g.Log().Warningf(ctx, "telegram notification failed for merchant %d event %s: %v", merchantId, event, err)
	}
}

func init() {
	listener := NewInternalWebhookListener()
	redismq.RegisterListener(listener)
}

func NewInternalWebhookListener() *InternalWebhookListener {
	return &InternalWebhookListener{}
}
