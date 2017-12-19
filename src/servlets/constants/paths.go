// this file defines the http API URLS

package constants

const (
	// hello world
	// http method : GET
	HelloWorld = "/helloworld"

	// Echo message
	// http method : POST
	Echo = "/demo/v1/echo"

	// Account Register
	// http method : POST
	ACCOUNT_REGISTER = "/user/v1/account/register"

	// Account login
	// http method : POST
	ACCOUNT_LOGIN = "/user/v1/account/login"

	// Account auto login
	// http method : POST
	ACCOUNT_AUTOLOGIN = "/user/v1/account/autologin"

	// Account logout
	// http method : POST
	ACCOUNT_LOGOUT = "/user/v1/account/logout"

	// 获取图像验证码（Get image verification code）
	// http method : POST
	ACCOUNT_GET_IMG_VCODE = "/user/v1/account/get_img_vcode"

	// 发送验证码（Send verification code）
	// http method : POST
	ACCOUNT_SEND_VCODE = "/user/v1/account/send_vcode"

	// Check verification code
	// http method : POST
	ACCOUNT_CHECK_VCODE = "/user/v1/account/check_vcode"

	// Modify login password
	// http method : POST
	ACCOUNT_MODIFY_PWD = "/user/v1/account/modify_pwd"

	// Reset login password
	// http method : POST
	ACCOUNT_RESET_PWD = "/user/v1/account/reset_pwd"

	// Set transaction password
	// http method : POST
	ACCOUNT_RESET_TX_PWD = "/user/v1/account/reset_tx_pwd"

	// Bind mobile phone
	// http method : POST
	ACCOUNT_BIND_PHONE = "/user/v1/account/bind_phone"

	// Bind email
	// http method : POST
	ACCOUNT_BIND_EMAIL = "/user/v1/account/bind_email"

	// Get profile
	// http method : POST
	ACCOUNT_GET_PROFILE = "/user/v1/profile"

	// Profile modify
	// http method : POST
	ACCOUNT_MODIFY_PROFILE = "/user/v1/profile/modify"

	// Contacts Sync APIs
)
