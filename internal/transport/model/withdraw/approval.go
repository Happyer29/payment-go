package withdraw

import "encoding/json"

type WithdrawApprovalDto struct {
	WithdrawID uint `json:"withdraw_id"`
}

func ApprovalDtoFromJSON(data []byte) (*WithdrawApprovalDto, error) {
	var dto *WithdrawApprovalDto
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, err
	}
	return dto, nil
}
