package order

import (
	"encoding/json"
	"payment-go/internal/transport/model/shared"
)

type ReadOrderDto struct {
	Search  string                       `json:"search,omitempty"`
	Sort    *shared.Sorting              `json:"sort,omitempty"`
	ID      []uint                       `json:"id,omitempty"`
	OwnerID []uint                       `json:"owner_id,omitempty"`
	ShopID  []uint                       `json:"shop_id,omitempty"`
	CardID  []uint                       `json:"card_id,omitempty"`
	Amount  *shared.RangeFilter[float64] `json:"amount,omitempty"`
	Status  []string                     `json:"status,omitempty"`
}

func BuildFindDto(data []byte) (*ReadOrderDto, error) {
	var dto = &ReadOrderDto{}
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, err
	}
	return dto, nil
}
