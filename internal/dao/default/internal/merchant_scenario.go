// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MerchantScenarioDao is the data access object for table merchant_scenario.
type MerchantScenarioDao struct {
	table   string
	group   string
	columns MerchantScenarioColumns
}

// MerchantScenarioColumns defines and stores column names for table merchant_scenario.
type MerchantScenarioColumns struct {
	Id           string // id
	MerchantId   string // merchantId
	Name         string // scenario name
	Description  string // scenario description
	ScenarioJson string // scenario JSON DSL
	Enabled      string // 0-disabled, 1-enabled
	TriggerType  string // trigger type
	TriggerValue string // trigger value
	GmtCreate    string // create time
	GmtModify    string // update time
	IsDeleted    string // 0-UnDeleted, 1-Deleted
	CreateTime   string // create utc time
}

var merchantScenarioColumns = MerchantScenarioColumns{
	Id:           "id",
	MerchantId:   "merchant_id",
	Name:         "name",
	Description:  "description",
	ScenarioJson: "scenario_json",
	Enabled:      "enabled",
	TriggerType:  "trigger_type",
	TriggerValue: "trigger_value",
	GmtCreate:    "gmt_create",
	GmtModify:    "gmt_modify",
	IsDeleted:    "is_deleted",
	CreateTime:   "create_time",
}

// NewMerchantScenarioDao creates and returns a new DAO object for table data access.
func NewMerchantScenarioDao() *MerchantScenarioDao {
	return &MerchantScenarioDao{
		group:   "default",
		table:   "merchant_scenario",
		columns: merchantScenarioColumns,
	}
}

func (dao *MerchantScenarioDao) DB() gdb.DB                       { return g.DB(dao.group) }
func (dao *MerchantScenarioDao) Table() string                    { return dao.table }
func (dao *MerchantScenarioDao) Columns() MerchantScenarioColumns { return dao.columns }
func (dao *MerchantScenarioDao) Group() string                    { return dao.group }

func (dao *MerchantScenarioDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

func (dao *MerchantScenarioDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
