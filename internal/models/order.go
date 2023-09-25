package models

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-go/internal/config"
)
import _ "github.com/google/uuid"

type Order struct {
	gorm.Model
	// Number Запрещено редактировать. Указывается при создании
	Number  uuid.UUID `gorm:"column:number;type:char(36);unique;not null;<-:create"`
	Payload string    `gorm:"column:payload;type:text(511);not null"`
	ShopID  uint      `gorm:"index:shop_id"`
	Shop    Shop      //`gorm:"foreignKey:ID;references:shop_id;constraint:OnUpdate:CASCADE,OnDelete:SET DEFAULT;"`

	// Запрещено редактировать
	PaymentMethod string  `gorm:"column:payment_method;type:char(63);<-:create"`
	CardID        uint    `gorm:"index:card_id"`
	Card          Card    //`gorm:"foreignKey:ID;references:card_id;constraint:OnUpdate:CASCADE,OnDelete:SET DEFAULT;"`
	Amount        float64 `gorm:"column:amount;not null"`
	DatePaid      *uint   `gorm:"column:date_paid;index"`
	Status        string  `gorm:"column:status;type:char(63);not null"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) (err error) {
	o.Number = uuid.New()
	if len(o.Status) == 0 {
		o.Status = DefaultStatus
	}
	assertMethod := false
	for _, method := range config.GetConfig().Bank.PaymentMethods {
		if method == o.PaymentMethod {
			assertMethod = true
			break
		}
	}
	if !assertMethod {
		return fmt.Errorf("unknown payment method")
	}

	return nil
}

func (o *Order) GetCardId() uint {
	return o.Card.ID
}

func (o *Order) GetShopId() uint {
	return o.Shop.ID
}

func (o *Order) IsFinished() bool {
	return o.Status == StatusCompleted || o.Status == StatusFailed
}

func (o *Order) GetCardType() string {
	if o.PaymentMethod == config.PaymentMethodKapitalBank {
		return config.CardTypeVisa
	}
	return config.CardTypeNone
}

func (o *Order) HaveLink() bool {
	return o.PaymentMethod == config.PaymentMethodKapitalBank
}

func (o *Order) GetCardIfPublic() *string {
	if o.HaveLink() {
		return nil
	}
	cardNumber := o.Card.GetCardNumber()
	return cardNumber
}
