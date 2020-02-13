package common

type (
	DtTransferFee struct {
		FeeCurrency string `json:"fee_currency"`
		FeeRate     string `json:"fee_rate"`
		Discount    string `json:"discount"`
		FeeMin      string `json:"fee_min"`
		FeeMax      string `json:"fee_max"`
	}
	DtTransferAmount struct {
		Currency        string  `json:"currency"`
		SingleAmountMin float64 `json:"single_amount_min"`
		DailyAmountMax  float64 `json:"daily_amount_max"`
		UpdateTime      int64   `json:"update_time"`
	}
	DtWithdrawalFee struct {
		FeeCurrency string `json:"fee_currency"`
		FeeType     int    `json:"fee_type"`
		FeeFixed    string `json:"fee_fixed"`
		FeeRate     string `json:"fee_rate"`
		FeeMin      string `json:"fee_min"`
		FeeMax      string `json:"fee_max"`
		Discount    string `json:"discount"`
	}
)
