package merchant

import (
	"context"

	_scenario "unibee/api/merchant/scenario"
	_interface "unibee/internal/interface/context"
	"unibee/internal/logic/scenario"
)

func (c *ControllerScenario) New(ctx context.Context, req *_scenario.NewReq) (res *_scenario.NewRes, err error) {
	merchantId := _interface.GetMerchantId(ctx)

	dsl, err := scenario.ParseDSL(req.ScenarioJson)
	if err != nil {
		return nil, err
	}

	row, err := scenario.CreateScenario(ctx, merchantId, req.Name, req.Description, dsl)
	if err != nil {
		return nil, err
	}

	return &_scenario.NewRes{ScenarioId: row.Id}, nil
}
