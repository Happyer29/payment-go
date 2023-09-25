package order

import (
	"fmt"
	"payment-go/internal/config"
	"payment-go/internal/models"
)

//type CreateOrderWebhook struct {
//	CallbackUrl string `json:"url"`
//	Key         string `json:"hash_key"`
//}

type CreateOrderDto struct {
	//ShopPublicKey          string          `json:"shop_public_key"`
	Auth                   models.ShopKeys `json:"auth"`
	Amount                 float64         `json:"amount"`
	PaymentMethod          string          `json:"payment_method"`
	Payload                string          `json:"payload"`
	LinkCreatedCallbackUrl string          `json:"link_callback_url"`
	PaymentCallbackUrl     string          `json:"payment_callback_url"`
}

func (dto *CreateOrderDto) Validate() error {
	if dto.Amount <= 0 {
		return fmt.Errorf("amount must be positive float")
	}
	//if len(dto.Webhook.LinkCreatedCallbackUrl) != 0 && len(dto.Webhook.Key) < 8 {
	//	return fmt.Errorf("hash_key is required for using")
	//}
	assertMethod := false
	for _, method := range config.GetConfig().Bank.PaymentMethods {
		if method == dto.PaymentMethod {
			assertMethod = true
			break
		}
	}
	if !assertMethod {
		return fmt.Errorf("unknown payment method")
	}

	return nil
}

func (dto *CreateOrderDto) GetCredentials() models.ShopKeys {
	return dto.Auth
}

//func (dto *CreateOrderDto) GetPublicKey() string {
//	return dto.ShopPublicKey
//}
