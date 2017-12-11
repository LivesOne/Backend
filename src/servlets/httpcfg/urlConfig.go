// this file defines the http API URLS

package httpCfg

import (
	"strconv"
)

var (
	// HTTP_DOMAIN = "api.livesone.com" // http api domain name
	// HTTP_PORT   = ":9876"            // http port

	// HTTP_DOMAIN = "openapi.lives.one:9876" // http api domain name & port
	HTTP_DOMAIN string // http api domain name & port

	// hello world
	// http method : GET
	HELLO_WORLD string

	// Echo message
	// http method : POST
	ECHO_MSG string
)

// InitHTTPConfig initialize the http configuration
func InitHTTPConfig(domain string, port int) {

	HTTP_DOMAIN = domain + ":" + strconv.Itoa(port)

	HELLO_WORLD = HTTP_DOMAIN + "/demo/v1/hello"

	ECHO_MSG = HTTP_DOMAIN + "/demo/v1/echo"

	// fmt.Println(HELLO_WORLD, ECHO_MSG)
}

/*
const (
	// HTTP_DOMAIN = "api.livesone.com" // http api domain name
	// HTTP_PORT   = ":9876"            // http port

	// HTTP_DOMAIN = "openapi.lives.one:9876" // http api domain name & port
	HTTP_DOMAIN = "" // http api domain name & port

	// hello world
	// http method : GET
	HELLO_WORLD = HTTP_DOMAIN + "/demo/v1/hello"

	// Echo message
	// http method : POST
	ECHO_MSG = HTTP_DOMAIN + "/demo/v1/echo"

	// Account Register
	// http method : POST
	ACCOUNT_REGISTER = HTTP_DOMAIN + "/user/v1/account/register"

	// Account login
	// http method : POST
	ACCOUNT_LOGIN = HTTP_DOMAIN + "/user/v1/account/login"

	// Account auto login
	// http method : POST
	ACCOUNT_AUTOLOGIN = HTTP_DOMAIN + "/user/v1/account/autologin"

	// Account logout
	// http method : POST
	ACCOUNT_LOGOUT = HTTP_DOMAIN + "/user/v1/account/logout"

	// Get verification code
	// http method : POST
	ACCOUNT_GET_VCODE = HTTP_DOMAIN + "/user/v1/account/get_vcode"

	// Check verification code
	// http method : POST
	ACCOUNT_CHECK_VCODE = HTTP_DOMAIN + "/user/v1/account/check_vcode"

	// Modify login password
	// http method : POST
	ACCOUNT_MODIFY_PWD = HTTP_DOMAIN + "/user/v1/account/modify_pwd"

	// Reset login password
	// http method : POST
	ACCOUNT_RESET_PWD = HTTP_DOMAIN + "/user/v1/account/reset_pwd"

	// Set transaction password
	// http method : POST
	ACCOUNT_SET_TX_PWD = HTTP_DOMAIN + "/user/v1/account/set_tx_pwd"

	// Bind mobile phone
	// http method : POST
	ACCOUNT_BIND_PHONE = HTTP_DOMAIN + "/user/v1/account/bindphone"

	// Bind email
	// http method : POST
	ACCOUNT_BIND_EMAIL = HTTP_DOMAIN + "/user/v1/account/bindemail"

	// Get profile
	// http method : POST
	ACCOUNT_GET_PROFILE = HTTP_DOMAIN + "/user/v1/profile"

	// Profile modify
	// http method : POST
	ACCOUNT_MODIFY_PROFILE = HTTP_DOMAIN + "/user/v1/profile/modify"

	// Contacts Sync APIs

)
*/
