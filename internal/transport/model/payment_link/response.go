package payment_link

import (
	"payment-go/internal/models"
	"payment-go/internal/transport/model/order"
)

type PaymentLinkResponseDto struct {
	ID              uint                    `json:"id"`
	OrderId         uint                    `json:"order_id"`
	Order           *order.OrderResponseDto `json:"order"`
	CheckMerchantId string                  `json:"check_merchant_id"`
	Amount          float64                 `json:"amount"`
	CardType        string                  `json:"card_type"`
	URL             string                  `json:"url"`
	TransactionId   string                  `json:"transaction_id"`
	DatePaid        *uint                   `json:"date_paid"`
	Status          string                  `json:"status"`
}

func FromPaymentLink(link *models.PaymentLink) *PaymentLinkResponseDto {
	return &PaymentLinkResponseDto{
		ID:              link.ID,
		OrderId:         link.OrderID,
		Order:           order.FromOrder(&link.Order),
		CheckMerchantId: link.CheckMerchantId,
		Amount:          link.Amount,
		CardType:        link.CardType,
		URL:             link.URL,
		TransactionId:   link.TransactionId,
		DatePaid:        link.DatePaid,
		Status:          link.Status,
	}
}

func FromPaymentLinks(links []*models.PaymentLink) []*PaymentLinkResponseDto {
	res := make([]*PaymentLinkResponseDto, len(links))
	for i, link := range links {
		res[i] = FromPaymentLink(link)
	}
	return res
}
