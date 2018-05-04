package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
	"utils/db_factory"
)

type bindTGSecret struct {
	Code string `json:"code"`
}

type bindTGParam struct {
	Secret string `json:"secret"`
}

type bindTGRequest struct {
	Param bindTGParam `json:"param"`
}
type bindTGResData struct {
	Awarded bool `json:"awarded"`
}

// bindTGHandler
type bindTGHandler struct {
}

func (handler *bindTGHandler) Method() string {
	return http.MethodPost
}

func (handler *bindTGHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(false)
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


	requestData := new(bindTGRequest)
	common.ParseHttpBodyParams(request, requestData)


	if len(aesKey) != constants.AES_totalLen {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		log.Info("bind phone: read aes key from db error, length of aes key is:", len(aesKey))
		return
	}

	// 解码 secret 参数
	secretString := requestData.Param.Secret

	secret := new(bindTGSecret)
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]
	if err := DecryptSecret(secretString, key, iv, secret); err != constants.RC_OK {
		response.SetResponseBase(err)
		log.Info("bind TG: Decrypt Secret error:", err)
		return
	}

	if ok,res := common.AuthTG(uidString,secret.Code);ok {

		err := common.InitAccountExtend(uid)
		if err != nil {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
			return
		}

		r, err := common.SetTGId(uid,res.Data.Telegram)
		if err != nil {
			if db_factory.CheckDuplicateByColumn(err,"tg_id"){
				response.SetResponseBase(constants.RC_DUP_TG_ID)
				return
			} else {
				response.SetResponseBase(constants.RC_SYSTEM_ERR)
				return
			}
		} else {
			//r ==0 没有有效记录被修改
			if r == 0 {
				response.SetResponseBase(constants.RC_DUP_BIND_TG)
				return
			} else {
				//绑定Telegram成功，加算力,内部识别，加不加，加多少
				response.Data = bindTGResData{
					Awarded: common.AddBindActiveHashRateByTG(uid),
				}
			}
		}
	} else {
		if res == nil {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
			return
		}

		response.SetResponseBase(constants.RC_INVALID_TG_CODE)
		return
	}


}
