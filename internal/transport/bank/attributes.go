package bank

import "fmt"

type AttributesDto struct {
	Merchant *MerchantDto             `json:"merchant"`
	Payment  []*PaymentInformationDto `json:"paymentInformation"`
}

type MerchantDto struct {
	Id string `json:"id"`
}

type PaymentInformationDto struct {
	Id string `json:"id"`
}

func (ad *AttributesDto) Validate() error {
	if len(ad.Merchant.Id) == 0 {
		return fmt.Errorf("merchant.id is empty")
	}
	if len(ad.Payment) == 0 {
		return fmt.Errorf("paymentInformation is empty")
	}
	if len(ad.Payment[0].Id) == 0 {
		return fmt.Errorf("paymentInformation[0].id is empty")
	}
	return nil
}
