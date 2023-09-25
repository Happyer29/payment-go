package card

import "payment-go/internal/models"

type CardResponseDto struct {
	ID          uint     `json:"id"`
	Type        string   `json:"type"`
	Balance     *float64 `json:"balance,omitempty"`
	PhonePrefix *string  `json:"phone_prefix"`
	PhoneNumber *string  `json:"phone_number"`
	CardNumber  *string  `json:"card_number"`
	Status      string   `json:"status"`
}

func FromCard(card *models.Card, info *models.CardInfo) *CardResponseDto {
	res := &CardResponseDto{
		ID:          card.ID,
		Type:        card.Type,
		PhonePrefix: card.PhonePrefix,
		PhoneNumber: card.PhoneNumber,
		CardNumber:  card.CardNumber,
		Status:      card.Status,
	}

	if info != nil {
		res.Balance = &info.Balance
	}

	return res
}
