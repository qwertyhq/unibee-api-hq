// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MerchantScenarioDelayedTask is the golang structure for table merchant_scenario_delayed_task.
type MerchantScenarioDelayedTask struct {
	Id          uint64      `json:"id"          description:"id"`
	MerchantId  uint64      `json:"merchantId"  description:"merchantId"`
	ExecutionId uint64      `json:"executionId" description:"execution id"`
	StepId      string      `json:"stepId"      description:"step id to resume"`
	ExecuteAt   int64       `json:"executeAt"   description:"unix timestamp to execute"`
	Status      string      `json:"status"      description:"pending, executed, cancelled"`
	GmtCreate   *gtime.Time `json:"gmtCreate"   description:"create time"`
	GmtModify   *gtime.Time `json:"gmtModify"   description:"update time"`
}
