package models

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	string2 "payment-go/internal/utils/random/string"
	"regexp"
	"strings"
)

type Shop struct {
	gorm.Model
	Name          string       `gorm:"column:name;type:char(63);not null"`
	Host          string       `gorm:"column:host;type:char(255)"`
	OwnerId       uint         `gorm:"column:owner_id"`
	Active        bool         `gorm:"column:active;not null;default:false"`
	HostValidated bool         `gorm:"column:host_validated;not null;default:false"`
	Moderated     bool         `gorm:"column:moderated;not null;default:false"`
	Keys          ShopKeys     `gorm:"embedded"`
	Webhooks      ShopWebhooks `gorm:"embedded;embeddedPrefix:webhook_"`
}

type ShopKeys struct {
	PublicKey  uuid.UUID `gorm:"column:public_key;type:char(36);unique;not null;<-:create" json:"public_key"`
	PrivateKey string    `gorm:"column:private_key;type:char(255);not null" json:"private_key"`
}

type ShopWebhooks struct {
	LinkCreated       *string `gorm:"column:link_created;type:text(1023)"`
	OnSuccess         *string `gorm:"column:on_success;type:text(1023)"`
	OnFailure         *string `gorm:"column:on_failure;type:text(1023)"`
	OnWithdrawUpdated *string `gorm:"column:on_withdraw_updated;type:text(1023)"`
}

var rsg = string2.New(string2.LettersAnyCase + string2.Numbers)

const PrivateKeyLength = 120

func NewShop(name, host string, ownerId uint) (*Shop, error) {
	if !ValidateHost(host) {
		return nil, fmt.Errorf("invalid shop host")
	}

	sh := &Shop{
		Name:          name,
		Host:          host,
		OwnerId:       ownerId,
		Active:        false,
		HostValidated: false,
		Moderated:     false,
		Keys: ShopKeys{
			PublicKey:  uuid.New(),
			PrivateKey: NewPrivateKey(),
		},
	}
	return sh, nil
}

func NewPrivateKey() string {
	return rsg.String(PrivateKeyLength)
}

const HOST_REGEXP = "^https?:\\/\\/(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)+([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9])$"

func ValidateHost(host string) bool {
	if len(host) < 10 {
		return false
	}

	blackList := []string{"localhost", "127.0", "192.168", "0.0"}
	for _, rule := range blackList {
		if strings.Contains(host, rule) {
			return false
		}
	}

	re, err := regexp.Compile(HOST_REGEXP)
	if err != nil {
		return false
	}

	return re.MatchString(host)
}

func (shop *Shop) RegeneratePrivateKey() {
	shop.Keys.PrivateKey = NewPrivateKey()
}

func (shop *Shop) BeforeSave(tx *gorm.DB) error {
	if shop.OwnerId == 0 {
		return fmt.Errorf("shop.ownerId must be greater than 0")
	}
	if len(shop.Host) == 0 {
		shop.Active = false
	}
	if !shop.Moderated {
		shop.Active = false
	}
	if shop.ID == 0 {
		shop.Active = false
		shop.HostValidated = false
		shop.Moderated = false
	}
	if err := shop.ValidateWebhooks(); err != nil {
		return err
	}

	return nil
}

func (shop *Shop) ValidateWebhooks() error {
	wh := shop.Webhooks
	if wh.LinkCreated != nil && len(*wh.LinkCreated) != 0 && (*wh.LinkCreated)[0] != '/' {
		return fmt.Errorf("webhook must start from / if specified")
	}
	if wh.OnSuccess != nil && len(*wh.OnSuccess) != 0 && (*wh.OnSuccess)[0] != '/' {
		return fmt.Errorf("webhook must start from / if specified")
	}
	if wh.OnFailure != nil && len(*wh.OnFailure) != 0 && (*wh.OnFailure)[0] != '/' {
		return fmt.Errorf("webhook must start from / if specified")
	}
	if wh.OnWithdrawUpdated != nil && len(*wh.OnWithdrawUpdated) != 0 && (*wh.OnWithdrawUpdated)[0] != '/' {
		return fmt.Errorf("webhook must start from / if specified")
	}
	return nil
}

func (shop *Shop) IsAvailable() bool {
	return shop.Active && shop.Moderated
}
