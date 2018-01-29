package constants

const (
	TX_TYPE_ALL               = 0
	TX_TYPE_REWARD            = 1
	TX_TYPE_PRIVATE_PLACEMENT = 2
	TX_TYPE_ACTIVITY_REWARD   = 3
	TX_TYPE_TRANS             = 4
	TX_TYPE_RECEIVABLES       = 5
	TX_STATUS_DEFAULT         = 0
	TX_STATUS_COMMIT          = 1

	ASSET_STATUS_DEF = 0

	ASSET_STATUS_LIMITED = 1

	TX_CODE_SUCC = 0

	AUTH_TYPE_LOGIN_PWD   = 1
	AUTH_TYPE_PAYMENT_PWD = 2

	ASSET_STATUS_INIT = 1

	TRANS_ERR_SUCC                 = 0
	TRANS_ERR_SYS                  = 1
	TRANS_ERR_INSUFFICIENT_BALANCE = 2
	TRANS_ERR_ASSET_LIMITED        = 3
)
