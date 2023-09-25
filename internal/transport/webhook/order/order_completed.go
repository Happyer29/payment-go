package order

type WebhookOrderCompletedDto struct {
	Success bool `json:"success"`
	//OrderId uint `json:"order_id"`
	OrderNumber string  `json:"order_number"`
	Amount      float64 `json:"amount"`
}
