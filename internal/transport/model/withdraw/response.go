package withdraw

import (
	"payment-go/internal/models"
	"payment-go/internal/transport/model/shop"
)

type ResponseDto struct {
	ID                 uint                  `json:"id"`
	Number             string                `json:"number"`
	ShopID             uint                  `json:"shop_id"`
	Shop               *shop.ShopResponseDto `json:"shop"`
	Type               string                `json:"type"`
	CardNumber         string                `json:"card_number"`
	CardExpirationDate string                `json:"card_expiration_date"`
	Phone              string                `json:"phone"`
	Amount             float64               `json:"amount"`
	Status             string                `json:"status"`
	CreatedAt          int64                 `json:"created_at"`
	UpdatedAt          int64                 `json:"updated_at"`
	FinishedAt         *uint                 `json:"finished_at"`
}

func FromWithdraw(wd *models.Withdraw) *ResponseDto {
	return &ResponseDto{
		ID:                 wd.ID,
		Number:             wd.Number.String(),
		ShopID:             wd.ShopID,
		Shop:               shop.FromShop(&wd.Shop),
		Type:               wd.Type,
		CardNumber:         wd.CardNumber,
		CardExpirationDate: wd.CardExpirationDate,
		Phone:              wd.Phone,
		Amount:             wd.Amount,
		Status:             wd.Status,
		CreatedAt:          wd.CreatedAt.Unix(),
		UpdatedAt:          wd.UpdatedAt.Unix(),
		FinishedAt:         wd.FinishedAt,
	}
}

func FromWithdraws(wds []*models.Withdraw) []*ResponseDto {
	l := len(wds)
	res := make([]*ResponseDto, l)
	for i, wd := range wds {
		res[i] = FromWithdraw(wd)
	}
	return res
}
