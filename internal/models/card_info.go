package models

import "gorm.io/gorm"

type CardInfo struct {
	gorm.Model
	CardID  uint    `gorm:"column:card_id;unique;not null;<-:create"`
	Balance float64 `gorm:"column:balance;not null"`
}
