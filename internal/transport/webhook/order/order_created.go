package order

type WebhookOrderCreatedDto struct {
	//OrderId     uint   `json:"order_id"`
	OrderNumber string `json:"order_number"`
	PaymentLink string `json:"payment_link"`
	Success     bool   `json:"success"`
}
