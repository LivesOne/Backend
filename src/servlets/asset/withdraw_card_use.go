package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
)

type withdrawCardUseParam struct {
	Password string `json:"password"`
}

type withdrawCardUseRequest struct {
	Base  *common.BaseInfo      `json:"base"`
	Param *withdrawCardUseParam `json:"param"`
}

type withdrawCardUserResData struct {
	Quota string `json:"quota"`
}

// sendVCodeHandler
type withdrawCardUseHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *withdrawCardUseHandler) Method() string {
	return http.MethodPost
}

func (handler *withdrawCardUseHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true)
	defer log.InfoAll()
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := withdrawCardUseRequest{} // request body

	common.ParseHttpBodyParams(request, &requestData)

	if requestData.Param == nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	httpHeader := common.ParseHttpHeaderParams(request)

	// if httpHeader.IsValid() == false {
	if !httpHeader.IsValidTimestamp() || !httpHeader.IsValidTokenhash() {
		log.Info("asset trans commited: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidStr, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := common.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("asset trans commited: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if len(aesKey) != constants.AES_totalLen {
		log.Info("asset trans commited: get aeskey from cache error:", len(aesKey))
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]

	password, err := utils.AesDecrypt(requestData.Param.Password, key, iv)
	if err != nil {
		log.Error("aes decrypt error ", err.Error())
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	uid := utils.Str2Int64(uidStr)

	if ok, resErr := common.CheckUserCardLimit(uid); !ok {
		response.SetResponseBase(resErr)
		return
	}

	if card := common.GetUserWithdrawCardByPwd(password); card != nil {
		ts := utils.GetTimestamp13()
		if card.ExpireTime > 0 && card.ExpireTime < ts {
			response.SetResponseBase(constants.RC_USE_CARD_EXPIRE)
			return
		}
		if card.Status == constants.WITHDRAW_CARD_STATUS_USE {
			response.SetResponseBase(constants.RC_USE_CARD_ALREADY_USED)
			return
		}
		if err := common.UseWithdrawCard(card, uid); err != nil {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
			return
		}
		response.Data = withdrawCardUserResData{
			Quota: utils.LVTintToFloatStr(card.Quota),
		}
	} else {
		common.AddUseCardLimit(uid)
		response.SetResponseBase(constants.RC_USE_CARD_FAILED)
	}

}
