package withdraw

import "payment-go/internal/models"

type WithdrawUpdateWebhookDto struct {
	Number     string  `json:"number"`
	Amount     float64 `json:"amount"`
	Status     string  `json:"status"`
	UpdatedAt  int64   `json:"updated_at"`
	FinishedAt *uint   `json:"finished_at"`
}

func FromWithdraw(wd *models.Withdraw) *WithdrawUpdateWebhookDto {
	return &WithdrawUpdateWebhookDto{
		Number:     wd.Number.String(),
		Amount:     wd.Amount,
		Status:     wd.Status,
		UpdatedAt:  wd.UpdatedAt.Unix(),
		FinishedAt: wd.FinishedAt,
	}
}
