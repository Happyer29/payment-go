package card

import (
	"fmt"
	"payment-go/internal/models"
)

type UpdateCardDto struct {
	ID          uint    `json:"id"`
	PhonePrefix *string `json:"phone_prefix"`
	PhoneNumber *string `json:"phone_number"`
	CardNumber  *string `json:"card_number"`
	Status      string  `json:"status"`
}

func (dto *UpdateCardDto) Validate() error {
	if dto.CardNumber != nil && len(*dto.CardNumber) == 0 {
		fmt.Errorf("card number cannot be empty")
	}
	if dto.PhonePrefix != nil && len(*dto.PhonePrefix) != 5 {
		return fmt.Errorf("phone_prefix must have length:5")
	}
	if dto.PhoneNumber != nil && len(*dto.PhoneNumber) != 7 {
		return fmt.Errorf("phone_number must have length:7")
	}
	if dto.Status == "" {
		dto.Status = models.CardStatusDisabled
	}
	if dto.Status != models.CardStatusEnabled && dto.Status != models.CardStatusDisabled {
		return fmt.Errorf("got invalid status")
	}

	return nil
}
