package services

import (
	"fmt"
	"payment-go/internal/config"
	"payment-go/internal/services/payment_method"
	"payment-go/internal/transport/model/order"
	"sync"
)

type IPaymentService interface {
	ValidateOrderBeforeCreate(dto *order.CreateOrderDto) error
}
type paymentService struct {
}

var payIns *paymentService
var payOnce = sync.Once{}

func PaymentService() IPaymentService {
	payOnce.Do(func() {
		payIns = &paymentService{}
	})
	return payIns
}

func (s *paymentService) ValidateOrderBeforeCreate(dto *order.CreateOrderDto) error {
	if dto.PaymentMethod == config.PaymentMethodBankTransfer {
		return payment_method.ValidateMethodBankTransfer(dto)
	} else if dto.PaymentMethod == config.PaymentMethodKapitalBank {
		return payment_method.ValidateMethodKapitalBank(dto)
	}
	return fmt.Errorf("unknown payment method")
}
