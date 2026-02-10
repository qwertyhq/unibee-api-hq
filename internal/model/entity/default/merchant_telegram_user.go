// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// MerchantTelegramUser is the golang structure for table merchant_telegram_user.
type MerchantTelegramUser struct {
	Id               uint64      `json:"id"               description:"id"`
	MerchantId       uint64      `json:"merchantId"       description:"merchantId"`
	UserId           uint64      `json:"userId"           description:"unibee user id"`
	TelegramChatId   string      `json:"telegramChatId"   description:"telegram chat id"`
	TelegramUsername string      `json:"telegramUsername" description:"telegram username"`
	FirstName        string      `json:"firstName"        description:"telegram first name"`
	LastName         string      `json:"lastName"         description:"telegram last name"`
	GmtCreate        *gtime.Time `json:"gmtCreate"        description:"create time"`
	GmtModify        *gtime.Time `json:"gmtModify"        description:"update time"`
	IsDeleted        int         `json:"isDeleted"        description:"0-UnDeleted, 1-Deleted"`
	CreateTime       int64       `json:"createTime"       description:"create utc time"`
}
