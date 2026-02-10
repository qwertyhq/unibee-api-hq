package merchant

import (
	"context"
	"fmt"

	_scenario "unibee/api/merchant/scenario"
	"unibee/internal/logic/scenario"
)

func (c *ControllerScenario) ExecutionDetail(ctx context.Context, req *_scenario.ExecutionDetailReq) (res *_scenario.ExecutionDetailRes, err error) {
	exec, err := scenario.GetExecution(ctx, req.ExecutionId)
	if err != nil {
		return nil, err
	}
	if exec == nil {
		return nil, fmt.Errorf("execution not found")
	}

	stepLogs, err := scenario.GetStepLogs(ctx, req.ExecutionId)
	if err != nil {
		return nil, err
	}

	logItems := make([]*_scenario.StepLogItem, 0, len(stepLogs))
	for _, sl := range stepLogs {
		logItems = append(logItems, &_scenario.StepLogItem{
			Id:           sl.Id,
			StepId:       sl.StepId,
			StepType:     sl.StepType,
			InputData:    sl.InputData,
			OutputData:   sl.OutputData,
			Status:       sl.Status,
			DurationMs:   sl.DurationMs,
			ErrorMessage: sl.ErrorMessage,
		})
	}

	return &_scenario.ExecutionDetailRes{
		Execution: &_scenario.ExecutionItem{
			Id:           exec.Id,
			ScenarioId:   exec.ScenarioId,
			TriggerData:  exec.TriggerData,
			Status:       exec.Status,
			CurrentStep:  exec.CurrentStep,
			StartedAt:    exec.StartedAt,
			FinishedAt:   exec.FinishedAt,
			ErrorMessage: exec.ErrorMessage,
		},
		StepLogs: logItems,
	}, nil
}
