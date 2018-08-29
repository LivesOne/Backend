package common

type (
	DtTransferFee struct {
		Currency    string  `json:"currency"`
		FeeCurrency string  `json:"fee_currency"`
		FeeRate     float64 `json:"fee_rate"`
		Discount    float64 `json:"discount"`
		FeeMax      int     `json:"fee_max"`
		UpdateTime  int64   `json:"update_time"`
	}
)
