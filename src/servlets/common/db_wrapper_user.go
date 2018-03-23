package common

type (
	Account struct {
		ID              int64  `json:"_i,omitempty"`
		UID             int64  `json:"_u,omitempty"`
		UIDString       string `json:"uid"`
		Nickname        string `json:"nickname,omitempty"`
		Email           string `json:"email"`
		Country         int    `json:"country"`
		Phone           string `json:"phone"`
		LoginPassword   string `json:"_l,omitempty"`
		PaymentPassword string `json:"_p,omitempty"`
		Language        string `json:"language,omitempty"`
		Region          string `json:"region,omitempty"`
		From            string `json:"_f,omitempty"`
		RegisterTime    int64  `json:"register_time"`
		UpdateTime      int64  `json:"update_time"`
		RegisterType    int    `json:"_r,omitempty"`
		Level           int    `json:"level"`
		TraderLevel     int    `json:"trader_level"`
		Status          int    `json:"-"`
	}

	Profile struct {
	}

	Contacts struct {
	}
)
