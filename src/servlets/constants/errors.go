// this file defines the http return code

package constants

// HTTP return code constants
const (
	RC_OK                       = 0 // ok
	RC_SYSTEM_ERROR             = 1 // system error
	RC_PROTOCOL_ERR             = 2 // protocol error
	RC_TOO_MANY_REQUEST         = 3 // too many request
	RC_IP_LIMITED               = 4 // ip limited
	RC_READ_REQUEST_PARAM_ERROR = 5 // read http request params error

	RC_INVALID_APPID     = 10001 // invalid appid
	RC_INVALID_PUB_KEY   = 10002 // invalid public key
	RC_INVALID_SIGNATURE = 10003 // invalid signature
	RC_INVALID_TOKEN     = 10004 // invalid token
	RC_INVALID_VER_CODE  = 10005 // invalid verification code
	RC_VER_CODE_EXPIRE   = 10006 // verification code expire

	RC_DUP_EMAIL           = 20001 // duplicate email
	RC_DUP_PHONE           = 20002 // duplicate phone
	RC_DUP_NICKNAME        = 20003 // duplicate nickname
	RC_INVALID_ACCOUNT     = 20004 // invalid account
	RC_INVALID_LOGIN_PWD   = 20005 // invalid login password
	RC_INVALID_PAYMENT_PWD = 20006 // invalid payment password

)

const (
	ERR_INT_OK           = 0 //internal errors
	ERR_INT_TK_DB        = -1
	ERR_INT_TK_DUPLICATE = -2
	ERR_INT_TK_NOTEXISTS = -3
)

const (
	ERR_EXT_OK = 0 //external errors
)
