package common

import (
	"database/sql"
	"servlets/constants"
	"strconv"
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
		Currency:	CURRENCY_LVT,
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
		Currency:	CURRENCY_LVTC,
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
	perPending.Status = constants.TX_STATUS_COMMIT
	// 只有转账进行限制
	var bizContent TransBizContent
	if perPending.Type == constants.TX_TYPE_TRANS {
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
	CheckAndInitAsset(perPending.From)
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
		var tradesArray []TradeInfo
		finishTime := utils.GetTimestamp13()
		// 插入交易记录单：转账
		trade := TradeInfo{
			TradeNo: perPending.TradeNo, Txid: perPending.Id, Status: constants.TRADE_STATUS_SUCC,
			Type: constants.TRADE_TYPE_TRANSFER, SubType: perPending.Type, From: perPending.From,
			To: perPending.To, Amount: perPending.Value, Decimal: constants.TRADE_DECIMAIL,
			Currency: constants.TRADE_CURRENCY_LVT, CreateTime: perPending.Ts, FinishTime: finishTime,
		}
		if feeTxid > 0 && len(feeTradeNo) > 0 {
			// 插入交易记录单：手续费
			trade.FeeTradeNo = feeTradeNo
			feeTrade := TradeInfo{
				TradeNo: feeTradeNo,OriginalTradeNo: perPending.TradeNo, Txid: feeTxid,
				Status: constants.TRADE_STATUS_SUCC, Type: constants.TRADE_TYPE_FEE, SubType: feeSubType,
				From: perPending.From, To: transFeeAcc, Amount: bizContent.Fee, Decimal: constants.TRADE_DECIMAIL,
				Currency: bizContent.FeeCurrency, CreateTime: finishTime, FinishTime: finishTime,
			}
			tradesArray = append(tradesArray, feeTrade)
		}
		tradesArray = append(tradesArray, trade)
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
	perPending.Status = constants.TX_STATUS_COMMIT
	// 只有转账进行限制
	var bizContent TransBizContent
	if perPending.Type == constants.TX_TYPE_TRANS {
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
	CheckAndInitAsset(perPending.From)

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
		var tradesArray []TradeInfo
		finishTime := utils.GetTimestamp13()
		// 插入交易记录单：转账
		trade := TradeInfo{
			TradeNo: perPending.TradeNo, Txid: perPending.Id, Status: constants.TRADE_STATUS_SUCC,
			Type: constants.TRADE_TYPE_TRANSFER, SubType: perPending.Type, From: perPending.From,
			To: perPending.To, Amount: perPending.Value, Decimal: constants.TRADE_DECIMAIL,
			Currency: CURRENCY_LVTC, CreateTime: perPending.Ts, FinishTime: finishTime,
		}
		if feeTxid > 0 && len(feeTradeNo) > 0 {
			// 插入交易记录单：手续费
			trade.FeeTradeNo = feeTradeNo
			feeTrade := TradeInfo{
				TradeNo: feeTradeNo,OriginalTradeNo: perPending.TradeNo, Txid: feeTxid,
				Status: constants.TRADE_STATUS_SUCC, Type: constants.TRADE_TYPE_FEE, SubType: feeSubType,
				From: perPending.From, To: transFeeAcc, Amount: bizContent.Fee, Decimal: constants.TRADE_DECIMAIL,
				Currency: bizContent.FeeCurrency, CreateTime: finishTime, FinishTime: finishTime,
			}
			tradesArray = append(tradesArray, feeTrade)
		}
		tradesArray = append(tradesArray, trade)
		err = InsertTradeInfo(tradesArray...)
		if err != nil {
			logger.Error("insert mongo db:dt_trades error ", err.Error())
			return constants.RC_SYSTEM_ERR
		}
		// 设置日限额
		SetDailyTransAmount(uid, CURRENCY_LVTC, perPending.Value)
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
	CheckAndInitAsset(tp.From)
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
			TradeNo: tp.TradeNo, Txid: txid, Status: constants.TRADE_STATUS_SUCC,
			Currency: CURRENCY_ETH, Type: constants.TRADE_TYPE_TRANSFER,
			SubType: tp.Type, From: tp.From, To: tp.To, Decimal: constants.TRADE_DECIMAIL,
			Amount: tp.Value, CreateTime: tp.Ts, FinishTime: finishTime,
		}
		if feeTxid > 0 && len(feeTradeNo) > 0 {
			// 插入交易记录单：手续费
			trade.FeeTradeNo = feeTradeNo
			feeTrade := TradeInfo{
				TradeNo: feeTradeNo,OriginalTradeNo: tp.TradeNo, Txid: feeTxid,
				Status: constants.TRADE_STATUS_SUCC, Type: constants.TRADE_TYPE_FEE, SubType: feeSubType,
				From: tp.From, To: transFeeAcc, Amount: bizContent.Fee, Decimal: constants.TRADE_DECIMAIL,
				Currency: bizContent.FeeCurrency, CreateTime: finishTime, FinishTime: finishTime,
			}
			tradesArray = append(tradesArray, feeTrade)
		}
		tradesArray = append(tradesArray, trade)
		err = InsertTradeInfo(tradesArray...)
		if err != nil {
			logger.Error("insert mongo db:dt_trades error ", err.Error())
			return constants.RC_SYSTEM_ERR
		}
		// 设置日限额
		SetDailyTransAmount(uid, CURRENCY_ETH, tp.Value)
	}

	return constants.RC_OK
}

func TransFeeCommit(tx *sql.Tx,from, fee int64, currency string) (int64, string, int, int64, constants.Error) {
	// todo: lvtc/eth fee account
	feeSubType := constants.TX_SUB_TYPE_TRANSFER_FEE
	feeTradeNo := GenerateTradeNo(constants.TRADE_TYPE_FEE, feeSubType)
	feeTxid := GenerateTxID()
	transFeeAcc := config.GetConfig().TransFeeAccountUid
	currency = strings.ToUpper(currency)
	var err = constants.RC_OK
	var intErr = constants.TRANS_ERR_SYS
	feeDth := &DTTXHistory{
		Id: feeTxid, TradeNo: feeTradeNo, Status: constants.TX_STATUS_COMMIT,
		Type: feeSubType, From: from, To: transFeeAcc, Currency: currency,
		Value: fee, Ts: utils.TXIDToTimeStamp13(feeTxid),
		Code: constants.TX_CODE_SUCC, BizContent: "",
	}
	switch currency {
	case CURRENCY_ETH:
		_, intErr = EthTransCommit(feeTxid, from, transFeeAcc,
			fee, feeTradeNo, feeSubType, tx)
	case CURRENCY_LVT:
		_, intErr = TransAccountLvt(tx, feeDth)
	case CURRENCY_LVTC:
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

func VerifyLVTTrans(from int64) constants.Error {
	// 非交易员禁止lvt转账
	level := GetTransLevel(from)
	if level == 0 {
		logger.Info("uid:", from, " not a trader, has no transfer permission.")
		return constants.RC_PERMISSION_DENIED
	}
	return constants.RC_OK
}

func VerifyLVTCTrans(uid int64, valueStr string) constants.Error {
	value := utils.FloatStrToLVTint(valueStr)
	// 校验单笔最低限额
	if e := CheckSingleTransAmount(CURRENCY_LVTC, value); e != constants.RC_OK {
		return e
	}
	// 校验转账日限额
	if f, e := CheckDailyTransAmount(uid, CURRENCY_LVTC, value); !f {
		return e
	}
	return constants.RC_OK
}

func VerifyEthTrans(uid int64, valueStr string) constants.Error {
	value := utils.FloatStrToLVTint(valueStr)
	// 校验单笔最低限额
	if e := CheckSingleTransAmount(CURRENCY_ETH, value); e != constants.RC_OK {
		return e
	}
	// 校验转账日限额
	if f, e := CheckDailyTransAmount(uid, CURRENCY_ETH, value); !f {
		return e
	}
	return constants.RC_OK
}

func CheckTransFee(value, fee, currency, feeCurrency string) constants.Error {
	transfee, err := QueryTransFee(currency, feeCurrency)
	if err != nil {
		return constants.RC_SYSTEM_ERR
	}
	if transfee == nil {
		return constants.RC_INVALID_CURRENCY
	}
	valueFloat := utils.Str2Float64(value)
	feeInt64 := utils.FloatStrToLVTint(fee)
	feeMaxStr := strconv.FormatFloat(transfee.FeeMax, 'f', -1, 64)
	feeMinStr := strconv.FormatFloat(transfee.FeeMin, 'f', -1, 64)
	feeMaxInt64 := utils.FloatStrToLVTint(feeMaxStr)
	feeMinInt64 := utils.FloatStrToLVTint(feeMinStr)
	if feeInt64 > feeMaxInt64 {
		return constants.RC_TRANSFER_FEE_ERROR
	}

	realFee := valueFloat * transfee.FeeRate * transfee.Discount
	realFeeStr := strconv.FormatFloat(realFee, 'f', -1, 64)
	realFeeInt64 := utils.FloatStrToLVTint(realFeeStr)
	if realFeeInt64 > feeMaxInt64 {
		// 大于转账费率最大值，实际转账费率等于最大值
		realFeeInt64 = feeMaxInt64
	} else if realFeeInt64 < feeMinInt64 {
		// 小于转账费率最小值，实际转账费率等于最小值
		realFeeInt64 = feeMinInt64
	}

	if realFeeInt64 != feeInt64 {
		return constants.RC_TRANSFER_FEE_ERROR
	}
	return constants.RC_OK
}
