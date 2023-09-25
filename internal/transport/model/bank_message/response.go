package bank_message

import "payment-go/internal/models"

type BankMessageResponseDto struct {
	ID            uint    `json:"id,omitempty"`
	OrderID       uint    `json:"order_id"`
	ShopID        uint    `json:"shop_id"`
	CreatedAt     int64   `json:"created_at"`
	SenderPhone   *string `json:"sender_phone,omitempty"`
	ReceiverPhone *string `json:"receiver_phone,omitempty"`
	RawMessage    string  `json:"raw_message,omitempty"`
	CardNumber    string  `json:"card_number,omitempty"`
	Amount        float64 `json:"amount,omitempty"`
	Error         string  `json:"error,omitempty"`
	Status        string  `json:"status,omitempty"`
}

func FromBankMessage(msg *models.BankMessage) *BankMessageResponseDto {
	return &BankMessageResponseDto{
		ID:            msg.ID,
		OrderID:       msg.OrderID,
		ShopID:        msg.ShopID,
		CreatedAt:     msg.CreatedAt.Unix(),
		SenderPhone:   msg.SenderPhone,
		ReceiverPhone: msg.ReceiverPhone,
		RawMessage:    msg.RawMessage,
		CardNumber:    msg.CardNumber,
		Amount:        msg.Amount,
		Error:         msg.Error,
		Status:        msg.Status,
	}
}

func FromBankMessages(messages []*models.BankMessage) []*BankMessageResponseDto {
	var ls = make([]*BankMessageResponseDto, len(messages))
	for i, msg := range messages {
		ls[i] = FromBankMessage(msg)
	}
	return ls
}
