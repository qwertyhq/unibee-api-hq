package merchant

import (
	"context"

	_scenario "unibee/api/merchant/scenario"
	_interface "unibee/internal/interface/context"
	"unibee/internal/logic/scenario"
)

func (c *ControllerScenario) List(ctx context.Context, req *_scenario.ListReq) (res *_scenario.ListRes, err error) {
	merchantId := _interface.GetMerchantId(ctx)
	list, err := scenario.ListScenarios(ctx, merchantId)
	if err != nil {
		return nil, err
	}

	items := make([]*_scenario.ScenarioItem, 0, len(list))
	for _, s := range list {
		items = append(items, &_scenario.ScenarioItem{
			Id:           s.Id,
			Name:         s.Name,
			Description:  s.Description,
			ScenarioJson: s.ScenarioJson,
			Enabled:      s.Enabled,
			TriggerType:  s.TriggerType,
			TriggerValue: s.TriggerValue,
			CreateTime:   s.CreateTime,
		})
	}

	return &_scenario.ListRes{Scenarios: items}, nil
}
