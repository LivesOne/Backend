package common

type (
	DtTransferFee struct {
		Currency    string  `json:"currency"`
		FeeCurrency string  `json:"fee_currency"`
		FeeRate     float64 `json:"fee_rate"`
		Discount    float64 `json:"discount"`
		FeeMin      float64 `json:"fee_min"`
		FeeMax      float64 `json:"fee_max"`
		UpdateTime  int64   `json:"update_time"`
	}
	DtTransferAmount struct {
		Currency        string  `json:"currency"`
		SingleAmountMin float64 `json:"single_amount_min"`
		DailyAmountMax  float64 `json:"daily_amount_max"`
		UpdateTime      int64   `json:"update_time"`
	}
)
