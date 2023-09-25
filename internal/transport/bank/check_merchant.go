package bank

import (
	"fmt"
	"payment-go/internal/config"
)

type CheckMerchantRequestDto struct {
	Merchant  string                     `json:"merchant"`
	Payment   string                     `json:"paymentInformation"`
	Values    CheckMerchantRequestValues `json:"values"`
	Step      int                        `json:"step"`
	Bonusable bool                       `json:"bonusable"`
}

type CheckMerchantRequestValues struct {
	ListOfValues []*CheckMerchantPhone `json:"list_of_values"`
}

type CheckMerchantPhone struct {
	AttributeName string `json:"attribute_name"`
	Value         string `json:"value"`
	Prefix        string `json:"prefix"`
}

type CheckMerchantResponseDto struct {
	CheckMerchantId string `json:"checkMerchantId"`
}

func NewCheckMerchantRequestDto(atts *AttributesDto, phone *config.Phone) *CheckMerchantRequestDto {
	return &CheckMerchantRequestDto{
		Merchant: atts.Merchant.Id,
		Payment:  atts.Payment[0].Id,
		Values: CheckMerchantRequestValues{
			ListOfValues: []*CheckMerchantPhone{
				{
					AttributeName: "phoneNumber",
					Value:         phone.Number,
					Prefix:        phone.Prefix,
				},
			},
		},
		Step:      2,
		Bonusable: false,
	}
}

func (dto *CheckMerchantResponseDto) Validate() error {
	if len(dto.CheckMerchantId) == 0 {
		return fmt.Errorf("checkMerchantId is empty")
	}
	return nil
}
