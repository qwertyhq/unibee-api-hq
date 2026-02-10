// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MerchantScenario is the golang structure for table merchant_scenario.
type MerchantScenario struct {
	Id           uint64      `json:"id"           description:"id"`
	MerchantId   uint64      `json:"merchantId"   description:"merchantId"`
	Name         string      `json:"name"         description:"scenario name"`
	Description  string      `json:"description"  description:"scenario description"`
	ScenarioJson string      `json:"scenarioJson" description:"scenario JSON DSL"`
	Enabled      int         `json:"enabled"      description:"0-disabled, 1-enabled"`
	TriggerType  string      `json:"triggerType"   description:"trigger type"`
	TriggerValue string      `json:"triggerValue"  description:"trigger value"`
	GmtCreate    *gtime.Time `json:"gmtCreate"    description:"create time"`
	GmtModify    *gtime.Time `json:"gmtModify"    description:"update time"`
	IsDeleted    int         `json:"isDeleted"    description:"0-UnDeleted, 1-Deleted"`
	CreateTime   int64       `json:"createTime"   description:"create utc time"`
}
