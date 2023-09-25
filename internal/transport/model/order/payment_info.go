package order

import (
	"fmt"
	"payment-go/internal/config"
	"payment-go/internal/models"
	"time"
)

type IOrderPaymentInfoDto interface {
	Map() map[string]any
}

type orderPaymentInfoWithLink struct {
	orderNumber string
	amount      float64
	timestamp   int64
}
type orderPaymentInfoWithCardNumber struct {
	cardNumber string
	amount     float64
	timestamp  int64
}

func NewPaymentInfo(ord *models.Order) (IOrderPaymentInfoDto, error) {
	if ord.HaveLink() {
		if ord.Status != models.StatusPending {
			return nil, fmt.Errorf("payment information is unavailable")
		}

		return &orderPaymentInfoWithLink{
			orderNumber: ord.Number.String(),
			amount:      ord.Amount,
			timestamp:   ord.CreatedAt.Unix(),
		}, nil
	} else if ord.PaymentMethod == config.PaymentMethodBankTransfer {
		if time.Since(ord.CreatedAt) >= config.CreateLinkTimeout+config.CheckLinkTimeout {
			return nil, fmt.Errorf("payment information is unavailable")
		}

		cardNumber := ord.Card.GetCardNumber()
		if cardNumber == nil || len(*cardNumber) == 0 {
			return nil, fmt.Errorf("card information is unavailable")
		}

		return &orderPaymentInfoWithCardNumber{
			cardNumber: *cardNumber,
			amount:     ord.Amount,
			timestamp:  ord.CreatedAt.Unix(),
		}, nil
	}

	return nil, fmt.Errorf("unknown payment method")
}

func (dto *orderPaymentInfoWithLink) Map() map[string]any {
	return map[string]any{
		"wait_for_link":   true,
		"amount":          dto.amount,
		"order_timestamp": dto.timestamp,
	}
}

func (dto *orderPaymentInfoWithCardNumber) Map() map[string]any {
	return map[string]any{
		"wait_for_link":   false,
		"amount":          dto.amount,
		"card_number":     dto.cardNumber,
		"order_timestamp": dto.timestamp,
	}
}
