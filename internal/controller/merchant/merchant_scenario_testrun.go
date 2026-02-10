package merchant

import (
	"context"
	"encoding/json"
	"fmt"

	_scenario "unibee/api/merchant/scenario"
	_interface "unibee/internal/interface/context"
	"unibee/internal/logic/scenario"
)

func (c *ControllerScenario) TestRun(ctx context.Context, req *_scenario.TestRunReq) (res *_scenario.TestRunRes, err error) {
	merchantId := _interface.GetMerchantId(ctx)

	sc, err := scenario.GetScenario(ctx, merchantId, req.ScenarioId)
	if err != nil {
		return nil, err
	}
	if sc == nil {
		return nil, fmt.Errorf("scenario not found")
	}

	triggerData := make(map[string]interface{})
	if req.TriggerData != "" {
		if err := json.Unmarshal([]byte(req.TriggerData), &triggerData); err != nil {
			return nil, fmt.Errorf("invalid trigger data JSON: %w", err)
		}
	}

	// Run in background
	go scenario.RunScenarioByIds(ctx, merchantId, sc.Id, sc.ScenarioJson, triggerData)

	// We return a placeholder execId — the real one will be created asynchronously.
	// In a future version, we can make this synchronous.
	return &_scenario.TestRunRes{ExecutionId: 0}, nil
}
