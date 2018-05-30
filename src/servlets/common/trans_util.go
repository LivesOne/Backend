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
func PrepareETHTrans(from int64,valueStr string,txTpye int,bizContent map[string]string)(string,constants.Error){
	tradeNo := GenerateTradeNo(txTpye,txTpye)//TODO 修改
	value := utils.FloatStrToLVTint(valueStr)
	if err := InsertTradePending(from,tradeNo,utils.ToJSON(bizContent),value,txTpye);err != nil {
		logger.Error("insert trade pending error",err.Error())
		return "",constants.RC_SYSTEM_ERR
	}
	return tradeNo,constants.RC_OK
}
func CommitETHTrans(uidStr,tradeNo string)constants.Error{

	uid := utils.Str2Int64(uidStr)

	tp,err := GetTradePendingByTradeNo(tradeNo,uid)
	if err != nil {
		return constants.RC_SYSTEM_ERR
	}
	if tp == nil {
		return constants.RC_PARAM_ERR
	}
	var (
		to int64
		bizContent map[string]string
	)
	//识别类型进行操作
	switch tp.Type {
	case constants.TX_TYPE_BUY_COIN_CARD:
		to = config.GetWithdrawalConfig().WithdrawalCardEthAcceptAccount // 手续费收款账号
	default:
		return constants.RC_PARAM_ERR
	}
	//解析业务数据，拿到具体数值
	if je := utils.FromJson(tp.BizContent,&bizContent);je != nil {
		return constants.RC_PARAM_ERR
	}
	tx ,err := gDBAsset.Begin()
	if err != nil {
		return constants.RC_SYSTEM_ERR
	}
	txId,e := EthTransCommit(uid,to,tp.Value,tp.TradeNo,tp.Type,tx)
	if txId <= 0 {
		tx.Rollback()
		switch e {
		case constants.TRANS_ERR_INSUFFICIENT_BALANCE:
			return constants.RC_INSUFFICIENT_BALANCE
		case constants.TRANS_ERR_SYS:
			return constants.RC_TRANS_IN_PROGRESS
		case constants.TRANS_ERR_ASSET_LIMITED:
			return constants.RC_ACCOUNT_ACCESS_LIMITED
		default:
			return constants.RC_SYSTEM_ERR
		}

	}

	//识别类型进行操作
	switch tp.Type {
	case constants.TX_TYPE_BUY_COIN_CARD:
		quota := utils.FloatStrToLVTint(bizContent["quota"])
		// 用卡记录
		wcu := &UserWithdrawalCardUse{
			TradeNo:    tp.TradeNo,
			Uid:        uid,
			Quota:      quota,
			Cost:       tp.Value,
			CreateTime: utils.TXIDToTimeStamp13(txId),
		}

		if err = InsertWithdrawalCardUseByTx(wcu,tx);err != nil {
			tx.Rollback()
			return constants.RC_SYSTEM_ERR
		}
		//临时额度
		if wr := InitUserWithdrawalByTx(uid,tx);wr != nil {
			if ok,_ := IncomeUserWithdrawalCasualQuotaByTx(uid,quota,tx);!ok{
				tx.Rollback()
				return constants.RC_SYSTEM_ERR
			}
		} else {
			tx.Rollback()
			return constants.RC_SYSTEM_ERR
		}
	default:
		tx.Rollback()
		return constants.RC_PARAM_ERR
	}




	err = DeleteTradePending(tp.TradeNo,uid,tx)
	if err != nil {
		tx.Rollback()
		return constants.RC_SYSTEM_ERR
	}

	err = tx.Commit()
	if err != nil {
		return constants.RC_SYSTEM_ERR
	}
	return constants.RC_OK
}


