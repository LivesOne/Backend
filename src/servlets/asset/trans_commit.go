package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
	"utils/config"
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


	if requestData.Param == nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}


	httpHeader := common.ParseHttpHeaderParams(request)

	// if httpHeader.IsValid() == false {
	if !httpHeader.IsValidTimestamp() || !httpHeader.IsValidTokenhash() {
		logger.Info("asset trans commited: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidStr, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		logger.Info("asset trans commited: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if len(aesKey) != constants.AES_totalLen {
		logger.Info("asset trans commited: get aeskey from cache error:", len(aesKey))
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}
	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]


	txIdStr, err := utils.AesDecrypt(requestData.Param.Txid, key, iv)
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
	if !flag || perPending.Status != constants.TX_STATUS_DEFAULT {
		response.SetResponseBase(constants.RC_TRANS_IN_PROGRESS)
		return
	}

	// 只有转账进行限制
	if perPending.Type == constants.TX_TYPE_TRANS {
		//非系统账号才进行限额校验
		if !config.GetConfig().CautionMoneyIdsExist(perPending.To) {
			level := common.GetTransLevel(perPending.From)
			//交易次数校验不通过，删除pending
			if f,e := common.CheckCommitLimit(perPending.From,level);!f {
				common.DeletePendingByInfo(perPending)
				response.SetResponseBase(e)
				return
			}
		}
	}





	//txid 时间戳检测

	ts := utils.GetTimestamp13()
	txid_ts := utils.TXIDToTimeStamp13(txid)

	//暂时写死10秒
	if ts-txid_ts > TRANS_TIMEOUT {
		//删除pending
		common.DeletePendingByInfo(perPending)
		response.SetResponseBase(constants.RC_TRANS_TIMEOUT)
		return

	}

	//查到数据 检测状态是否为不为1
	//if perPending.Status != constants.TX_STATUS_COMMIT {

	//在准备阶段判断to是否存在，不存在的交易 数据不入mongo
	//判断to是否存在
	//if common.ExistsUID(perPending.To) {

	//存在就检测资产初始化状况，未初始化的用户给初始化
	common.CheckAndInitAsset(perPending.To)

	//} else {
	//	response.SetResponseBase(constants.RC_INVALID_OBJECT_ACCOUNT)
	//	return
	//}
	f,c := common.TransAccountLvt(txid, perPending.From, perPending.To, perPending.Value)
	if f {
		//成功 插入commited
		err := common.InsertCommited(perPending)
		if common.CheckDup(err) {
			//删除pending
			common.DeletePendingByInfo(perPending)
			//不删除数据库中的txid

			if perPending.Type == constants.TX_TYPE_TRANS {
				//common.RemoveTXID(txid)
				if !config.GetConfig().CautionMoneyIdsExist(perPending.To) {
					common.SetTotalTransfer(perPending.From,perPending.Value)
				}

			}
		}

	} else {
		//删除pending
		common.DeletePendingByInfo(perPending)
		//失败设置返回信息
		switch c {
		case constants.TRANS_ERR_INSUFFICIENT_BALANCE:
			response.SetResponseBase(constants.RC_INSUFFICIENT_BALANCE)
		case constants.TRANS_ERR_SYS:
			response.SetResponseBase(constants.RC_TRANS_IN_PROGRESS)
		case constants.TRANS_ERR_ASSET_LIMITED:
			response.SetResponseBase(constants.RC_ACCOUNT_ACCESS_LIMITED)
		}
	}
	//}

}
