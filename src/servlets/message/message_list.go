package message

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/rpc"
	"utils"
	"utils/logger"
)

type (
	messageListHandler struct {
	}
	messageListResData struct {
		Message []common.DtMessage
	}
	messageListParam struct {
		Type int `json:"type"`
	}
	messageListReqData struct {
		Param *messageListParam `json:"param"`
	}
)

func (p *messageListReqData) IsValid() bool {
	return p.Param != nil
}

func (handler *messageListHandler) Method() string {
	return http.MethodPost
}

func (handler *messageListHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	log := logger.NewLvtLogger(true, "messageListHandler")
	defer log.InfoAll()

	res := common.NewResponseData()
	defer common.FlushJSONData2Client(res, writer)
	header := common.ParseHttpHeaderParams(request)
	if !header.IsValid() {
		log.Warn("header is not valid", utils.ToJSON(header))
		res.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidStr, aesKey, _, tokenErr := rpc.GetTokenInfo(header.TokenHash)
	if err := rpc.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		logger.Info("asset lockList: get info from cache error:", err)
		res.SetResponseBase(err)
		return
	}




	if len(aesKey) != constants.AES_totalLen {
		log.Info(" get aeskey from cache error:", len(aesKey))
		res.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if !utils.SignValid(aesKey, header.Signature, header.Timestamp) {
		log.Info("validate sign failed ")
		res.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	reqData := new(messageListReqData)

	if !common.ParseHttpBodyParams(request, reqData) {
		log.Info("decode json str error")
		res.SetResponseBase(constants.RC_PROTOCOL_ERR)
		return
	}
	if !reqData.IsValid() {
		log.Info("required param is nil")
		res.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	uid := utils.Str2Int64(uidStr)
	resData := new(messageListResData)
	msgArray := common.GetMsgByUidAndType(uid,reqData.Param.Type)
	if len(msgArray) > 0 {
		resData.Message = msgArray
	}
	res.Data = resData
}


