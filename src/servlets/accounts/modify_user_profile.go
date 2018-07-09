package accounts

import (
	"net/http"
	"regexp"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"strings"
	"utils"
	"utils/db_factory"
	"utils/logger"
)

type profileSecret struct {
	Nickname      string `json:"nickname"`
	WalletAddress string `json:"wallet_address"`
}

type modifyProfileParam struct {
	Secret string `json:"secret"`
}

type modifyProfileRequest struct {
	Param modifyProfileParam `json:"param"`
}

// modifyUserProfileHandler
type modifyUserProfileHandler struct {
}

func (handler *modifyUserProfileHandler) Method() string {
	return http.MethodPost
}

func (handler *modifyUserProfileHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true)
	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)
	defer log.InfoAll()
	requestData := modifyProfileRequest{}
	common.ParseHttpBodyParams(request, &requestData)
	header := common.ParseHttpHeaderParams(request)

	// fmt.Println("modify user profile: 111 \n", utils.ToJSONIndent(header), utils.ToJSONIndent(requestData))
	// request params check
	if !header.IsValid() || (len(requestData.Param.Secret) < 1) {
		log.Info("modify user profile: invalid request param")
		response.SetResponseBase(constants.RC_PARAM_ERR)
	}
	// fmt.Println("modify user profile: 22222222 \n", utils.ToJSONIndent(header), utils.ToJSONIndent(requestData))

	// 判断用户身份
	// _, aesKey, _, _ := token.GetAll(header.TokenHash)
	uidString, aesKey, _, tokenErr := token.GetAll(header.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		response.SetResponseBase(err)
		log.Info("modify user profile: read user info error:", err)
		return
	}

	if len(aesKey) != constants.AES_totalLen {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		log.Info("modify user profile: read aes key from db error, length of aes key is:", len(aesKey))
		return
	}
	if !utils.SignValid(aesKey, header.Signature, header.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}
	// fmt.Println("modify user profile: 333", aesKey)

	// 解码 secret 参数
	secret := new(profileSecret)
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]
	// fmt.Println("modify user profile: 444", iv, key)
	err := DecryptSecret(requestData.Param.Secret, key, iv, secret)
	// fmt.Println("modify user profile: 555", utils.ToJSONIndent(secret), err)
	if err != constants.RC_OK {
		response.SetResponseBase(err)
		log.Info("modify user profile: Decrypt Secret error:", err)
		return
	}

	uid := utils.Str2Int64(uidString)

	// if common.ExistsNickname(secret.Nickname) {
	// 	response.SetResponseBase(constants.RC_DUP_NICKNAME)
	// 	logger.Info("modify user profile: duplicate nickname:", secret.Nickname)
	// 	return
	// }

	if len(secret.Nickname) > 0 {
		if !validateNickName(secret.Nickname) {
			log.Error("validate nickname failed")
			response.SetResponseBase(constants.RC_INVALID_NICKNAME_FORMAT)
			return
		}
		dbErr := common.SetNickname(uid, secret.Nickname)
		if dbErr != nil {
			if db_factory.CheckDuplicateByColumn(dbErr, "nickname") {
				log.Info("modify user profile: duplicate nickname", dbErr)
				response.SetResponseBase(constants.RC_DUP_NICKNAME)
			} else {
				log.Info("modify user profile : save nickname to db error:", dbErr)
				response.SetResponseBase(constants.RC_SYSTEM_ERR)
			}
			// } else {
			// 	fmt.Println("modify user profile: success")
		}
	}
	if len(secret.WalletAddress) > 0 {
		walletAddress := strings.ToLower(secret.WalletAddress)
		if validateWalletAddress(walletAddress) {
			if !strings.HasPrefix(walletAddress, "0x") {
				walletAddress = "0x" + walletAddress
			}
			if common.CheckWalletAddressBlacklist(walletAddress) > 0 {
				logger.Info("blacklist exist this wallet address ", walletAddress)
				response.SetResponseBase(constants.RC_DUP_WALLET_ADDRESS)
			} else {
				rowsAffected, dbErr := common.SetWalletAddress(uid, walletAddress)
				if dbErr != nil {
					if rowsAffected == 0 || db_factory.CheckDuplicateByColumn(dbErr, "wallet_address") {
						log.Info("modify user profile: duplicate wallet_address", dbErr)
						response.SetResponseBase(constants.RC_DUP_WALLET_ADDRESS)
					} else {
						log.Info("modify user profile : save wallet_address to db error:", dbErr)
						response.SetResponseBase(constants.RC_SYSTEM_ERR)
					}
				}
			}
		} else {
			response.SetResponseBase(constants.RC_INVALID_WALLET_ADDRESS_FORMAT)
		}
	}

	log.Info("modify user profile success")

}

func validateNickName(name string) bool {
	l := len(name)
	if l < 4 || l > 30 {
		return false
	}
	reg := "^[-\u4e00-\u9fa5a-zA-Z0-9_]{2,30}$"
	ret, _ := regexp.MatchString(reg, name)
	return ret
}

func validateWalletAddress(walletAddress string) bool {
	reg := "^(0x)?[0-9a-f]{40}$"
	ret, _ := regexp.MatchString(reg, strings.ToLower(walletAddress))
	return ret
}
