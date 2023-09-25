package models

import (
	"fmt"
	"gorm.io/gorm"
	"payment-go/internal/config"
)

const CardWithNumber = "with_number"
const CardWithPhone = "with_phone"

type Card struct {
	gorm.Model
	Type        string    `gorm:"column:type;not null;<-:create"`
	PhonePrefix *string   `gorm:"column:phone_prefix;type:char(63);uniqueIndex:phone,length:20,priority:1"`
	PhoneNumber *string   `gorm:"column:phone_number;type:char(63);uniqueIndex:phone,length:20,priority:2"`
	CardNumber  *string   `gorm:"column:card_number;type:char(32);unique"`
	Status      string    `gorm:"column:status;type:char(63);not null"`
	Stats       CardStats `gorm:"embedded;embeddedPrefix:stats_"`
}

type CardStats struct {
	TotalPaymentSum float64 `gorm:"column:total_payment_sum"`
}

const CardStatusEnabled = "enabled"
const CardStatusDisabled = "disabled"

func NewCardWithNumber(cardNumber string) *Card {
	return &Card{
		Type:        CardWithNumber,
		Status:      CardStatusDisabled, // по умолчанию карта деактивирована
		PhoneNumber: nil,
		PhonePrefix: nil,
		CardNumber:  &cardNumber,
	}
}

func NewCardWithPhone(phonePrefix, phoneNumber string) *Card {
	return &Card{
		Type:        CardWithPhone,
		Status:      CardStatusDisabled, // по умолчанию карта деактивирована
		PhoneNumber: &phoneNumber,
		PhonePrefix: &phonePrefix,
		CardNumber:  nil,
	}
}

func (card *Card) SupportsPaymentMethod(method string) bool {
	if card.Type == CardWithNumber {
		return method == config.PaymentMethodBankTransfer
	}
	if card.Type == CardWithPhone {
		return method == config.PaymentMethodKapitalBank
	}
	return false
}

func (card *Card) GetPhone() string {
	if card.PhonePrefix != nil && card.PhoneNumber != nil {
		return *card.PhonePrefix + *card.PhoneNumber
	}
	return ""
}

func (card *Card) IsActive() bool {
	return card.Status == CardStatusEnabled
}

func (card *Card) Validate() error {
	if card.Type == CardWithNumber && (card.CardNumber == nil || len(*card.CardNumber) == 0) {
		return fmt.Errorf("card number cannot be empty")
	} else if card.Type == CardWithPhone {
		if card.PhonePrefix == nil || len(*card.PhonePrefix) != 5 {
			return fmt.Errorf("phone_prefix must have length:5")
		}
		if card.PhoneNumber == nil || len(*card.PhoneNumber) != 7 {
			return fmt.Errorf("phone_number must have length:7")
		}
	}
	if card.Status != CardStatusDisabled && card.Status != CardStatusEnabled {
		return fmt.Errorf("unknown card status")
	}
	return nil
}

func (card *Card) BeforeSave(tx *gorm.DB) error {
	if card.Type == CardWithPhone {
		card.CardNumber = nil
	} else if card.Type == CardWithNumber {
		card.PhonePrefix = nil
		card.PhoneNumber = nil
	}
	if err := card.Validate(); err != nil {
		return err
	}
	return nil
}

func (card *Card) GetCardNumber() *string {
	return card.CardNumber
}

func (card *Card) GetIdentifier() string {
	if card.Type == CardWithPhone {
		return card.GetPhone()
	}
	return *card.GetCardNumber()
}

func (card *Card) SupportsLocking() bool {
	return card.Type == CardWithNumber
}

func (card *Card) CanBeLocked() bool {
	return card.SupportsLocking() && card.ID > 0 && card.Status == CardStatusEnabled
}
