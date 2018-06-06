package constants

const (
	TX_TYPE_ALL               = 0
	TX_TYPE_REWARD            = 1
	TX_TYPE_PRIVATE_PLACEMENT = 2
	TX_TYPE_ACTIVITY_REWARD   = 3
	TX_TYPE_TRANS             = 4
	TX_TYPE_RECEIVABLES       = 5
	TX_TYPE_PENALTY_MONEY     = 6
	TX_TYPE_BUY               = 7
	TX_TYPE_REFUND            = 8

	TX_TYPE_WITHDRAW        = 11
	TX_TYPE_WITHDRAW_RETURN = 12
	TX_TYPE_THREAD_IN       = 13
	TX_TYPE_THREAD_OUT      = 14

	TX_TYPE_RECHANGE      = 21
	TX_TYPE_WITHDRAW_FEE  = 22
	TX_TYPE_BUY_COIN_CARD = 23

	TX_STATUS_DEFAULT = 0
	TX_STATUS_COMMIT  = 1

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
	TRANS_ERR_PARAM                = 4

	ASSET_LOCK_MONTH_TIMESTAMP = 30 * 24 * 60 * 60 * 1000

	ASSET_LOCK_MAX_VALUE = 500

	TX_TYPE_WITHDRAW_ETH_FEE = 22
	TX_TYPE_WITHDRAW_LVT     = 11

	WITHDRAW_CARD_TYPE_DIV  = 0
	WITHDRAW_CARD_TYPE_FULL = 1

	USER_WITHDRAWAL_REQUEST_WAIT_SEND = 0
	USER_WITHDRAWAL_REQUEST_SEND      = 1
	USER_WITHDRAWAL_REQUEST_SUCCESS   = 2
	USER_WITHDRAWAL_REQUEST_FAIL      = 3
	USER_WITHDRAWAL_REQUEST_UNKNOWN   = 4

	WITHDRAW_CARD_STATUS_DEF = 1
	WITHDRAW_CARD_STATUS_USE = 2

	PUSH_TX_HISTORY_LVT_QUEUE_NAME = "TX_HIS_LVT_QUEUE"
)
