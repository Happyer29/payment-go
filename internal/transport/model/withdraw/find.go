package withdraw

import (
	"encoding/json"
	"payment-go/internal/transport/model/shared"
)

type FindWithdrawDto struct {
	Search             string                       `json:"search,omitempty"`
	Sort               *shared.Sorting              `json:"sort,omitempty"`
	ID                 []uint                       `json:"id,omitempty"`
	Number             string                       `json:"number,omitempty"`
	ShopID             []uint                       `json:"shop_id,omitempty"`
	OwnerID            []uint                       `json:"owner_id,omitempty"`
	Type               []string                     `json:"type,omitempty"`
	CardNumber         string                       `json:"card_number,omitempty"`
	CardExpirationDate string                       `json:"card_expiration_date,omitempty"`
	Phone              string                       `json:"phone,omitempty"`
	Amount             *shared.RangeFilter[float64] `json:"amount,omitempty"`
	Status             []string                     `json:"status,omitempty"`
}

func BuildFindDto(data []byte) (*FindWithdrawDto, error) {
	var dto = &FindWithdrawDto{}
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, err
	}
	return dto, nil
}
