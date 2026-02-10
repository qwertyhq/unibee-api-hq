// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// MerchantTelegramUserDao is the data access object for table merchant_telegram_user.
type MerchantTelegramUserDao struct {
	table   string
	group   string
	columns MerchantTelegramUserColumns
}

// MerchantTelegramUserColumns defines column names for table merchant_telegram_user.
type MerchantTelegramUserColumns struct {
	Id               string // id
	MerchantId       string // merchantId
	UserId           string // unibee user id
	TelegramChatId   string // telegram chat id
	TelegramUsername string // telegram username
	FirstName        string // telegram first name
	LastName         string // telegram last name
	GmtCreate        string // create time
	GmtModify        string // update time
	IsDeleted        string // 0-UnDeleted, 1-Deleted
	CreateTime       string // create utc time
}

var merchantTelegramUserColumns = MerchantTelegramUserColumns{
	Id:               "id",
	MerchantId:       "merchant_id",
	UserId:           "user_id",
	TelegramChatId:   "telegram_chat_id",
	TelegramUsername: "telegram_username",
	FirstName:        "first_name",
	LastName:         "last_name",
	GmtCreate:        "gmt_create",
	GmtModify:        "gmt_modify",
	IsDeleted:        "is_deleted",
	CreateTime:       "create_time",
}

func NewMerchantTelegramUserDao() *MerchantTelegramUserDao {
	return &MerchantTelegramUserDao{
		group:   "default",
		table:   "merchant_telegram_user",
		columns: merchantTelegramUserColumns,
	}
}

func (dao *MerchantTelegramUserDao) DB() gdb.DB                           { return g.DB(dao.group) }
func (dao *MerchantTelegramUserDao) Table() string                        { return dao.table }
func (dao *MerchantTelegramUserDao) Columns() MerchantTelegramUserColumns { return dao.columns }
func (dao *MerchantTelegramUserDao) Group() string                        { return dao.group }

func (dao *MerchantTelegramUserDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

func (dao *MerchantTelegramUserDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
