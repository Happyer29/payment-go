package withdraw

import "encoding/json"

type UpdateWithdrawDto struct {
	Number string `json:"number"`
	Status string `json:"status"`
}

func ParseUpdateDtoFromJSON(data []byte) (*UpdateWithdrawDto, error) {
	var dto *UpdateWithdrawDto
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, err
	}

	return dto, nil
}
