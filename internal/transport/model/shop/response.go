package shop

import (
	"payment-go/internal/models"
	"payment-go/internal/transport/model/shop/parts"
)

type ShopResponseDto struct {
	ID            uint                  `json:"id"`
	Name          string                `json:"name"`
	Host          string                `json:"host"`
	OwnerId       uint                  `json:"owner_id"`
	Active        bool                  `json:"active"`
	HostValidated bool                  `json:"host_validated"`
	Moderated     bool                  `json:"moderated"`
	PublicKey     string                `json:"public_key"`
	ConfirmCode   string                `json:"confirm_code"`
	Webhooks      parts.ShopWebhooksDto `json:"webhooks"`
}

func FromShop(entity *models.Shop) *ShopResponseDto {
	return &ShopResponseDto{
		ID:            entity.ID,
		Name:          entity.Name,
		Host:          entity.Host,
		OwnerId:       entity.OwnerId,
		Active:        entity.Active,
		HostValidated: entity.HostValidated,
		Moderated:     entity.Moderated,
		PublicKey:     entity.Keys.PublicKey.String(),
		Webhooks: parts.ShopWebhooksDto{
			LinkCreated:       entity.Webhooks.LinkCreated,
			OnSuccess:         entity.Webhooks.OnSuccess,
			OnFailure:         entity.Webhooks.OnFailure,
			OnWithdrawUpdated: entity.Webhooks.OnWithdrawUpdated,
		},
	}
}

func FromShopWithConfirmCode(entity *models.Shop, code string) *ShopResponseDto {
	dto := FromShop(entity)
	dto.ConfirmCode = code
	return dto
}

func FromShops(entities []*models.Shop) []*ShopResponseDto {
	var res = make([]*ShopResponseDto, len(entities))
	for i, sh := range entities {
		res[i] = FromShop(sh)
	}
	return res
}
