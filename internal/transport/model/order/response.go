package order

import (
	"payment-go/internal/models"
	"payment-go/internal/transport/model/shop"
)

type OrderResponseDto struct {
	ID            uint                  `json:"id"`
	Payload       string                `json:"payload"`
	ShopID        uint                  `json:"shop_id"`
	OwnerID       uint                  `json:"owner_id"`
	Shop          *shop.ShopResponseDto `json:"shop"`
	Number        string                `json:"number"`
	PaymentMethod string                `json:"payment_method"`
	CardID        uint                  `json:"card_id"`
	Amount        float64               `json:"amount"`
	DatePaid      *uint                 `json:"date_paid"`
	Status        string                `json:"status"`
}

func FromOrder(ord *models.Order) *OrderResponseDto {
	return &OrderResponseDto{
		ID:            ord.ID,
		Payload:       ord.Payload,
		ShopID:        ord.GetShopId(),
		OwnerID:       ord.Shop.OwnerId,
		Shop:          shop.FromShop(&ord.Shop),
		Number:        ord.Number.String(),
		PaymentMethod: ord.PaymentMethod,
		CardID:        ord.GetCardId(),
		Amount:        ord.Amount,
		DatePaid:      ord.DatePaid,
		Status:        ord.Status,
	}
}

func FromOrders(orders []*models.Order) []*OrderResponseDto {
	var res = make([]*OrderResponseDto, len(orders))
	for i, ord := range orders {
		res[i] = FromOrder(ord)
	}
	return res
}
