// this file defines the http return code

package constants

type Error struct {
	Rc  int
	Msg string
}

var (
	RC_OK           = Error{0, "ok"}
	RC_SYSTEM_ERR   = Error{1, "system error"}
	RC_PROTOCOL_ERR = Error{2, "protocol error"}
	RC_PARAM_ERR    = Error{3, "param error"}
	RC_TOO_MANY_REQ = Error{4, "too many request"}
	RC_IP_LIMITED   = Error{5, "ip limited"}

	RC_INVALID_APPID    = Error{10001, "invalid appid"}
	RC_INVALID_PUB_KEY  = Error{10002, "invalid public key"}
	RC_INVALID_SIGN     = Error{10003, "invalid signature"}
	RC_INVALID_TOKEN    = Error{10004, "invalid token"}
	RC_INVALID_VCODE    = Error{10005, "invalid verification code"}
	RC_VCODE_EXPIRE     = Error{10006, "verification code expire"}
	RC_PUBLIC_EXPIRE    = Error{10007, "server public key expire"}
	RC_VCODE_NOT_EXISTS = Error{10008, "verification code not exists"}

	RC_DUP_EMAIL            = Error{20001, "duplicate email"}
	RC_DUP_PHONE            = Error{20002, "duplicate phone"}
	RC_DUP_NICKNAME         = Error{20003, "duplicate nickname"}
	RC_INVALID_ACCOUNT      = Error{20004, "invalid account"}
	RC_INVALID_LOGIN_PWD    = Error{20005, "invalid login password"}
	RC_INVALID_PAYMENT_PWD  = Error{20006, "invalid payment password"}
	RC_EMAIL_NOT_MATCH      = Error{20007, "email not match"}
	RC_PHONE_NOT_MATCH      = Error{20008, "phone not match"}
	RC_ACCOUNT_LIMITED      = Error{20009, "account limited"}
	RC_ACCOUNT_TEMP_LIMITED = Error{20010, "account temporary limited"}
	RC_UPGRAD_FAILED        = Error{20011, "upgrade failed"}
	RC_USER_LEVEL_LIMIT     = Error{20012, "level limit"}
	RC_INVALID_WX_CODE      = Error{20013, "invalid code"}
	RC_DUP_WX_ID            = Error{20014, "duplicate weixin id"}
	RC_DUP_BIND_WX          = Error{20015, "duplicate bind weixin"}

	RC_ACCOUNT_ACCESS_LIMITED       = Error{21001, "account access limited"}
	RC_INSUFFICIENT_BALANCE         = Error{21002, "insufficient balance"}
	RC_INVALID_OBJECT_ACCOUNT       = Error{21003, "invalid object account"}
	RC_TRANS_AUTH_FAILED            = Error{21004, "transaction authentication failed"}
	RC_TRANS_IN_PROGRESS            = Error{21005, "transaction in progress"}
	RC_TRANS_TIMEOUT                = Error{21006, "transaction timeout"}
	RC_INVALID_TXID                 = Error{21007, "invalid txid"}
	RC_TRANS_AMOUNT_TOO_LITTLE      = Error{21008, "transaction amount too little"}
	RC_TRANS_AMOUNT_EXCEEDING_LIMIT = Error{21009, "transaction amount exceeding limit"}

	// RC_AES_DECRYPT                     = Error{20101, "invalid params for AES decrypt"}
	// RC_JSON_UNMARSHAL                  = Error{20102, "invalid params for json unmarshal"}
	// RC_TOKEN_DB                        = Error{20103, "token db not connect"}
	// RC_TOKEN_DUPLICATE                 = Error{20104, "token duplicate"}
	// RC_TOKEN_NOTEXISTS                 = Error{20105, "token not exists"}
	// RC_DUP_LOGIN_PWD                   = Error{20106, "duplicate login password"}
	// RC_DUP_PAYMENT_PWD                 = Error{20107, "duplicate payment password"}
	// RC_MAIL_VCODE_NOT_FOUND_ERR        = Error{20108, "mail vcode not found err"}
	// RC_MAIL_VCODE_SERVER_ERR           = Error{20109, "mail vcode server err"}
	// RC_MAIL_VCODE_NO_PARAMS_ERR        = Error{20110, "mail vcode no params err"}
	// RC_MAIL_VCODE_PARAMS_ERR           = Error{20111, "mail vcode params err"}
	// RC_MAIL_VCODE_JSON_PARSE_ERR       = Error{20112, "mail vcode json parse err"}
	// RC_MAIL_VCODE_CODE_EXPIRED_ERR     = Error{20113, "mail vcode code expired err"}
	// RC_MAIL_VCODE_VALIDATE_CODE_FAILD  = Error{20114, "mail vcode validate code failed"}
	// RC_MAIL_VCODE_EMAIL_VALIDATE_FAILD = Error{20115, "mail vcode email validate failed"}
	// RC_MAIL_VCODE_HTTP_ERR             = Error{20116, "mail vcode http err"}
	// RC_MAIL_VCODE_UNKOWN_ERR           = Error{20117, "mail vcode unkown err"}
	// RC_ACCOUNT_NOT_EXIST               = Error{20118, "account not exist"}
	// RC_INVALIDE_EMAIL_ADDRESS          = Error{20119, "invalid email address"}
	// RC_INVALIDE_PHONE_NUM              = Error{20120, "invalid phone number"}
)

// HTTP return code constants
/*
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
*/

const (
	ERR_INT_OK           = 0 //internal errors
	ERR_INT_TK_DB        = -1
	ERR_INT_TK_DUPLICATE = -2
	ERR_INT_TK_NOTEXISTS = -3
)

const (
	ERR_EXT_OK = 0 //external errors
)
