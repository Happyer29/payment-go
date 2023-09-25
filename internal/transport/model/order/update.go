package order

import (
	"fmt"
	"payment-go/internal/models"
)

type UpdateOrderDto struct {
	ID uint `json:"id"`
	//UserId   uint `json:"user_id"`
	//ShopId   uint `json:"shop_id"`
	//CardType string `json:"card_type"`
	//CardID   uint `json:"card_id"`
	//Amount   float64 `json:"amount"`
	Status string `json:"status"`
}

func (dto *UpdateOrderDto) Validate() error {
	if !models.IsStatusValid(dto.Status) {
		return fmt.Errorf("invalid status: " + dto.Status)
	}

	return nil
}
