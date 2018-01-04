package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils"
	"utils/logger"
	"servlets/token"
)

type transCommitParam struct {
	Txid string `json:"txid"`
}

type transCommitRequest struct {
	Base  *common.BaseInfo   `json:"base"`
	Param *transCommitParam `json:"param"`
}

// sendVCodeHandler
type transCommitHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *transCommitHandler) Method() string {
	return http.MethodPost
}

func (handler *transCommitHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
		Data: 0, // data expire Int 失效时间，单位秒
	}
	defer common.FlushJSONData2Client(response, writer)




	httpHeader := common.ParseHttpHeaderParams(request)

	// if httpHeader.IsValid() == false {
	if  !httpHeader.IsValidTimestamp() || !httpHeader.IsValidTokenhash()  {
		logger.Info("modify pwd: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		logger.Info("asset balance: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if len(aesKey) != constants.AES_totalLen {
		logger.Info("asset balance: get aeskey from cache error:", len(aesKey))
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}



	requestData := transCommitRequest{} // request body
	//header := common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &requestData)
	uid := utils.Str2Int64(uidString)


	//
	txid := utils.Str2Int64(requestData.Param.Txid)
	//获取原pending
	pending := common.FindPending(txid)
	pending.Status = constants.TX_STATUS_COMMIT
	//修改原pending 并返回修改之前的值 如果status 是默认值0 继续  不是就停止
	perPending := common.FindAndModify(txid,pending)
	if perPending.Id == txid && perPending.Status == constants.TX_STATUS_DEFAULT {
		//校验token对应的uid 和panding 中的from uid 是否一致
		if uid != perPending.From {
			response.SetResponseBase(constants.RC_INVALID_TOKEN)
			return
		}
		//判断to是否存在
		if !common.ExistsUID(perPending.To) {
			response.SetResponseBase(constants.RC_INVALID_OBJECT_ACCOUNT)
			return
		}

		if !common.TransAccountLvt(perPending.From,perPending.To,perPending.Value) {
			response.SetResponseBase(constants.RC_INSUFFICIENT_BALANCE)
			return
		}
	}

}
