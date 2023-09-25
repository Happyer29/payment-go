package models

import (
	"fmt"
	"gorm.io/gorm"
)

type PaymentLink struct {
	gorm.Model
	OrderID         uint `gorm:"index"`
	Order           Order
	CheckMerchantId string  `gorm:"column:check_merchant_id;type:char(255);not null"`
	Amount          float64 `gorm:"column:amount;not null"`
	CardType        string  `gorm:"column:card_type;type:char(63);not null"`
	URL             string  `gorm:"column:url;type:text(1023);not null"`
	TransactionId   string  `gorm:"column:transaction_id;type:char(127)"`
	DatePaid        *uint   `gorm:"column:date_paid"`
	Status          string  `gorm:"column:status;type:char(63);not null"`
}

const StatusNew = "new"
const StatusPending = "pending"
const StatusCompleted = "completed"
const StatusFailed = "failed"
const DefaultStatus = StatusNew

func GetAvailableStatuses() []string {
	return []string{
		StatusNew,
		StatusPending,
		StatusCompleted,
		StatusFailed,
	}
}

func IsStatusValid(status string) bool {
	for _, val := range GetAvailableStatuses() {
		if status == val {
			return true
		}
	}
	return false
}

func (pl *PaymentLink) BeforeSave(tx *gorm.DB) error {
	if pl.Status == "" {
		pl.Status = DefaultStatus
	}
	switch pl.Status {
	case StatusNew:
	case StatusPending:
	case StatusCompleted:
	case StatusFailed:
		break
	default:
		return fmt.Errorf("invalid status: " + pl.Status)
	}
	return nil
}
