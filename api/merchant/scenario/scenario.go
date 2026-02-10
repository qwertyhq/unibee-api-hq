package scenario

import (
	"github.com/gogf/gf/v2/frame/g"
)

// New — create a scenario
type NewReq struct {
	g.Meta       `path:"/new" tags:"Scenario" method:"post" summary:"Create Scenario"`
	Name         string `json:"name" dc:"Scenario name" v:"required"`
	Description  string `json:"description" dc:"Scenario description"`
	ScenarioJson string `json:"scenarioJson" dc:"Scenario JSON DSL" v:"required"`
}
type NewRes struct {
	ScenarioId uint64 `json:"scenarioId" dc:"Created scenario ID"`
}

// Edit — update a scenario
type EditReq struct {
	g.Meta       `path:"/edit" tags:"Scenario" method:"post" summary:"Edit Scenario"`
	ScenarioId   uint64 `json:"scenarioId" dc:"Scenario ID" v:"required"`
	Name         string `json:"name" dc:"Scenario name" v:"required"`
	Description  string `json:"description" dc:"Scenario description"`
	ScenarioJson string `json:"scenarioJson" dc:"Scenario JSON DSL" v:"required"`
}
type EditRes struct {
}

// Delete — soft-delete a scenario
type DeleteReq struct {
	g.Meta     `path:"/delete" tags:"Scenario" method:"post" summary:"Delete Scenario"`
	ScenarioId uint64 `json:"scenarioId" dc:"Scenario ID" v:"required"`
}
type DeleteRes struct {
}

// Toggle — enable/disable a scenario
type ToggleReq struct {
	g.Meta     `path:"/toggle" tags:"Scenario" method:"post" summary:"Toggle Scenario Enabled/Disabled"`
	ScenarioId uint64 `json:"scenarioId" dc:"Scenario ID" v:"required"`
	Enabled    bool   `json:"enabled" dc:"Enabled flag"`
}
type ToggleRes struct {
}

// List — list all scenarios for merchant
type ListReq struct {
	g.Meta `path:"/list" tags:"Scenario" method:"get" summary:"List Scenarios"`
}
type ListRes struct {
	Scenarios []*ScenarioItem `json:"scenarios" dc:"Scenario list"`
}
type ScenarioItem struct {
	Id           uint64 `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	ScenarioJson string `json:"scenarioJson"`
	Enabled      int    `json:"enabled"`
	TriggerType  string `json:"triggerType"`
	TriggerValue string `json:"triggerValue"`
	CreateTime   int64  `json:"createTime"`
}

// Detail — get scenario details
type DetailReq struct {
	g.Meta     `path:"/detail" tags:"Scenario" method:"get" summary:"Get Scenario Detail"`
	ScenarioId uint64 `json:"scenarioId" dc:"Scenario ID" v:"required"`
}
type DetailRes struct {
	Scenario *ScenarioItem `json:"scenario" dc:"Scenario detail"`
}

// TestRun — test-run a scenario with sample data
type TestRunReq struct {
	g.Meta      `path:"/test_run" tags:"Scenario" method:"post" summary:"Test Run Scenario"`
	ScenarioId  uint64 `json:"scenarioId" dc:"Scenario ID" v:"required"`
	TriggerData string `json:"triggerData" dc:"Trigger data JSON (optional)"`
}
type TestRunRes struct {
	ExecutionId uint64 `json:"executionId" dc:"Execution ID"`
}

// ExecutionList — list execution history
type ExecutionListReq struct {
	g.Meta     `path:"/execution_list" tags:"Scenario" method:"get" summary:"List Scenario Executions"`
	ScenarioId uint64 `json:"scenarioId" dc:"Filter by scenario ID (optional)"`
	Page       int    `json:"page" dc:"Page number" d:"1"`
	Size       int    `json:"size" dc:"Page size" d:"20"`
}
type ExecutionListRes struct {
	Executions []*ExecutionItem `json:"executions" dc:"Execution list"`
	Total      int              `json:"total" dc:"Total count"`
}
type ExecutionItem struct {
	Id           uint64 `json:"id"`
	ScenarioId   uint64 `json:"scenarioId"`
	TriggerData  string `json:"triggerData"`
	Status       string `json:"status"`
	CurrentStep  string `json:"currentStep"`
	StartedAt    int64  `json:"startedAt"`
	FinishedAt   int64  `json:"finishedAt"`
	ErrorMessage string `json:"errorMessage"`
}

// ExecutionDetail — get execution detail with step logs
type ExecutionDetailReq struct {
	g.Meta      `path:"/execution_detail" tags:"Scenario" method:"get" summary:"Get Execution Detail"`
	ExecutionId uint64 `json:"executionId" dc:"Execution ID" v:"required"`
}
type ExecutionDetailRes struct {
	Execution *ExecutionItem `json:"execution" dc:"Execution detail"`
	StepLogs  []*StepLogItem `json:"stepLogs" dc:"Step execution logs"`
}
type StepLogItem struct {
	Id           uint64 `json:"id"`
	StepId       string `json:"stepId"`
	StepType     string `json:"stepType"`
	InputData    string `json:"inputData"`
	OutputData   string `json:"outputData"`
	Status       string `json:"status"`
	DurationMs   int    `json:"durationMs"`
	ErrorMessage string `json:"errorMessage"`
}

// ActionList — list available action types
type ActionListReq struct {
	g.Meta `path:"/action_list" tags:"Scenario" method:"get" summary:"List Available Actions"`
}
type ActionListRes struct {
	Actions []*ActionType `json:"actions" dc:"Available action types"`
}
type ActionType struct {
	Type        string   `json:"type" dc:"Action type identifier"`
	Name        string   `json:"name" dc:"Human-readable name"`
	Description string   `json:"description" dc:"Action description"`
	Params      []string `json:"params" dc:"Required/optional params"`
}

// TriggerList — list available trigger types
type TriggerListReq struct {
	g.Meta `path:"/trigger_list" tags:"Scenario" method:"get" summary:"List Available Triggers"`
}
type TriggerListRes struct {
	Triggers []*TriggerType `json:"triggers" dc:"Available trigger types"`
}
type TriggerType struct {
	Type        string `json:"type" dc:"Trigger type identifier"`
	Name        string `json:"name" dc:"Human-readable name"`
	Description string `json:"description" dc:"Trigger description"`
}

// Validate — validate scenario JSON DSL
type ValidateReq struct {
	g.Meta       `path:"/validate" tags:"Scenario" method:"post" summary:"Validate Scenario JSON"`
	ScenarioJson string `json:"scenarioJson" dc:"Scenario JSON DSL to validate" v:"required"`
}
type ValidateRes struct {
	Valid  bool     `json:"valid" dc:"Whether the scenario is valid"`
	Errors []string `json:"errors" dc:"Validation errors"`
}
