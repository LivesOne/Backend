package accounts

import (
	"encoding/json"
	"servlets/constants"
	"utils/vcode"
	"utils"
)

func DecryptSecret(secret string, key string, iv string, instance interface{}) constants.Error {
	dataStr, err := utils.AesDecrypt(secret, key, iv)
	if err != nil {
		return constants.RC_AES_DECRYPT
	}
	if err := json.Unmarshal([]byte(dataStr), instance); err != nil {
		return constants.RC_JSON_UNMARSHAL
	}
	return constants.RC_OK
}

func TokenErr2RcErr(tokenErr int) constants.Error {
	switch tokenErr {
	case constants.ERR_INT_OK:
		return constants.RC_OK
	case constants.ERR_INT_TK_DB:
		return constants.RC_TOKEN_DB
	case constants.ERR_INT_TK_DUPLICATE:
		return constants.RC_TOKEN_DUPLICATE
	case constants.ERR_INT_TK_NOTEXISTS:
		return constants.RC_TOKEN_NOTEXISTS
	default:
		return constants.RC_SYSTEM_ERR
	}
}

func ValidateMailVCodeErr2RcErr(validateErr int) constants.Error {
	switch validateErr {
	case vcode.SUCCESS:
		return constants.RC_OK
	case vcode.NOT_FOUND_ERR:
		return constants.RC_MAIL_VCODE_NOT_FOUND_ERR
	case vcode.SERVER_ERR:
		return constants.RC_MAIL_VCODE_SERVER_ERR
	case vcode.NO_PARAMS_ERR:
		return constants.RC_MAIL_VCODE_NO_PARAMS_ERR
	case vcode.PARAMS_ERR:
		return constants.RC_MAIL_VCODE_PARAMS_ERR
	case vcode.JSON_PARSE_ERR:
		return constants.RC_MAIL_VCODE_JSON_PARSE_ERR
	case vcode.CODE_EXPIRED_ERR:
		return constants.RC_MAIL_VCODE_CODE_EXPIRED_ERR
	case vcode.VALIDATE_CODE_FAILD:
		return constants.RC_MAIL_VCODE_VALIDATE_CODE_FAILD
	case vcode.EMAIL_VALIDATE_FAILD:
		return constants.RC_MAIL_VCODE_EMAIL_VALIDATE_FAILD
	case vcode.HTTP_ERR:
		return constants.RC_MAIL_VCODE_HTTP_ERR
	default:
		return constants.RC_MAIL_VCODE_UNKOWN_ERR
	}
}

