package bank_message

import (
	"encoding/json"
)

type BankMessageApprovalDto struct {
	MessageID uint    `json:"message_id"`
	Amount    float64 `json:"amount,omitempty"`
}

func ApprovalDtoFromJSON(data []byte) (*BankMessageApprovalDto, error) {
	var dto *BankMessageApprovalDto
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, err
	}
	return dto, nil
}
