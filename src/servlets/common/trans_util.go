package common

import (
	"database/sql"
	"servlets/constants"
	"utils"
	"utils/config"
	"utils/logger"
)

const (
	TRANS_TIMEOUT = 10 * 1000
)

func PrepareLVTTrans(from, to int64, txTpye int, value, bizContent string) (string, constants.Error) {
	txid := GenerateTxID()

	if txid == -1 {
		logger.Error("txid is -1  ")
		return "", constants.RC_SYSTEM_ERR
	}
	txh := DTTXHistory{
		Id:         txid,
		Status:     constants.TX_STATUS_DEFAULT,
		Type:       txTpye,
		From:       from,
		To:         to,
		Value:      utils.FloatStrToLVTint(value),
		Ts:         utils.TXIDToTimeStamp13(txid),
		Code:       constants.TX_CODE_SUCC,
		BizContent: bizContent,
	}
	err := InsertPending(&txh)
	if err != nil {
		logger.Error("insert mongo db:dt_pending error ", err.Error())
		return "", constants.RC_SYSTEM_ERR
	}
	return utils.Int642Str(txid), constants.RC_OK
}
func PrepareLVTCTrans(from, to int64, txTpye int, value, bizContent string) (string, constants.Error) {
	tradeNo := GenerateTradeNo(constants.TRADE_TYPE_TRANSFER, constants.TX_TYPE_TRANS)
	txid := GenerateTxID()
	if txid == -1 {
		logger.Error("txid is -1  ")
		return "", constants.RC_SYSTEM_ERR
	}
	txh := &DTTXHistory{
		Id:         txid,
		TradeNo: tradeNo,
		Status:     constants.TX_STATUS_DEFAULT,
		Type:       txTpye,
		From:       from,
		To:         to,
		Value:      utils.FloatStrToLVTint(value),
		Ts:         utils.TXIDToTimeStamp13(txid),
		Code:       constants.TX_CODE_SUCC,
		BizContent: bizContent,
	}
	err := InsertLVTCPending(txh)
	if err != nil {
		logger.Error("insert mongo db:dt_pending error ", err.Error())
		return "", constants.RC_SYSTEM_ERR
	}
	return utils.Int642Str(txid), constants.RC_OK
}

func CommitLVTTrans(uidStr, txIdStr string) constants.Error {
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

			//识别类型进行操作
			switch perPending.Type {
			case constants.TX_TYPE_TRANS:
				if !config.GetConfig().CautionMoneyIdsExist(perPending.To) {
					SetTotalTransfer(perPending.From, perPending.Value)
				}
			case constants.TX_TYPE_BUY_COIN_CARD:
				var bizContent map[string]string
				utils.FromJson(perPending.BizContent, &bizContent)
				quota := utils.FloatStrToLVTint(bizContent["quota"])
				// 用卡记录
				wcu := &UserWithdrawalCardUse{
					TradeNo:    GenerateTradeNo(constants.TRADE_NO_BASE_TYPE, constants.TRADE_NO_TYPE_BUY_COIN_CARD),
					Uid:        uid,
					Quota:      quota,
					Cost:       perPending.Value,
					CreateTime: utils.TXIDToTimeStamp13(txid),
					Type:       constants.WITHDRAW_CARD_TYPE_DIV,
					Txid:       txid,
					Currency:   CURRENCY_LVT,
				}

				if err = InsertWithdrawalCardUse(wcu); err != nil {
					return constants.RC_SYSTEM_ERR
				}
				//临时额度
				if wr := InitUserWithdrawal(uid); wr != nil {
					if ok, _ := IncomeUserWithdrawalCasualQuota(uid, quota); !ok {
						return constants.RC_SYSTEM_ERR
					}
				} else {
					return constants.RC_SYSTEM_ERR
				}
			}
			tradeNo := GenerateTradeNo(constants.TRADE_TYPE_TRANSFER, constants.TX_TYPE_TRANS)
			trade := TradeInfo{
				TradeNo: tradeNo,
				Txid: perPending.Id,
				Status: perPending.Status,
				Type: constants.TRADE_TYPE_TRANSFER,
				From: perPending.From,
				To: perPending.To,
				Amount: perPending.Value,
				Decimal: 8,
				Currency: CURRENCY_LVT,
				CreateTime: perPending.Ts,
				FinishTime: utils.GetTimestamp13(),
			}
			err = InsertTradeInfo(trade)
			if err != nil {
				logger.Error("insert mongo db:dt_trades error ", err.Error())
				//return "", constants.RC_SYSTEM_ERR
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


func CommitLVTCTrans(uidStr, txIdStr, currency string) ( retErr constants.Error ) {
	txid := utils.Str2Int64(txIdStr)
	uid := utils.Str2Int64(uidStr)
	perPending, flag := FindAndModifyLVTCPending(txid, uid, constants.TX_STATUS_COMMIT)
	//未查到数据，返回处理中
	if !flag || perPending.Status != constants.TX_STATUS_DEFAULT {
		return constants.RC_TRANS_IN_PROGRESS
	}
	if currency != ""  && perPending.Currency != currency {
		return constants.RC_PARAM_ERR
	}
	// 只有转账进行限制
	if perPending.Type == constants.TX_TYPE_TRANS {
		//非系统账号才进行限额校验
		if !config.GetConfig().CautionMoneyIdsExist(perPending.To) {
			level := GetTransLevel(perPending.From)
			//交易次数校验不通过，删除pending
			if f, e := CheckCommitLimit(perPending.From, level); !f {
				DeleteLVTCPendingByInfo(perPending)
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
		DeleteLVTCPendingByInfo(perPending)
		return constants.RC_TRANS_TIMEOUT
	}
	//存在就检测资产初始化状况，未初始化的用户给初始化
	CheckAndInitAsset(perPending.To)

	var bizContent TransBizContent
	if perPending.Remark != "" {
		utils.FromJson(perPending.BizContent, &bizContent)
	}
	tx, err := gDBAsset.Begin()
	if err != nil {
		logger.Error("db pool begin error ", err.Error())
		return constants.RC_TRANS_IN_PROGRESS
	}
	defer func() {
		if retErr != constants.RC_OK {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	// 正常转账流程
	var feeTxid, transFeeAcc int64
	var feeTradeNo string
	var feeSubType int
	_, intErr := TransAccountLvtc(tx, perPending)
	if intErr == constants.TRANS_ERR_SUCC  {
		// 手续费转账流程
		if bizContent.Fee > 0 {
			feeTxid, feeTradeNo, feeSubType, transFeeAcc, intErr =
				TransFeeCommit(tx, perPending.From, bizContent.Fee, bizContent.FeeCurrency)
		}
	}
	if intErr != constants.TRANS_ERR_SUCC  {
		//删除dt_pending
		//DeleteLVTCPendingByInfo(perPending)
		//失败设置返回信息
		switch intErr {
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
	//删除pending
	DeleteLVTCPendingByInfo(perPending)
	//识别类型进行操作
	switch perPending.Type {
	case constants.TX_TYPE_TRANS:
		if !config.GetConfig().CautionMoneyIdsExist(perPending.To) {
			SetTotalTransfer(perPending.From, perPending.Value)
		}
		var tradesArray []TradeInfo
		finishTime := utils.GetTimestamp13()
		// 插入交易记录单：转账
		trade := TradeInfo{
			TradeNo: perPending.TradeNo, Txid: perPending.Id, Status: constants.TX_STATUS_COMMIT,
			Type: constants.TRADE_TYPE_TRANSFER, SubType: perPending.Type, From: perPending.From,
			To: perPending.To, Amount: perPending.Value, Decimal: 8,
			Currency: perPending.Currency, CreateTime: perPending.Ts, FinishTime: finishTime,
		}
		tradesArray = append(tradesArray, trade)
		if feeTxid > 0 && len(feeTradeNo) > 0 {
			// 插入交易记录单：手续费
			trade.FeeTradeNo = feeTradeNo
			feeTrade := TradeInfo{
				TradeNo: feeTradeNo,OriginalTradeNo: perPending.TradeNo, Txid: feeTxid,
				Status: constants.TX_STATUS_COMMIT, Type: constants.TRADE_TYPE_FEE, SubType: feeSubType,
				From: perPending.From, To: transFeeAcc, Amount: bizContent.Fee, Decimal: 8,
				Currency: bizContent.FeeCurrency, CreateTime: finishTime, FinishTime: finishTime,
			}
			tradesArray = append(tradesArray, feeTrade)
		}
		err = InsertTradeInfo(tradesArray...)
		if err != nil {
			logger.Error("insert mongo db:dt_trades error ", err.Error())
			return constants.RC_SYSTEM_ERR
		}
	}
	return constants.RC_OK
}

func PrepareETHTrans(from, to int64, valueStr string, txTpye int, bizContent string) (
	string, string, constants.Error) {
	tradeNo := GenerateTradeNo(constants.TRADE_TYPE_TRANSFER, txTpye)
	txid := GenerateTxID()
	if txid == -1 {
		logger.Error("txid is -1  ")
		return "", "", constants.RC_SYSTEM_ERR
	}
	value := utils.FloatStrToLVTint(valueStr)
	if err := InsertTradePending(txid, from, to, tradeNo, bizContent, value, txTpye); err != nil {
		logger.Error("insert trade pending error", err.Error())
		return "", "", constants.RC_SYSTEM_ERR
	}
	return utils.Int642Str(txid), tradeNo, constants.RC_OK
}
func CommitETHTrans(uidStr, txidStr, currency string ) (retErr constants.Error) {

	uid := utils.Str2Int64(uidStr)

	tp, err := GetTradePendingByTxid(txidStr, uid)
	if err != nil {
		return constants.RC_SYSTEM_ERR
	}
	if tp == nil || currency != CURRENCY_ETH {
		return constants.RC_PARAM_ERR
	}
	ts := utils.GetTimestamp13()
	//暂时写死10秒
	if ts-tp.Ts > TRANS_TIMEOUT {
		//删除pending
		DeleteTradePending(tp.TradeNo, uid, nil)
		return constants.RC_TRANS_TIMEOUT
	}

	var (
		to         int64
		bizContent TransBizContent
	)
	//识别类型进行操作
	switch tp.Type {
	//case constants.TX_TYPE_BUY_COIN_CARD:
	//	to = config.GetWithdrawalConfig().WithdrawalCardEthAcceptAccount // 手续费收款账号
	case constants.TX_TYPE_TRANS:
		to = tp.To
	default:
		return constants.RC_PARAM_ERR
	}

	//存在就检测资产初始化状况，未初始化的用户给初始化
	CheckAndInitAsset(to)

	if tp.BizContent != "" {
		//解析业务数据，拿到具体数值
		if je := utils.FromJson(tp.BizContent, &bizContent); je != nil {
			return constants.RC_PARAM_ERR
		}
	}
	tx, err := gDBAsset.Begin()
	if err != nil {
		return constants.RC_SYSTEM_ERR
	}
	defer func() {
		if retErr != constants.RC_OK {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	var feeTxid int64
	var feeTradeNo string
	var feeSubType = constants.TX_TYPE_TRANS
	transFeeAcc := config.GetConfig().PenaltyMoneyAccountUid

	txid := utils.Str2Int64(txidStr)
	_, intErr := EthTransCommit(txid, tp.From, to, tp.Value, tp.TradeNo, tp.Type, tx)
	if intErr == constants.TRANS_ERR_SUCC {
		// 手续费转账流程
		if bizContent.Fee > 0 {
			feeTxid, feeTradeNo, feeSubType, transFeeAcc, intErr =
				TransFeeCommit(tx, tp.From, bizContent.Fee, bizContent.FeeCurrency)
		}
	}
	if intErr != constants.TRANS_ERR_SUCC {
		//删除trade_pending
		//DeleteTradePending(tp.TradeNo, uid, tx)
		//失败设置返回信息
		switch intErr {
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
	err = DeleteTradePending(tp.TradeNo, uid, tx)
	if err != nil {
		return constants.RC_SYSTEM_ERR
	}
	var tradesArray []TradeInfo
	finishTime := utils.GetTimestamp13()
	// 插入交易记录单：转账
	trade := TradeInfo{
		TradeNo: tp.TradeNo, Txid: txid, Status: constants.TX_STATUS_COMMIT, Currency: currency,
		Type: constants.TRADE_TYPE_TRANSFER, SubType: tp.Type, From: tp.From, To: tp.To,
		Amount: tp.Value, Decimal: 8, CreateTime: tp.Ts, FinishTime: finishTime,
	}
	tradesArray = append(tradesArray, trade)
	if feeTxid > 0 && len(feeTradeNo) > 0 {
		// 插入交易记录单：手续费
		trade.FeeTradeNo = feeTradeNo
		feeTrade := TradeInfo{
			TradeNo: feeTradeNo,OriginalTradeNo: tp.TradeNo, Txid: feeTxid,
			Status: constants.TX_STATUS_COMMIT, Type: constants.TRADE_TYPE_FEE, SubType: feeSubType,
			From: tp.From, To: transFeeAcc, Amount: bizContent.Fee, Decimal: 8,
			Currency: bizContent.FeeCurrency, CreateTime: finishTime, FinishTime: finishTime,
		}
		tradesArray = append(tradesArray, feeTrade)
	}
	err = InsertTradeInfo(trade)
	if err != nil {
		logger.Error("insert mongo db:dt_trades error ", err.Error())
		return constants.RC_SYSTEM_ERR
	}
	return constants.RC_OK
}

func TransFeeCommit(tx *sql.Tx,from, fee int64, currency string) (int64, string, int, int64, int) {
	// 转账subType 待定
	feeSubType := constants.TX_TYPE_TRANS
	feeTradeNo := GenerateTradeNo(constants.TRADE_TYPE_FEE, feeSubType)
	feeTxid := GenerateTxID()
	transFeeAcc := config.GetConfig().LvtcTransFeeAccountUid
	var intErr = constants.TRANS_ERR_SYS
	switch currency {
	case CURRENCY_ETH:
		transFeeAcc = config.GetConfig().ETHTransFeeAccountUid
		_, intErr = EthTransCommit(feeTxid, from, transFeeAcc,
			fee, feeTradeNo, feeSubType, tx)
	case CURRENCY_LVTC:
		feeDth := &DTTXHistory{
			Id: feeTxid, TradeNo: feeTradeNo, Status: constants.TX_STATUS_COMMIT,
			Type: feeSubType, From: from, To: transFeeAcc,
			Value: fee, Ts: utils.TXIDToTimeStamp13(feeTxid),
			Code: constants.TX_CODE_SUCC, BizContent: "",
		}
		_, intErr = TransAccountLvtc(tx, feeDth)
	}
	return feeTxid, feeTradeNo, feeSubType, transFeeAcc, intErr
}
