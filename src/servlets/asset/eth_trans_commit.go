package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
)



type ethTransCommitParam struct {
	TradeNo string `json:"trade_no"`
}

type ethTransCommitRequest struct {
	Base  *common.BaseInfo  `json:"base"`
	Param *ethTransCommitParam `json:"param"`
}

// sendVCodeHandler
type ethTransCommitHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *ethTransCommitHandler) Method() string {
	return http.MethodPost
}

func (handler *ethTransCommitHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true)
	defer log.InfoAll()
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := ethTransCommitRequest{} // request body

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
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
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

	tradeNo, err := utils.AesDecrypt(requestData.Param.TradeNo, key, iv)
	if err != nil {
		log.Error("aes decrypt error ", err.Error())
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	//调用统一确认交易流程
	response.SetResponseBase(common.CommitETHTrans(uidStr,tradeNo))

}
