package config

import (
	"fmt"
	"time"
)

type Phone struct {
	Prefix string `json:"prefix"`
	Number string `json:"number"`
}
type Links struct {
	GetAttsUrl       string `json:"get_atts_url"`
	CheckMerchantUrl string `json:"check_merchant_url"`
	PaymentUrl       string `json:"payment_url"`
	CheckPaymentUrl  string `json:"check_payment_url"`
	TransactionInfo  string `json:"transaction_info"`
}

type BankConfig struct {
	Host            string   `json:"host"`
	Links           Links    `json:"links"`
	CardTypes       []string `json:"-"`
	PaymentMethods  []string `json:"-"`
	CardTypeMapping map[string]string
}

const PaymentMethodBankTransfer = "bank_transfer"
const PaymentMethodKapitalBank = "kapital_bank"

const CardTypeNone = "none"
const CardTypeVisa = "visa"
const CardTypeMastercard = "mastercard"

const BankSMSWebhookPassword = "K5G9o1xDSMhw8dM8P74AtGxEGhxh7V6v"
const CardLockingTimeout = 10 * time.Minute
const CardLockerGCInterval = 30 * time.Second
const CardDisableAmount = 5000 // При достижении этой отметки карта будет заблокирована

func buildBankConfig() (*BankConfig, error) {
	conf := &BankConfig{}
	if err := readJSONConfig("bank.json", conf); err != nil {
		return nil, err
	}

	conf.PaymentMethods = []string{
		PaymentMethodBankTransfer,
		PaymentMethodKapitalBank,
	}
	conf.CardTypes = []string{
		CardTypeNone,
		CardTypeVisa,
		CardTypeMastercard,
	}
	conf.CardTypeMapping = map[string]string{
		CardTypeVisa:       "Visa",
		CardTypeMastercard: "Mastercard",
	}

	if err := validateBankConfig(conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func validateBankConfig(conf *BankConfig) error {
	// host
	if !validateURL(conf.Host) {
		return fmt.Errorf("bank.host is not a correct URL")
	}

	// links
	if !validateURL(conf.Links.GetAttsUrl) {
		return fmt.Errorf("bank.links.get_atts_url is not a correct URL")
	}
	if !validateURL(conf.Links.CheckMerchantUrl) {
		return fmt.Errorf("bank.links.check_merchant_url is not a correct URL")
	}
	if !validateURL(conf.Links.PaymentUrl) {
		return fmt.Errorf("bank.links.payment_url is not a correct URL")
	}
	if !validateURL(conf.Links.CheckPaymentUrl) {
		return fmt.Errorf("bank.links.check_payment_url is not a correct URL")
	}
	if !validateURL(conf.Links.TransactionInfo) {
		return fmt.Errorf("bank.links.transaction_info is not a correct URL")
	}

	// card types
	if len(conf.CardTypes) == 0 {
		return fmt.Errorf("bank.card_types is empty")
	}
	for _, cardType := range conf.CardTypes {
		if len(cardType) == 0 {
			return fmt.Errorf("bank.card_types cannot contain empty strings")
		}
	}

	return nil
}
