// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MerchantScenarioStepLogDao is the data access object for table merchant_scenario_step_log.
type MerchantScenarioStepLogDao struct {
	table   string
	group   string
	columns MerchantScenarioStepLogColumns
}

// MerchantScenarioStepLogColumns defines column names for table merchant_scenario_step_log.
type MerchantScenarioStepLogColumns struct {
	Id           string // id
	ExecutionId  string // execution id
	StepId       string // step id within scenario
	StepType     string // step type
	InputData    string // step input JSON
	OutputData   string // step output JSON
	Status       string // success, failed, skipped
	DurationMs   string // execution duration in ms
	ErrorMessage string // error if failed
	GmtCreate    string // create time
}

var merchantScenarioStepLogColumns = MerchantScenarioStepLogColumns{
	Id:           "id",
	ExecutionId:  "execution_id",
	StepId:       "step_id",
	StepType:     "step_type",
	InputData:    "input_data",
	OutputData:   "output_data",
	Status:       "status",
	DurationMs:   "duration_ms",
	ErrorMessage: "error_message",
	GmtCreate:    "gmt_create",
}

func NewMerchantScenarioStepLogDao() *MerchantScenarioStepLogDao {
	return &MerchantScenarioStepLogDao{
		group:   "default",
		table:   "merchant_scenario_step_log",
		columns: merchantScenarioStepLogColumns,
	}
}

func (dao *MerchantScenarioStepLogDao) DB() gdb.DB                              { return g.DB(dao.group) }
func (dao *MerchantScenarioStepLogDao) Table() string                           { return dao.table }
func (dao *MerchantScenarioStepLogDao) Columns() MerchantScenarioStepLogColumns { return dao.columns }
func (dao *MerchantScenarioStepLogDao) Group() string                           { return dao.group }

func (dao *MerchantScenarioStepLogDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

func (dao *MerchantScenarioStepLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
