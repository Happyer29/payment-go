package bank

import (
	"net/url"
)

type LinkRequestDto struct {
	CheckMerchant string `json:"checkMerchant"`
	Amount        string `json:"amount"`
	CardType      string `json:"cardType"`
}

type LinkResponseDto struct {
	PaymentUrl string `json:"paymentUrl"`
}

func NewLinkRequestDto(resp *CheckMerchantResponseDto, amount string, cardType string) *LinkRequestDto {
	return &LinkRequestDto{
		CheckMerchant: resp.CheckMerchantId,
		Amount:        amount,
		CardType:      cardType,
	}
}

func (dto *LinkResponseDto) Validate() error {
	_, err := url.Parse(dto.PaymentUrl)
	if err != nil {
		return err
	}
	return nil
}
