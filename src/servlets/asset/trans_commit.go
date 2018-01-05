package asset

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils"
	"time"
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
		Data: 0,
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := transCommitRequest{} // request body

	common.ParseHttpBodyParams(request, &requestData)


	//
	txid := utils.Str2Int64(requestData.Param.Txid)
	//修改原pending 并返回修改之前的值 如果status 是默认值0 继续  不是就停止
	perPending := common.FindAndModifyPending(txid,constants.TX_STATUS_COMMIT)
	//未查到数据，返回处理中
	if perPending.Id != txid {
		response.SetResponseBase(constants.RC_TRANS_IN_PROGRESS)
		return
	}


	//TODO txid 时间戳检测

	ts := utils.GetTimestamp13()
	txid_ts := utils.TXIDToTimeStamp13(txid)


	//暂时写死10秒
	if ts - txid_ts > 10*1000{
		//删除pending
		common.DeletePending(txid)
		response.SetResponseBase(constants.RC_TRANS_TIMEOUT)
		return

	}


	//查到数据 检测状态是否为不为1
	if perPending.Status != constants.TX_STATUS_COMMIT {
		//判断to是否存在
		if common.ExistsUID(perPending.To) {
			//存在就检测资产初始化状况，未初始化的用户给初始化
			common.CheckAndInitAsset(perPending.To)
		}else{
			response.SetResponseBase(constants.RC_INVALID_OBJECT_ACCOUNT)
			return
		}

		if common.TransAccountLvt(txid,perPending.From,perPending.To,perPending.Value) {
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
			response.SetResponseBase(constants.RC_INSUFFICIENT_BALANCE)
		}
	}

}
