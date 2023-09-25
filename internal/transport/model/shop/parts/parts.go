package parts

import "payment-go/internal/models"

type ShopWebhooksDto struct {
	LinkCreated       *string `json:"link_created,omitempty"`
	OnSuccess         *string `json:"on_success,omitempty"`
	OnFailure         *string `json:"on_failure,omitempty"`
	OnWithdrawUpdated *string `json:"on_withdraw_updated,omitempty"`
}

func (sw *ShopWebhooksDto) Resolve() models.ShopWebhooks {
	return models.ShopWebhooks{
		LinkCreated:       sw.LinkCreated,
		OnSuccess:         sw.OnSuccess,
		OnFailure:         sw.OnFailure,
		OnWithdrawUpdated: sw.OnWithdrawUpdated,
	}
}
