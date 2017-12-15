package common

type (
	Account struct {
		ID              int64
		UID             int64
		UIDString       string
		Nickname        string
		Email           string
		Country         int
		Phone           string
		LoginPassword   string
		PaymentPassword string
		Language        string
		Region          string
		From            string
		RegisterTime    int64
		UpdateTime      int64
		RegisterType    int
	}

	Profile struct {
	}

	Contacts struct {
	}
)
