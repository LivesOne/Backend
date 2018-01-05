package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
)

const (
	TRANS_TIMEOUT = 10 * 1000
)

type transCommitParam struct {
	Txid string `json:"txid"`
}

type transCommitRequest struct {
	Base  *common.BaseInfo  `json:"base"`
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
		Data: 0,
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := transCommitRequest{} // request body

	common.ParseHttpBodyParams(request, &requestData)

	httpHeader := common.ParseHttpHeaderParams(request)

	// if httpHeader.IsValid() == false {
	if !httpHeader.IsValidTimestamp() || !httpHeader.IsValidTokenhash() {
		logger.Info("modify pwd: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidStr, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
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
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]

	base64TxId := utils.Base64Decode(requestData.Param.Txid)

	txIdStr, err := utils.AesDecrypt(string(base64TxId), key, iv)
	if err != nil {
		logger.Error("aes decrypt error ", err.Error())
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	//获取解密后的txid
	txid := utils.Str2Int64(txIdStr)
	uid := utils.Str2Int64(uidStr)
	//修改原pending 并返回修改之前的值 如果status 是默认值0 继续  不是就停止
	perPending,flag := common.FindAndModifyPending(txid, uid, constants.TX_STATUS_COMMIT)
	//未查到数据，返回处理中
	if !flag {
		response.SetResponseBase(constants.RC_TRANS_IN_PROGRESS)
		return
	}

	//txid 时间戳检测

	ts := utils.GetTimestamp13()
	txid_ts := utils.TXIDToTimeStamp13(txid)

	//暂时写死10秒
	if ts-txid_ts > TRANS_TIMEOUT {
		//删除pending
		common.DeletePending(txid)
		response.SetResponseBase(constants.RC_TRANS_TIMEOUT)
		return

	}

	//查到数据 检测状态是否为不为1
	//if perPending.Status != constants.TX_STATUS_COMMIT {
	//判断to是否存在
	if common.ExistsUID(perPending.To) {
		//存在就检测资产初始化状况，未初始化的用户给初始化
		common.CheckAndInitAsset(perPending.To)
	} else {
		response.SetResponseBase(constants.RC_INVALID_OBJECT_ACCOUNT)
		return
	}
	f,c := common.TransAccountLvt(txid, perPending.From, perPending.To, perPending.Value)
	if f {
		//成功 插入commited
		common.InsertCommited(perPending)
		//删除pending
		common.DeletePending(txid)
		//删除数据库中txid
		common.RemoveTXID(txid)
	} else {
		//删除pending
		common.DeletePending(txid)
		//失败设置返回信息

		switch c {
		case constants.TRANS_ERR_INSUFFICIENT_BALANCE:
			response.SetResponseBase(constants.RC_INSUFFICIENT_BALANCE)
		case constants.TRANS_ERR_SYS:
			response.SetResponseBase(constants.RC_TRANS_IN_PROGRESS)
		}
	}
	//}

}
