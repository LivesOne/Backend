package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
	"utils"
	"utils/logger"
)

const (
	QR_CODE_TYPE_TRANS = "transfer"
	QR_CODE_TYPE_UINFO = "user"
	QR_CODE_TRANS_TYPE_CHAIN = "chain"
	QR_CODE_TRANS_TYPE_LIVESONE = "livesone"
)

type transInfo struct {
	Type     string `json:"type"`
	To       string `json:"to"`
	Currency string `json:"currency"`
	Amount   string `json:"amount"`
}

type qrCodeGenerateParam struct {
	Type      string    `json:"type"`
	TransInfo *transInfo `json:"trans_info"`
}

type qrCodeGenerateRequest struct {
	Param *qrCodeGenerateParam `json:"param"`
}

type qrCodeGenerateResData struct {
	QrInfo     string `json:"qr_info"`
	ExpireTime int64  `json:"expire_time"`
}

// checkVCodeHandler
type qrCodeGenerateHandler struct {
}

func (handler *qrCodeGenerateHandler) Method() string {
	return http.MethodPost
}

func (handler *qrCodeGenerateHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true, "user qrCodeGenerate")
	defer log.InfoAll()
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	httpHeader := common.ParseHttpHeaderParams(request)

	if httpHeader.Timestamp < 1 {
		log.Error("timestamp check failed")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := rpc.GetTokenInfo(httpHeader.TokenHash)
	if err := rpc.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Error("get cache failed")
		response.SetResponseBase(err)
		return
	}

	log.Info("uid", uidString)

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		log.Error("validate sign failed")
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	requestData := new(qrCodeGenerateRequest)

	common.ParseHttpBodyParams(request, requestData)

	param := requestData.Param

	if param == nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	switch requestData.Param.Type {
	case QR_CODE_TYPE_TRANS:
		ti := param.TransInfo
		if err := validateTransInfo(ti);err != constants.RC_OK {
			response.SetResponseBase(err)
			return
		}
		response.Data = buildTransQrInfo(ti)
	case QR_CODE_TYPE_UINFO:
		response.Data = buildUserQrInfo(uidString)
	default:
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

}

func buildUserQrInfo(uid string)qrCodeGenerateResData{
	code,expire := common.BuildUserInfoQrCodeCache(uid)
	r := qrCodeGenerateResData{
		QrInfo:     common.BuildQrCodeContent(code),
		ExpireTime: int64(expire),
	}
	return r
}

func buildTransQrInfo(ti *transInfo)qrCodeGenerateResData{
	code,expire := common.BuildTransQrCodeCache(ti.Type,ti.To,ti.Currency,ti.Amount)
	r := qrCodeGenerateResData{
		QrInfo:     common.BuildQrCodeContent(code),
		ExpireTime: int64(expire),
	}
	return r
}

func validateTransInfo(ti *transInfo)constants.Error{

	switch ti.Type {
	case QR_CODE_TRANS_TYPE_CHAIN:
		if utils.ValidateWithdrawalAddress(ti.To,ti.Currency) {
			return constants.RC_OK
		}
	case QR_CODE_TRANS_TYPE_LIVESONE:
		if len(ti.To) == constants.LEN_uid {
			return constants.RC_OK
		}
	}
	return constants.RC_PARAM_ERR
}