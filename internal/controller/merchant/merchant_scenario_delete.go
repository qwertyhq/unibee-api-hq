package merchant

import (
	"context"

	_scenario "unibee/api/merchant/scenario"
	_interface "unibee/internal/interface/context"
	"unibee/internal/logic/scenario"
)

func (c *ControllerScenario) Delete(ctx context.Context, req *_scenario.DeleteReq) (res *_scenario.DeleteRes, err error) {
	merchantId := _interface.GetMerchantId(ctx)
	err = scenario.DeleteScenario(ctx, merchantId, req.ScenarioId)
	if err != nil {
		return nil, err
	}
	return &_scenario.DeleteRes{}, nil
}
