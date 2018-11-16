package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/db_factory"
	"utils/logger"
)

type bindWXSecret struct {
	Code    string `json:"code"`
	AppType string `json:"app_type"`
}

type bindWXParam struct {
	Secret string `json:"secret"`
}

type bindWXRequest struct {
	Param bindWXParam `json:"param"`
}

type bindWXResData struct {
	Awarded bool `json:"awarded"`
}

// bindWXHandler
type bindWXHandler struct {
}

func (handler *bindWXHandler) Method() string {
	return http.MethodPost
}

func (handler *bindWXHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true)
	defer log.InfoAll()
	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	header := common.ParseHttpHeaderParams(request)

	// fmt.Println("modify user profile: 111 \n", utils.ToJSONIndent(header), utils.ToJSONIndent(requestData))
	// request params check
	if !header.IsValid() {
		log.Info("modify user profile: invalid request param")
		response.SetResponseBase(constants.RC_PARAM_ERR)
	}
	// fmt.Println("modify user profile: 22222222 \n", utils.ToJSONIndent(header), utils.ToJSONIndent(requestData))

	// 判断用户身份
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

	uid := utils.Str2Int64(uidString)

	requestData := new(bindWXRequest)
	common.ParseHttpBodyParams(request, requestData)

	if len(aesKey) != constants.AES_totalLen {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		log.Info("bind phone: read aes key from db error, length of aes key is:", len(aesKey))
		return
	}

	// 解码 secret 参数
	secretString := requestData.Param.Secret

	secret := new(bindWXSecret)
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]
	if err := DecryptSecret(secretString, key, iv, secret); err != constants.RC_OK {
		response.SetResponseBase(err)
		log.Info("bind wx: Decrypt Secret error:", err)
		return
	}

	if ok, res := common.AuthWX(secret.AppType, secret.Code); ok {

		err := common.InitAccountExtend(uid)
		if err != nil {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
			return
		}

		r, err := common.SetWxId(uid, res.Openid, res.Unionid)
		if err != nil {
			if db_factory.CheckDuplicateByColumn(err, "wx_openid") ||
				db_factory.CheckDuplicateByColumn(err, "wx_unionid") {
				response.SetResponseBase(constants.RC_DUP_WX_ID)
				return
			} else {
				response.SetResponseBase(constants.RC_SYSTEM_ERR)
				return
			}
		} else {
			//r ==0 没有有效记录被修改
			if r == 0 {
				response.SetResponseBase(constants.RC_DUP_BIND_WX)
				return
			} else {
				//绑定微信成功，加算力,内部识别，加不加，加多少
				response.Data = bindWXResData{
					Awarded: common.AddBindActiveHashRateByWX(uid),
				}
			}
		}
	} else {
		response.SetResponseBase(constants.RC_INVALID_WX_CODE)
		return
	}

}
