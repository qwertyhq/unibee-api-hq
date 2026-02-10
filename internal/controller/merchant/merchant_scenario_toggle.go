package merchant

import (
	"context"

	_scenario "unibee/api/merchant/scenario"
	_interface "unibee/internal/interface/context"
	"unibee/internal/logic/scenario"
)

func (c *ControllerScenario) Toggle(ctx context.Context, req *_scenario.ToggleReq) (res *_scenario.ToggleRes, err error) {
	merchantId := _interface.GetMerchantId(ctx)
	err = scenario.ToggleScenario(ctx, merchantId, req.ScenarioId, req.Enabled)
	if err != nil {
		return nil, err
	}

	// Restart bot polling if a bot_command/button_click scenario was toggled
	sc, _ := scenario.GetScenario(ctx, merchantId, req.ScenarioId)
	if sc != nil && (sc.TriggerType == scenario.TriggerBotCommand || sc.TriggerType == scenario.TriggerButtonClick) {
		if req.Enabled {
			_ = scenario.StartBotPolling(ctx, merchantId)
		}
	}

	return &_scenario.ToggleRes{}, nil
}
