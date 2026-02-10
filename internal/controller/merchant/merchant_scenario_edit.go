package merchant

import (
	"context"

	_scenario "unibee/api/merchant/scenario"
	_interface "unibee/internal/interface/context"
	"unibee/internal/logic/scenario"
)

func (c *ControllerScenario) Edit(ctx context.Context, req *_scenario.EditReq) (res *_scenario.EditRes, err error) {
	merchantId := _interface.GetMerchantId(ctx)

	dsl, err := scenario.ParseDSL(req.ScenarioJson)
	if err != nil {
		return nil, err
	}

	err = scenario.UpdateScenario(ctx, merchantId, req.ScenarioId, req.Name, req.Description, dsl)
	if err != nil {
		return nil, err
	}

	return &_scenario.EditRes{}, nil
}
