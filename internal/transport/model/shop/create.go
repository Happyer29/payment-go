package shop

import (
	"fmt"
	"net/url"
)

type CreateShopDto struct {
	Name    string `json:"name"`
	Host    string `json:"host"`
	OwnerId uint   `json:"owner_id"`
}

func (dto *CreateShopDto) Validate() error {
	if len(dto.Name) < 3 {
		return fmt.Errorf("shop name must be at least 3 characters long")
	}
	if len(dto.Host) < 13 {
		return fmt.Errorf("shop host is too short")
	}
	_, err := url.Parse(dto.Host)
	if err != nil {
		return fmt.Errorf("invalid shop host")
	}
	if dto.OwnerId == 0 {
		return fmt.Errorf("please, specify the owner_id")
	}

	return nil
}
