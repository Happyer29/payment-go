package models

func GetModels() []interface{} {
	models := make([]interface{}, 0)
	models = append(
		models,
		&Order{},
		&PaymentLink{},
		&Card{},
		&Shop{},
		&CardInfo{},
		&BankMessage{},
		&Withdraw{},
	)
	return models
}
