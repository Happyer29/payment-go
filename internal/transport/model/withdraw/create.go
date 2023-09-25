package withdraw

import (
	"encoding/json"
	"errors"
	"payment-go/internal/config"
	"payment-go/internal/models"
	"strings"
)

type CreateWithdrawDto struct {
	Auth               models.ShopKeys `json:"auth"`
	Type               string          `json:"type"`
	CardNumber         string          `json:"card_number,omitempty"`
	CardExpirationDate string          `json:"card_expiration_date,omitempty"`
	Phone              string          `json:"phone,omitempty"`
	Amount             float64         `json:"amount"`
}

var (
	ErrParseJson           = errors.New("invalid json")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrNoRecipientInfo     = errors.New("recipient information is empty")
	ErrUnknownWithdrawType = errors.New("unknown withdrawal method")
)

func ParseCreateDtoFromJSON(data []byte) (*CreateWithdrawDto, error) {
	var dto *CreateWithdrawDto
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, ErrParseJson
	}

	if dto.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	dto.Type = strings.ToLower(dto.Type)
	assertType := false
	for _, typ := range config.GetConfig().Withdraw.SupportedTypes {
		if typ == dto.Type {
			assertType = true
			break
		}
	}
	if !assertType {
		return nil, ErrUnknownWithdrawType
	}

	dto.Phone = strings.Trim(dto.Phone, " ")
	dto.CardNumber = strings.Trim(dto.CardNumber, " ")
	dto.CardExpirationDate = strings.Trim(dto.CardExpirationDate, " ")

	if (len(dto.CardNumber) == 0 || len(dto.CardExpirationDate) == 0) && len(dto.Phone) == 0 {
		return nil, ErrNoRecipientInfo
	}

	return dto, nil
}

func (dto *CreateWithdrawDto) GetCredentials() models.ShopKeys {
	return dto.Auth
}
