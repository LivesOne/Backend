package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
	"utils"
	"utils/logger"
)


type lvtc2bsvResData struct {
	Lvtc string `json:"lvtc"`
	Bsv  string `json:"bsv"`
}
type lvtc2bsvRequestParams struct {
	Secret    string `json:"secret"`
}

type lvtc2bsvRequest struct {
	Param *lvtc2bsvRequestParams `json:"param"`
}

type lvtc2bsvRequestSecret struct {
	LvtcNum  string `json:"lvtc_num"`
	BsvNum    string `json:"bsv_num"`
}

// sendVCodeHandler
type lvtc2bsvHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *lvtc2bsvHandler) Method() string {
	return http.MethodPost
}
func (wqs *lvtc2bsvRequestSecret) isValids() bool {
	return len(wqs.LvtcNum) > 0  && len(wqs.BsvNum) > 0
}
func (handler *lvtc2bsvHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true)
	defer log.InfoAll()
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)


	httpHeader := common.ParseHttpHeaderParams(request)

	// if httpHeader.IsValid() == false {
	if !httpHeader.IsValidTimestamp() || !httpHeader.IsValidTokenhash() {
		log.Info("asset trans prepare: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := rpc.GetTokenInfo(httpHeader.TokenHash)
	if err := rpc.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("asset trans prepare: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if len(aesKey) != constants.AES_totalLen {
		log.Info("asset trans prepare: get aeskey from cache error:", len(aesKey))
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	uid := utils.Str2Int64(uidString)

	requestData := lvtc2bsvRequest{} // request body

	parseFlag := common.ParseHttpBodyParams(request, &requestData)
	if !parseFlag {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]
	secrets := new(lvtc2bsvRequestSecret)
	if err := utils.DecodeSecret(requestData.Param.Secret, key, iv, secrets); err != nil {
		logger.Info("secret decode error", err.Error())
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if !secrets.isValids() {
		logger.Info("lvtc2bsv secret valid failed")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	//初始化
	common.	CheckAndInitAsset(uid)

	if lvtc,bsv,e := common.Lvtc2Bsv(uid,secrets.LvtcNum, secrets.BsvNum);e == constants.RC_OK {
		response.Data = &lvtc2bsvResData{
			Lvtc:  utils.LVTintToFloatStr(lvtc),
			Bsv: utils.LVTintToFloatStr(bsv),
		}
	} else {
		response.SetResponseBase(e)
	}
}
