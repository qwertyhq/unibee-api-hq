// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MerchantScenarioDelayedTaskDao is the data access object for table merchant_scenario_delayed_task.
type MerchantScenarioDelayedTaskDao struct {
	table   string
	group   string
	columns MerchantScenarioDelayedTaskColumns
}

// MerchantScenarioDelayedTaskColumns defines column names for table merchant_scenario_delayed_task.
type MerchantScenarioDelayedTaskColumns struct {
	Id          string // id
	MerchantId  string // merchantId
	ExecutionId string // execution id
	StepId      string // step id to resume
	ExecuteAt   string // unix timestamp to execute
	Status      string // pending, executed, cancelled
	GmtCreate   string // create time
	GmtModify   string // update time
}

var merchantScenarioDelayedTaskColumns = MerchantScenarioDelayedTaskColumns{
	Id:          "id",
	MerchantId:  "merchant_id",
	ExecutionId: "execution_id",
	StepId:      "step_id",
	ExecuteAt:   "execute_at",
	Status:      "status",
	GmtCreate:   "gmt_create",
	GmtModify:   "gmt_modify",
}

func NewMerchantScenarioDelayedTaskDao() *MerchantScenarioDelayedTaskDao {
	return &MerchantScenarioDelayedTaskDao{
		group:   "default",
		table:   "merchant_scenario_delayed_task",
		columns: merchantScenarioDelayedTaskColumns,
	}
}

func (dao *MerchantScenarioDelayedTaskDao) DB() gdb.DB    { return g.DB(dao.group) }
func (dao *MerchantScenarioDelayedTaskDao) Table() string { return dao.table }
func (dao *MerchantScenarioDelayedTaskDao) Columns() MerchantScenarioDelayedTaskColumns {
	return dao.columns
}
func (dao *MerchantScenarioDelayedTaskDao) Group() string { return dao.group }

func (dao *MerchantScenarioDelayedTaskDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

func (dao *MerchantScenarioDelayedTaskDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
