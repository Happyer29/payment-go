package card

import (
	"fmt"
	"strings"
)

type ChangeBalanceDto struct {
	CardID    uint    `json:"card_id"`
	Amount    float64 `json:"amount"`
	Operation string  `json:"operation"`
}

type changeBalanceRequest struct {
	cardID     uint
	amount     float64
	isIncrease bool
}

type IChangeBalanceRequest interface {
	IsIncrease() bool
	Amount() float64
	CardID() uint
}

func (r *ChangeBalanceDto) GetRequest() (IChangeBalanceRequest, error) {
	if r.CardID == 0 {
		return nil, fmt.Errorf("invalid card_id")
	}
	if r.Amount <= 0 {
		return nil, fmt.Errorf("amount must be positive float")
	}
	var res = &changeBalanceRequest{
		cardID: r.CardID,
		amount: r.Amount,
	}

	operation := strings.ToLower(r.Operation)
	if operation == "increase" {
		res.isIncrease = true
	} else if operation == "decrease" {
		res.isIncrease = false
	} else {
		return nil, fmt.Errorf("unknown operation")
	}

	return res, nil
}

func NewChangeBalanceRequest(cardID uint, amount float64, isIncrease bool) (IChangeBalanceRequest, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("error while creating changeBalanceRequest: amount must be positive")
	}

	return &changeBalanceRequest{
		cardID:     cardID,
		amount:     amount,
		isIncrease: isIncrease,
	}, nil
}

func (r *changeBalanceRequest) IsIncrease() bool {
	return r.isIncrease
}

func (r *changeBalanceRequest) Amount() float64 {
	return r.amount
}

func (r *changeBalanceRequest) CardID() uint {
	return r.cardID
}
