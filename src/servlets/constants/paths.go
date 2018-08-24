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
	ACCOUNT_REGISTER         = "/user/v1/account/register"
	ACCOUNT_REGISTER_UPVCODE = "/user/v1/account/register_upvcode"
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

	ACCOUNT_CHECK_ACCOUNT = "/user/v1/account/check"

	CHECK_WITH_VCODE = "/user/v1/account/check_with_vcode"

	ACCOUNT_SET_STATUS = "/user/v1/account/set_status"

	ACCOUNT_UPGRADE = "/user/v1/profile/upgrade"

	ACCOUNT_BIND_WX = "/user/v1/profile/bind_wx"

	ACCOUNT_BIND_TG = "/user/v1/profile/bind_tg"

	ACCOUNT_PROFILE_USERINFO = "/user/v1/profile/userinfo"
	// Contacts Sync APIs

	//Assets Management APIs

	ASSET_REWARD = "/asset/v1/reward"

	ASSET_REWARD_EXTRACT = "/asset/v1/reward/extract"

	ASSET_REWARD_DETAIL = "/asset/v1/reward/detail"

	ASSET_BALANCE = "/asset/v1/balance"

	ASSET_TRANS_PREPARE = "/asset/v1/trans/prepare"

	ASSET_TRANS_COMMIT = "/asset/v1/trans/commit"



	ASSET_TRANS_RESULT = "/asset/v1/trans/result"

	ASSET_TRANS_HISTORY = "/asset/v1/trans/history"

	ASSET_LVTC_TRANS_PREPARE = "/asset/v1/trans/lvtc_prepare"

	ASSET_LVTC_TRANS_COMMIT = "/asset/v1/trans/lvtc_commit"

	ASSET_LVTC_TRANS_RESULT = "/asset/v1/trans/lvtc_result"

	ASSET_LVTC_TRANS_HISTORY = "/asset/v1/trans/lvtc_history"


	ASSET_ETH_TRANS_PREPARE = "/asset/v1/trans/eth_prepare"

	ASSET_ETH_TRANS_COMMIT = "/asset/v1/trans/eth_commit"

	ASSET_ETH_TRANS_RESULT = "/asset/v1/trans/eth_result"

	ASSET_ETH_TRANS_HISTORY = "/asset/v1/trans/eth_history"

	ASSET_LOGK_LIST = "/asset/v1/lock/list"

	ASSET_LOCK_CREATE = "/asset/v1/lock/create"

	ASSET_LOCK_REMOVE = "/asset/v1/lock/remove"

	ASSET_LOCK_UPGRADE = "/asset/v1/lock/upgrade"

	ASSET_WITHDRAWAL_QUOTA = "/asset/v1/withdraw/quota"

	ASSET_WITHDRAWAL_LIST = "/asset/v1/withdraw/list"

	ASSET_WITHDRAWAL_REQUEST = "/asset/v1/withdraw/request"

	ASSET_WITHDRAWAL_CARD_LIST = "/asset/v1/withdraw/card/list"

	ASSET_WITHDRAWAL_CARD_USE = "/asset/v1/withdraw/card/use"

	ASSET_WITHDRAWAL_CARD_USE_LIST = "/asset/v1/withdraw/card/use_list"

	ASSET_LVT2LVTC = "/asset/v1/misc/lvt2lvtc"

	ASSET_LVT2LVTC_DELAY = "/asset/v1/misc/lvt2lvtc_delay"

	ASSET_LVT2LVTC_COUNT = "/asset/v1/misc/lvt2lvtc_count"

	DEVICE_BIND_DEVICE = "/user/v1/device/bind"

	DEVICE_UNBIND_DEVICE = "/user/v1/device/unbind"

	DEVICE_DEVICE_INFO = "/user/v1/device/info"

	DEVICE_DEVICE_LIST = "/user/v1/device/list"

	DEVICE_FORCE_UNBIND = "/user/v1/device/force_unbind"
)
