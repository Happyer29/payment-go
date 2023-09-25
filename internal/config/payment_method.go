package config

type PaymentMethodConfig struct {
	AmountMin map[string]float64
	AmountMax map[string]float64
}

func GetPaymentMethodConfig() *PaymentMethodConfig {
	return &PaymentMethodConfig{
		AmountMin: map[string]float64{
			PaymentMethodBankTransfer: 5,
			PaymentMethodKapitalBank:  1,
		},
		AmountMax: map[string]float64{
			PaymentMethodBankTransfer: 100_000,
			PaymentMethodKapitalBank:  100_000,
		},
	}
}
