// this file defines constants related with Accounts Feature

package constants

const (
	// 登录账号类型，  1 uid/2 email/3 phone
	LOGIN_TYPE_UID   = 1
	LOGIN_TYPE_EMAIL = 2
	LOGIN_TYPE_PHONE = 3
)

const (
	// right now, length of UID is 9
	LEN_uid = 9
)

const (
	// length of token-hash in HTTP request header
	LEN_HEADER_TOKEN_HASH = 64
	// length of signature in HTTP request header
	LEN_HEADER_SIGNATURE = 64
)

const (
	// AES encryption: iv length
	AES_ivLen = 16
	// AES encryption: key length
	AES_keyLen = 32

	// AES total length
	AES_totalLen = AES_ivLen + AES_keyLen
)
