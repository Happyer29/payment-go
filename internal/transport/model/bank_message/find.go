package bank_message

import (
	"encoding/json"
	"payment-go/internal/transport/model/shared"
)

type FindBankMessageDto struct {
	Search     string                       `json:"search,omitempty"`
	Sort       *shared.Sorting              `json:"sort,omitempty"`
	ID         []uint                       `json:"id,omitempty"`
	OrderID    []uint                       `json:"order_id,omitempty"`
	ShopID     []uint                       `json:"shop_id,omitempty"`
	CardNumber []string                     `json:"card_number,omitempty"`
	Amount     *shared.RangeFilter[float64] `json:"amount,omitempty"`
	Status     []string                     `json:"status,omitempty"`
}

func BuildFindDto(data []byte) (*FindBankMessageDto, error) {
	var dto = &FindBankMessageDto{}
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, err
	}
	return dto, nil
}
