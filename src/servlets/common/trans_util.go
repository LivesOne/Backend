package common

import (
	"database/sql"
	"servlets/constants"
	"strings"
	"utils"
	"utils/config"
	"utils/logger"
)

const (
	TRANS_TIMEOUT = 10 * 1000
)

func PrepareLVTTrans(from, to int64, txTpye int, value, bizContent string) (string, constants.Error) {
	tradeNo := GenerateTradeNo(constants.TRADE_TYPE_TRANSFER, constants.TX_TYPE_TRANS)
	txid := GenerateTxID()
	if txid == -1 {
		logger.Error("txid is -1  ")
		return "", constants.RC_SYSTEM_ERR
	}
	txh := DTTXHistory{
		Id:         txid,
		TradeNo:	tradeNo,
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

func CommitLVTTrans(uidStr, txIdStr string) (retErr constants.Error) {
	txid := utils.Str2Int64(txIdStr)
	uid := utils.Str2Int64(uidStr)
	perPending, flag := FindAndModifyPending(txid, uid, constants.TX_STATUS_COMMIT)
	//未查到数据，返回处理中
	if !flag || perPending.Status != constants.TX_STATUS_DEFAULT {
		return constants.RC_TRANS_IN_PROGRESS

	}
	// 只有转账进行限制
	var bizContent TransBizContent
	if perPending.Type == constants.TX_TYPE_TRANS {
		//非系统账号才进行限额校验
		e := VerifyLVTTrans(perPending.From, perPending.To,
			utils.LVTintToFloatStr(perPending.Value), false)
		if e != constants.RC_OK {
			DeletePendingByInfo(perPending)
			return e
		}
		if perPending.BizContent != "" {
			err := utils.FromJson(perPending.BizContent, &bizContent)
			if err != nil {
				logger.Info("dt_pending uid:", uidStr,
					" biz_content unmarshal to json failed,", err)
				return constants.RC_SYSTEM_ERR
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
	_, intErr := TransAccountLvt(tx, perPending)
	if intErr == constants.TRANS_ERR_SUCC {
		// 手续费转账流程
		if bizContent.Fee > 0 {
			feeTxid, feeTradeNo, feeSubType, transFeeAcc, retErr =
				TransFeeCommit(tx, perPending.From, bizContent.Fee, bizContent.FeeCurrency)
			if retErr != constants.RC_OK {
				return
			}
		}
	} else {
		//失败设置返回信息
		transInt2Error(intErr)
	}
	//删除pending
	DeletePendingByInfo(perPending)
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
			To: perPending.To, Amount: perPending.Value, Decimal: constants.TRADE_DECIMAIL,
			Currency: constants.TRADE_CURRENCY_LVT, CreateTime: perPending.Ts, FinishTime: finishTime,
		}
		tradesArray = append(tradesArray, trade)
		if feeTxid > 0 && len(feeTradeNo) > 0 {
			// 插入交易记录单：手续费
			trade.FeeTradeNo = feeTradeNo
			feeTrade := TradeInfo{
				TradeNo: feeTradeNo,OriginalTradeNo: perPending.TradeNo, Txid: feeTxid,
				Status: constants.TX_STATUS_COMMIT, Type: constants.TRADE_TYPE_FEE, SubType: feeSubType,
				From: perPending.From, To: transFeeAcc, Amount: bizContent.Fee, Decimal: constants.TRADE_DECIMAIL,
				Currency: bizContent.FeeCurrency, CreateTime: finishTime, FinishTime: finishTime,
			}
			tradesArray = append(tradesArray, feeTrade)
		}
		err = InsertTradeInfo(tradesArray...)
		if err != nil {
			logger.Error("insert mongo db:dt_trades error ", err.Error())
			return constants.RC_SYSTEM_ERR
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
	return constants.RC_OK
}


func CommitLVTCTrans(uidStr, txIdStr string) ( retErr constants.Error ) {
	txid := utils.Str2Int64(txIdStr)
	uid := utils.Str2Int64(uidStr)
	perPending, flag := FindAndModifyLVTCPending(txid, uid, constants.TX_STATUS_COMMIT)
	//未查到数据，返回处理中
	if !flag || perPending.Status != constants.TX_STATUS_DEFAULT {
		return constants.RC_TRANS_IN_PROGRESS
	}
	// 只有转账进行限制
	var bizContent TransBizContent
	if perPending.Type == constants.TX_TYPE_TRANS {
		//非系统账号才进行限额校验
		e := VerifyLVTCTrans(perPending.From, perPending.To,
			utils.LVTintToFloatStr(perPending.Value), false)
		if e != constants.RC_OK {
			DeletePendingByInfo(perPending)
			return e
		}
		if perPending.BizContent != "" {
			err := utils.FromJson(perPending.BizContent, &bizContent)
			if err != nil {
				logger.Info("dt_pending uid:", uidStr,
					" biz_content unmarshal to json failed,", err)
				return constants.RC_SYSTEM_ERR
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
	if intErr == constants.TRANS_ERR_SUCC {
		// 手续费转账流程
		if bizContent.Fee > 0 {
			feeTxid, feeTradeNo, feeSubType, transFeeAcc, retErr =
				TransFeeCommit(tx, perPending.From, bizContent.Fee, bizContent.FeeCurrency)
			if retErr != constants.RC_OK {
				return
			}
		}
	} else {
		//失败设置返回信息
		return transInt2Error(intErr)
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
			To: perPending.To, Amount: perPending.Value, Decimal: constants.TRADE_DECIMAIL,
			Currency: constants.TRADE_CURRENCY_LVTC, CreateTime: perPending.Ts, FinishTime: finishTime,
		}
		tradesArray = append(tradesArray, trade)
		if feeTxid > 0 && len(feeTradeNo) > 0 {
			// 插入交易记录单：手续费
			trade.FeeTradeNo = feeTradeNo
			feeTrade := TradeInfo{
				TradeNo: feeTradeNo,OriginalTradeNo: perPending.TradeNo, Txid: feeTxid,
				Status: constants.TX_STATUS_COMMIT, Type: constants.TRADE_TYPE_FEE, SubType: feeSubType,
				From: perPending.From, To: transFeeAcc, Amount: bizContent.Fee, Decimal: constants.TRADE_DECIMAIL,
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
func CommitETHTrans(uidStr, txidStr string) (retErr constants.Error) {

	uid := utils.Str2Int64(uidStr)

	tp, err := GetTradePendingByTxid(txidStr, uid)
	if err != nil {
		return constants.RC_SYSTEM_ERR
	}
	if tp == nil {
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

	var feeTxid, transFeeAcc int64
	var feeTradeNo string
	var feeSubType int
	txid := utils.Str2Int64(txidStr)
	_, intErr := EthTransCommit(txid, tp.From, to, tp.Value, tp.TradeNo, tp.Type, tx)
	if intErr == constants.TRANS_ERR_SUCC {
		// 手续费转账流程
		if bizContent.Fee > 0 {
			feeTxid, feeTradeNo, feeSubType, transFeeAcc, retErr =
				TransFeeCommit(tx, tp.From, bizContent.Fee, bizContent.FeeCurrency)
			if retErr != constants.RC_OK {
				return
			}
		}
	} else {
		//失败设置返回信息
		return transInt2Error(intErr)
	}
	err = DeleteTradePending(tp.TradeNo, uid, tx)
	if err != nil {
		return constants.RC_SYSTEM_ERR
	}

	//识别类型进行操作
	switch tp.Type {
	case constants.TX_TYPE_TRANS:
		var tradesArray []TradeInfo
		finishTime := utils.GetTimestamp13()
		// 插入交易记录单：转账
		trade := TradeInfo{
			TradeNo: tp.TradeNo, Txid: txid, Status: constants.TX_STATUS_COMMIT,
			Currency: constants.TRADE_CURRENCY_ETH, Type: constants.TRADE_TYPE_TRANSFER,
			SubType: tp.Type, From: tp.From, To: tp.To, Decimal: constants.TRADE_DECIMAIL,
			Amount: tp.Value, CreateTime: tp.Ts, FinishTime: finishTime,
		}
		tradesArray = append(tradesArray, trade)
		if feeTxid > 0 && len(feeTradeNo) > 0 {
			// 插入交易记录单：手续费
			trade.FeeTradeNo = feeTradeNo
			feeTrade := TradeInfo{
				TradeNo: feeTradeNo,OriginalTradeNo: tp.TradeNo, Txid: feeTxid,
				Status: constants.TX_STATUS_COMMIT, Type: constants.TRADE_TYPE_FEE, SubType: feeSubType,
				From: tp.From, To: transFeeAcc, Amount: bizContent.Fee, Decimal: constants.TRADE_DECIMAIL,
				Currency: bizContent.FeeCurrency, CreateTime: finishTime, FinishTime: finishTime,
			}
			tradesArray = append(tradesArray, feeTrade)
		}
		err = InsertTradeInfo(trade)
		if err != nil {
			logger.Error("insert mongo db:dt_trades error ", err.Error())
			return constants.RC_SYSTEM_ERR
		}
	}

	return constants.RC_OK
}

func TransFeeCommit(tx *sql.Tx,from, fee int64, currency string) (int64, string, int, int64, constants.Error) {
	// todo: lvtc/eth fee account
	feeSubType := constants.TX_SUB_TYPE_TRANSFER_FEE
	feeTradeNo := GenerateTradeNo(constants.TRADE_TYPE_FEE, feeSubType)
	feeTxid := GenerateTxID()
	transFeeAcc := config.GetConfig().LvtcTransFeeAccountUid
	currency = strings.ToUpper(currency)
	var err = constants.RC_OK
	var intErr = constants.TRANS_ERR_SYS
	feeDth := &DTTXHistory{
		Id: feeTxid, TradeNo: feeTradeNo, Status: constants.TX_STATUS_COMMIT,
		Type: feeSubType, From: from, To: transFeeAcc,
		Value: fee, Ts: utils.TXIDToTimeStamp13(feeTxid),
		Code: constants.TX_CODE_SUCC, BizContent: "",
	}
	switch currency {
	case CURRENCY_ETH:
		transFeeAcc = config.GetConfig().ETHTransFeeAccountUid
		_, intErr = EthTransCommit(feeTxid, from, transFeeAcc,
			fee, feeTradeNo, feeSubType, tx)
	case CURRENCY_LVT:
		err = VerifyLVTTrans(from, transFeeAcc, utils.Int642Str(fee), false)
		if err != constants.RC_OK {
			return 0, "", 0, 0, err
		}
		_, intErr = TransAccountLvt(tx, feeDth)
	case CURRENCY_LVTC:
		err = VerifyLVTCTrans(from, transFeeAcc, utils.Int642Str(fee), false)
		if err != constants.RC_OK {
			return 0, "", 0, 0, err
		}
		_, intErr = TransAccountLvtc(tx, feeDth)
	}
	if intErr != constants.TRANS_ERR_SUCC {
		err = transInt2Error(intErr)
	}
	return feeTxid, feeTradeNo, feeSubType, transFeeAcc, err
}

func transInt2Error(intErr int) constants.Error {
	switch intErr {
	case constants.TRANS_ERR_INSUFFICIENT_BALANCE:
		return constants.RC_INSUFFICIENT_BALANCE
	case constants.TRANS_ERR_SYS:
		return constants.RC_TRANS_IN_PROGRESS
	case constants.TRANS_ERR_ASSET_LIMITED:
		return constants.RC_ACCOUNT_ACCESS_LIMITED
	}
	return constants.RC_SYSTEM_ERR
}

func VerifyLVTTrans(from, to int64, valueStr string, prepare bool) constants.Error {
	// 非交易员禁止lvt转账
	level := GetTransLevel(from)
	if level == 0 {
		logger.Info("uid:", from, " not a trader, has no transfer permission.")
		return constants.RC_PERMISSION_DENIED
	}
	//目标账号非系统账号才校验额度
	if !config.GetConfig().CautionMoneyIdsExist(to) {
		if prepare {
			//在转账的情况下，目标为非系统账号，要校验目标用户是否有收款权限，交易员不受收款权限限制
			transLevelOfTo := GetTransLevel(to)
			if transLevelOfTo == 0 && ! CanBeTo(to) {
				logger.Info("asset trans prepare: target account has't receipt rights, to:", to)
				return constants.RC_INVALID_OBJECT_ACCOUNT
			}
			//金额校验
			if f, e := CheckAmount(from, utils.FloatStrToLVTint(valueStr), level); !f {
				logger.Info("asset trans prepare: transfer out amount level limit exceeded, from:", from)
				return e
			}
			//校验用户的交易限制
			if f, e := CheckPrepareLimit(from, level); !f {
				logger.Info("asset trans prepare: transfer out amount day limit exceeded, from:", from)
				return e
			}
		} else {
			if f, e := CheckCommitLimit(from, level); !f {
				return e
			}
		}
	}
	return constants.RC_OK
}

func VerifyLVTCTrans(from, to int64, valueStr string, prepare bool) constants.Error {
	//目标账号非系统账号才校验额度
	if !config.GetConfig().CautionMoneyIdsExist(to) {
		level := GetTransLevel(from)
		if prepare {
			//在转账的情况下，目标为非系统账号，要校验目标用户是否有收款权限，交易员不受收款权限限制
			transLevelOfTo := GetTransLevel(to)
			if transLevelOfTo == 0 && ! CanBeTo(to) {
				return constants.RC_INVALID_OBJECT_ACCOUNT
			}
			//金额校验
			if f, e := CheckAmount(from, utils.FloatStrToLVTint(valueStr), level); !f {
				return e
			}
			//校验用户的交易限制
			if f, e := CheckPrepareLimit(from, level); !f {
				return e
			}
		} else {
			if f, e := CheckCommitLimit(from, level); !f {
				return e
			}
		}
	}
	return constants.RC_OK
}
