package shop

import (
	"encoding/json"
	"payment-go/internal/transport/model/shared"
)

type FindShopDto struct {
	Search        string          `json:"search,omitempty"`
	Sort          *shared.Sorting `json:"sort,omitempty"`
	ID            []uint          `json:"id,omitempty"`
	OwnerID       []uint          `json:"owner_id,omitempty"`
	Active        *bool           `json:"active,omitempty"`
	HostValidated *bool           `json:"host_validated,omitempty"`
	Moderated     *bool           `json:"moderated,omitempty"`
}

func BuildFindDto(data []byte) (*FindShopDto, error) {
	var dto = &FindShopDto{}
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, err
	}
	return dto, nil
}
