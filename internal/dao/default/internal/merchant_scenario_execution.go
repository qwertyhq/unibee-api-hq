// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MerchantScenarioExecutionDao is the data access object for table merchant_scenario_execution.
type MerchantScenarioExecutionDao struct {
	table   string
	group   string
	columns MerchantScenarioExecutionColumns
}

// MerchantScenarioExecutionColumns defines and stores column names for table merchant_scenario_execution.
type MerchantScenarioExecutionColumns struct {
	Id           string // id
	MerchantId   string // merchantId
	ScenarioId   string // scenario id
	TriggerData  string // trigger input data JSON
	Status       string // pending, running, completed, failed, waiting
	CurrentStep  string // current step id
	Variables    string // current variables JSON
	StartedAt    string // start utc time
	FinishedAt   string // finish utc time
	ErrorMessage string // error message if failed
	GmtCreate    string // create time
	GmtModify    string // update time
	CreateTime   string // create utc time
}

var merchantScenarioExecutionColumns = MerchantScenarioExecutionColumns{
	Id:           "id",
	MerchantId:   "merchant_id",
	ScenarioId:   "scenario_id",
	TriggerData:  "trigger_data",
	Status:       "status",
	CurrentStep:  "current_step",
	Variables:    "variables",
	StartedAt:    "started_at",
	FinishedAt:   "finished_at",
	ErrorMessage: "error_message",
	GmtCreate:    "gmt_create",
	GmtModify:    "gmt_modify",
	CreateTime:   "create_time",
}

func NewMerchantScenarioExecutionDao() *MerchantScenarioExecutionDao {
	return &MerchantScenarioExecutionDao{
		group:   "default",
		table:   "merchant_scenario_execution",
		columns: merchantScenarioExecutionColumns,
	}
}

func (dao *MerchantScenarioExecutionDao) DB() gdb.DB    { return g.DB(dao.group) }
func (dao *MerchantScenarioExecutionDao) Table() string { return dao.table }
func (dao *MerchantScenarioExecutionDao) Columns() MerchantScenarioExecutionColumns {
	return dao.columns
}
func (dao *MerchantScenarioExecutionDao) Group() string { return dao.group }

func (dao *MerchantScenarioExecutionDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

func (dao *MerchantScenarioExecutionDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
