package merchant

import (
	"context"
	"fmt"

	_scenario "unibee/api/merchant/scenario"
	_interface "unibee/internal/interface/context"
	"unibee/internal/logic/scenario"
)

func (c *ControllerScenario) Detail(ctx context.Context, req *_scenario.DetailReq) (res *_scenario.DetailRes, err error) {
	merchantId := _interface.GetMerchantId(ctx)
	s, err := scenario.GetScenario(ctx, merchantId, req.ScenarioId)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, fmt.Errorf("scenario not found")
	}

	return &_scenario.DetailRes{
		Scenario: &_scenario.ScenarioItem{
			Id:           s.Id,
			Name:         s.Name,
			Description:  s.Description,
			ScenarioJson: s.ScenarioJson,
			Enabled:      s.Enabled,
			TriggerType:  s.TriggerType,
			TriggerValue: s.TriggerValue,
			CreateTime:   s.CreateTime,
		},
	}, nil
}
