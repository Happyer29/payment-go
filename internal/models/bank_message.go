package models

import (
	"errors"
	"gorm.io/gorm"
)

const BankMessageStatusNew = "new"
const BankMessageStatusError = "error"
const BankMessageStatusPendingApproval = "pending_approval"
const BankMessageStatusDeclined = "declined"
const BankMessageStatusSuccess = "success"

type BankMessage struct {
	gorm.Model
	SenderPhone   *string `gorm:"sender_phone;type:char(63)"`
	OrderID       uint    `gorm:"order_id;not null"`
	ShopID        uint    `gorm:"shop_id;not null"`
	ReceiverPhone *string `gorm:"receiver_phone;type:char(63)"`
	RawMessage    string  `gorm:"raw_message;text"`
	CardNumber    string  `gorm:"card_number;type:char(63); not null"`
	Amount        float64 `gorm:"amount;not null"`
	Error         string  `gorm:"error;type:char(255);not null"`
	Status        string  `gorm:"status;type:char(63);not null"`
}

var ErrUnknownStatus = errors.New("unknown status")

func (msg *BankMessage) BeforeSave(tx *gorm.DB) error {
	if len(msg.Status) == 0 {
		msg.Status = BankMessageStatusNew
	} else {
		if msg.Status != BankMessageStatusError &&
			msg.Status != BankMessageStatusPendingApproval &&
			msg.Status != BankMessageStatusDeclined &&
			msg.Status != BankMessageStatusSuccess {
			return ErrUnknownStatus
		}
	}
	return nil
}

func (msg *BankMessage) SetStatus(status string) {
	msg.Status = status
}

func (msg *BankMessage) Approve() bool {
	if msg.Status == BankMessageStatusPendingApproval {
		msg.Status = BankMessageStatusSuccess
		return true
	}
	return false
}

func (msg *BankMessage) Decline() bool {
	if msg.Status == BankMessageStatusPendingApproval {
		msg.Status = BankMessageStatusDeclined
		return true
	}
	return false
}
