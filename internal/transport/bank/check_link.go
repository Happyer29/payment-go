package bank

type CheckLinkResponseMerchant struct {
	Id string `json:"id"`
}
type CheckLinkResponseDto struct {
	TransactionId string                     `json:"id"`
	Amount        float64                    `json:"amount"`
	CardType      string                     `json:"cardType"`
	Merchant      *CheckLinkResponseMerchant `json:"merchant"`
	Status        int                        `json:"status"`
}

func (dto *CheckLinkResponseDto) Success() bool {
	if dto.TransactionId == "" {
		return false
	}
	if dto.Merchant.Id == "" {
		return false
	}
	return true
}
