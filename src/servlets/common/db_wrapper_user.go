package common

type (
	Account struct {
		ID              int64  `json:"_i,omitempty"`
		UID             int64  `json:"_u,omitempty"`
		UIDString       string `json:"uid,omitempty"`
		Nickname        string `json:"nickname,omitempty"`
		Email           string `json:"email,omitempty"`
		Country         int    `json:"country,omitempty"`
		Phone           string `json:"phone,omitempty"`
		LoginPassword   string `json:"_l,omitempty"`
		PaymentPassword string `json:"_p,omitempty"`
		Language        string `json:"language,omitempty"`
		Region          string `json:"region,omitempty"`
		From            string `json:"_f,omitempty"`
		RegisterTime    int64  `json:"register_time,omitempty"`
		UpdateTime      int64  `json:"update_time,omitempty"`
		RegisterType    int    `json:"_r,omitempty"`
	}

	Profile struct {
	}

	Contacts struct {
	}

)
