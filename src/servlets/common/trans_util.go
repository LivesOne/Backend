package common

import (
	"servlets/constants"
	"utils"
	"utils/logger"
	"utils/config"
)


const (
	TRANS_TIMEOUT = 10 * 1000
)

func PrepareLVTTrans(from,to int64,txTpye int,value string)(string,constants.Error){
	txid := GenerateTxID()

	if txid == -1 {
		logger.Error("txid is -1  ")
		return "",constants.RC_SYSTEM_ERR
	}
	txh := DTTXHistory{
		Id:     txid,
		Status: constants.TX_STATUS_DEFAULT,
		Type:   txTpye,
		From:   from,
		To:     to,
		Value:  utils.FloatStrToLVTint(value),
		Ts:     utils.TXIDToTimeStamp13(txid),
		Code:   constants.TX_CODE_SUCC,
	}
	err := InsertPending(&txh)
	if err != nil {
		logger.Error("insert mongo db error ", err.Error())
		return "",constants.RC_SYSTEM_ERR
	}
	return utils.Int642Str(txid),constants.RC_OK
}



func CommitLVTTrans(uidStr,txIdStr string)constants.Error{
	txid := utils.Str2Int64(txIdStr)
	uid := utils.Str2Int64(uidStr)
	perPending, flag := FindAndModifyPending(txid, uid, constants.TX_STATUS_COMMIT)
	//未查到数据，返回处理中
	if !flag || perPending.Status != constants.TX_STATUS_DEFAULT {
		return constants.RC_TRANS_IN_PROGRESS

	}
	// 只有转账进行限制
	if perPending.Type == constants.TX_TYPE_TRANS {
		//非系统账号才进行限额校验
		if !config.GetConfig().CautionMoneyIdsExist(perPending.To) {
			level := GetTransLevel(perPending.From)
			//交易次数校验不通过，删除pending
			if f, e := CheckCommitLimit(perPending.From, level); !f {
				DeletePendingByInfo(perPending)
				return e
			}
		}
	}
	//txid 时间戳检测
	ts := utils.GetTimestamp13()
	txid_ts := utils.TXIDToTimeStamp13(txid)
	//暂时写死10秒
	if ts-txid_ts > TRANS_TIMEOUT {
		//删除pending
		DeletePendingByInfo(perPending)
		return constants.RC_TRANS_TIMEOUT

	}
	//存在就检测资产初始化状况，未初始化的用户给初始化
	CheckAndInitAsset(perPending.To)

	f, c := TransAccountLvt(txid, perPending.From, perPending.To, perPending.Value)
	if f {
		//成功 插入commited
		err := InsertCommited(perPending)
		if CheckDup(err) {
			//删除pending
			DeletePendingByInfo(perPending)
			//不删除数据库中的txid

			if perPending.Type == constants.TX_TYPE_TRANS {
				//common.RemoveTXID(txid)
				if !config.GetConfig().CautionMoneyIdsExist(perPending.To) {
					SetTotalTransfer(perPending.From, perPending.Value)
				}

			}
		}

	} else {
		//删除pending
		DeletePendingByInfo(perPending)
		//失败设置返回信息
		switch c {
		case constants.TRANS_ERR_INSUFFICIENT_BALANCE:
			return constants.RC_INSUFFICIENT_BALANCE
		case constants.TRANS_ERR_SYS:
			return constants.RC_TRANS_IN_PROGRESS
		case constants.TRANS_ERR_ASSET_LIMITED:
			return constants.RC_ACCOUNT_ACCESS_LIMITED
		}
	}
	return constants.RC_OK
}