package shop

import (
	"fmt"
	"payment-go/internal/transport/model/shop/parts"
)

type UpdateShopDto struct {
	ID        uint                  `json:"id"`
	Name      string                `json:"name"`
	Active    bool                  `json:"active"`
	Moderated bool                  `json:"moderated"`
	Webhooks  parts.ShopWebhooksDto `json:"webhooks"`
}

func (dto *UpdateShopDto) Validate() error {
	if dto.ID == 0 {
		return fmt.Errorf("shop id is 0")
	}
	if len(dto.Name) < 3 {
		return fmt.Errorf("shop name is too short")
	}
	if dto.Active && !dto.Moderated {
		return fmt.Errorf("unable to activate shop before moderation")
	}

	return nil
}
