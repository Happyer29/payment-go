package payment_method

import (
	"fmt"
	"payment-go/internal/config"
	"payment-go/internal/transport/model/order"
)

func ValidateMethodBankTransfer(dto *order.CreateOrderDto) error {
	amountMin, ok := config.GetConfig().PaymentMethod.AmountMin[config.PaymentMethodBankTransfer]
	if ok && dto.Amount < amountMin {
		return fmt.Errorf("minimum amount for selected payment method is %.2f", amountMin)
	}
	amountMax, ok := config.GetConfig().PaymentMethod.AmountMax[config.PaymentMethodBankTransfer]
	if ok && dto.Amount > amountMax {
		return fmt.Errorf("maximum amount for selected payment method is %.2f", amountMax)
	}
	return nil
}
