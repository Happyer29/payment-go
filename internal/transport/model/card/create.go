package card

import (
	"fmt"
	"payment-go/internal/models"
)

type CreateCardDto struct {
	Type        string `json:"type"`
	CardNumber  string `json:"card_number"`
	PhonePrefix string `json:"phone_prefix"`
	PhoneNumber string `json:"phone_number"`
	Status      string `json:"status"`
}

func (dto *CreateCardDto) Validate() error {
	if dto.Type != models.CardWithNumber && dto.Type != models.CardWithPhone {
		return fmt.Errorf("unknown card type")
	}
	if dto.Type == models.CardWithNumber && len(dto.CardNumber) == 0 {
		return fmt.Errorf("card number cannot be empty")
	}
	if dto.Type == models.CardWithPhone {
		if len(dto.PhonePrefix) != 5 {
			return fmt.Errorf("phone_prefix must have length:5")
		}
		if len(dto.PhoneNumber) != 7 {
			return fmt.Errorf("phone_number must have length:7")
		}
	}

	if dto.Status == "" {
		dto.Status = models.CardStatusDisabled
	} else if dto.Status != models.CardStatusEnabled && dto.Status != models.CardStatusDisabled {
		return fmt.Errorf("got invalid status")
	}

	return nil
}

func (dto *CreateCardDto) ToCard() (*models.Card, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}
	if dto.Type == models.CardWithNumber {
		return models.NewCardWithNumber(dto.CardNumber), nil
	}
	return models.NewCardWithPhone(dto.PhonePrefix, dto.PhoneNumber), nil
}
