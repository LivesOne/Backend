package common

const (
	REDIS_WITHDRAW_QUOTA_KEY = "withdraw:quota:"
)

type (
	WithdrawFee struct {
		Id          int64   `json:"id,omitempty"`
		Currency    string  `json:"currency,omitempty"`
		FeeCurrency string  `json:"fee_currency"`
		FeeType     int     `json:"fee_type"`
		FeeFixed    float64 `json:"fee_fixed"`
		FeeRate     float64 `json:"fee_rate"`
		FeeMin      float64 `json:"fee_min"`
		FeeMax      float64 `json:"fee_max"`
		Discount    float64 `json:"discount"`
		UpdateTime  int64   `json:"update_time"`
	}

	WithdrawQuota struct {
		Id              int64         `json:"id,omitempty"`
		Currency        string        `json:"currency,omitempty"`
		SingleAmountMin float64       `json:"single_amount_min,omitempty"`
		DailyAmountMax  float64       `json:"daily_amount_max,omitempty"`
		UpdateTime      int64         `json:"update_time,omitempty"`
		Fee             []WithdrawFee `json:"fee,omitempty"`
	}

	TransferFee struct {
		Id          int64   `json:"id,omitempty"`
		Currency    string  `json:"currency,omitempty"`
		FeeCurrency string  `json:"fee_currency"`
		FeeType     int     `json:"fee_type"`
		FeeFixed    float64 `json:"fee_fixed"`
		FeeRate     float64 `json:"fee_rate"`
		FeeMin      float64 `json:"fee_min"`
		FeeMax      float64 `json:"fee_max"`
		Discount    float64 `json:"discount"`
		UpdateTime  int64   `json:"update_time"`
	}

	TransferQuota struct {
		Id              int64       `json:"id,omitempty"`
		Currency        string      `json:"currency,omitempty"`
		SingleAmountMin float64     `json:"single_amount_min,omitempty"`
		DailyAmountMax  float64     `json:"daily_amount_max,omitempty"`
		UpdateTime      int64       `json:"update_time,omitempty"`
		Fee             TransferFee `json:"fee,omitempty"`
	}
)
