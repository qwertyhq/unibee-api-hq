// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MerchantScenarioStepLog is the golang structure for table merchant_scenario_step_log.
type MerchantScenarioStepLog struct {
	Id           uint64      `json:"id"           description:"id"`
	ExecutionId  uint64      `json:"executionId"  description:"execution id"`
	StepId       string      `json:"stepId"       description:"step id within scenario"`
	StepType     string      `json:"stepType"     description:"step type"`
	InputData    string      `json:"inputData"    description:"step input JSON"`
	OutputData   string      `json:"outputData"   description:"step output JSON"`
	Status       string      `json:"status"       description:"success, failed, skipped"`
	DurationMs   int         `json:"durationMs"   description:"execution duration in ms"`
	ErrorMessage string      `json:"errorMessage" description:"error if failed"`
	GmtCreate    *gtime.Time `json:"gmtCreate"    description:"create time"`
}
