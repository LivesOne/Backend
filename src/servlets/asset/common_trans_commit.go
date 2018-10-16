package asset

import (
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"strings"
	"utils"
	"net/http"
	"utils/logger"
)

type commonTransCommitParam struct {
	Txid string `json:"txid"`
	Currency string `json:"currency"`
}

type commonTransCommitRequest struct {
	Base  *common.BaseInfo     `json:"base"`
	Param *commonTransCommitParam `json:"param"`
}

// sendVCodeHandler
type commonTransCommitHandler struct {
}

func (handler *commonTransCommitHandler) Method() string {
	return http.MethodPost
}

func (handler *commonTransCommitHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true)
	defer log.InfoAll()
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := commonTransCommitRequest{} // request body

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

	txid, err := utils.AesDecrypt(requestData.Param.Txid, key, iv)
	if err != nil {
		log.Error("aes decrypt error ", err.Error())
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	currency := strings.ToUpper(requestData.Param.Currency)
	switch currency {
	case common.CURRENCY_ETH:
		response.SetResponseBase(common.CommitETHTrans(uidStr, txid))
	case common.CURRENCY_LVT:
		response.SetResponseBase(common.CommitLVTTrans(uidStr, txid))
	case common.CURRENCY_LVTC:
		response.SetResponseBase(common.CommitLVTCTrans(uidStr, txid))
	default:
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
}
