// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MerchantScenarioExecution is the golang structure for table merchant_scenario_execution.
type MerchantScenarioExecution struct {
	Id           uint64      `json:"id"           description:"id"`
	MerchantId   uint64      `json:"merchantId"   description:"merchantId"`
	ScenarioId   uint64      `json:"scenarioId"   description:"scenario id"`
	TriggerData  string      `json:"triggerData"  description:"trigger input data JSON"`
	Status       string      `json:"status"       description:"pending, running, completed, failed, waiting"`
	CurrentStep  string      `json:"currentStep"  description:"current step id"`
	Variables    string      `json:"variables"    description:"current variables JSON"`
	StartedAt    int64       `json:"startedAt"    description:"start utc time"`
	FinishedAt   int64       `json:"finishedAt"   description:"finish utc time"`
	ErrorMessage string      `json:"errorMessage" description:"error message if failed"`
	GmtCreate    *gtime.Time `json:"gmtCreate"    description:"create time"`
	GmtModify    *gtime.Time `json:"gmtModify"    description:"update time"`
	CreateTime   int64       `json:"createTime"   description:"create utc time"`
}
