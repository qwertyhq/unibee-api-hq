package merchant

import (
	"context"

	_scenario "unibee/api/merchant/scenario"
	_interface "unibee/internal/interface/context"
	"unibee/internal/logic/scenario"
)

func (c *ControllerScenario) ExecutionList(ctx context.Context, req *_scenario.ExecutionListReq) (res *_scenario.ExecutionListRes, err error) {
	merchantId := _interface.GetMerchantId(ctx)

	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.Size
	if size <= 0 {
		size = 20
	}

	list, total, err := scenario.ListExecutions(ctx, merchantId, req.ScenarioId, page, size)
	if err != nil {
		return nil, err
	}

	items := make([]*_scenario.ExecutionItem, 0, len(list))
	for _, e := range list {
		items = append(items, &_scenario.ExecutionItem{
			Id:           e.Id,
			ScenarioId:   e.ScenarioId,
			TriggerData:  e.TriggerData,
			Status:       e.Status,
			CurrentStep:  e.CurrentStep,
			StartedAt:    e.StartedAt,
			FinishedAt:   e.FinishedAt,
			ErrorMessage: e.ErrorMessage,
		})
	}

	return &_scenario.ExecutionListRes{
		Executions: items,
		Total:      total,
	}, nil
}
