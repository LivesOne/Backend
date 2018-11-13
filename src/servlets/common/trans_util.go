package common

import (
	"database/sql"
	"fmt"
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

func PrepareLVTTrans(from, to int64, txTpye int, value, bizContent, remark string) (string, constants.Error) {
	tradeNo := GenerateTradeNo(constants.TRADE_TYPE_TRANSFER, constants.TX_TYPE_TRANS)
	txid := GenerateTxID()
	if txid == -1 {
		logger.Error("txid is -1  ")
		return "", constants.RC_SYSTEM_ERR
	}
	txh := DTTXHistory{
		Id:         txid,
		TradeNo:    tradeNo,
		Status:     constants.TX_STATUS_DEFAULT,
		Type:       txTpye,
		From:       from,
		To:         to,
		Value:      utils.FloatStrToLVTint(value),
		Ts:         utils.TXIDToTimeStamp13(txid),
		Code:       constants.TX_CODE_SUCC,
		BizContent: bizContent,
		Remark:     remark,
		Currency:   CURRENCY_LVT,
	}
	err := InsertPending(&txh)
	if err != nil {
		logger.Error("insert mongo db:dt_pending error ", err.Error())
		return "", constants.RC_SYSTEM_ERR
	}
	return utils.Int642Str(txid), constants.RC_OK
}
func PrepareLVTCTrans(from, to int64, txTpye int, value, bizContent, remark string) (string, constants.Error) {
	tradeNo := GenerateTradeNo(constants.TRADE_TYPE_TRANSFER, constants.TX_TYPE_TRANS)
	txid := GenerateTxID()
	if txid == -1 {
		logger.Error("txid is -1  ")
		return "", constants.RC_SYSTEM_ERR
	}
	txh := &DTTXHistory{
		Id:         txid,
		TradeNo:    tradeNo,
		Status:     constants.TX_STATUS_DEFAULT,
		Type:       txTpye,
		From:       from,
		To:         to,
		Value:      utils.FloatStrToLVTint(value),
		Ts:         utils.TXIDToTimeStamp13(txid),
		Code:       constants.TX_CODE_SUCC,
		BizContent: bizContent,
		Remark:     remark,
		Currency:   CURRENCY_LVTC,
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
		return constants.RC_TRANS_TIMEOUT
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
		return transInt2Error(intErr)
	}
	//删除pending
	DeletePendingByInfo(perPending)
	//识别类型进行操作
	switch perPending.Type {
	case constants.TX_TYPE_TRANS:
		var tradesArray []TradeInfo
		finishTime := utils.GetTimestamp13()
		// 插入交易记录单：转账
		fromName, err := GetCacheUserField(perPending.From, USER_CACHE_REDIS_FIELD_NAME_NICKNAME)
		if err != nil {
			logger.Info("get uid:", perPending.From, " nick name err,", err)
		}
		toName, err := GetCacheUserField(perPending.To, USER_CACHE_REDIS_FIELD_NAME_NICKNAME)
		if err != nil {
			logger.Info("get uid:", perPending.To, " nick name err,", err)
		}
		trade := TradeInfo{
			TradeNo: perPending.TradeNo, Txid: perPending.Id, Status: constants.TRADE_STATUS_SUCC,
			Type: constants.TRADE_TYPE_TRANSFER, SubType: perPending.Type, From: perPending.From,
			To: perPending.To, Amount: perPending.Value, Decimal: constants.TRADE_DECIMAIL,
			FromName: fromName, ToName: toName, Subject: bizContent.Remark,
			Currency: constants.TRADE_CURRENCY_LVT, CreateTime: perPending.Ts, FinishTime: finishTime,
		}
		if feeTxid > 0 && len(feeTradeNo) > 0 {
			// 插入交易记录单：手续费
			feeToName, err := GetCacheUserField(transFeeAcc, USER_CACHE_REDIS_FIELD_NAME_NICKNAME)
			if err != nil {
				logger.Info("get uid:", transFeeAcc, " nick name err,", err)
			}
			trade.FeeTradeNo = feeTradeNo
			feeTrade := TradeInfo{
				TradeNo: feeTradeNo, OriginalTradeNo: perPending.TradeNo, Txid: feeTxid,
				Status: constants.TRADE_STATUS_SUCC, Type: constants.TRADE_TYPE_FEE, SubType: feeSubType,
				From: perPending.From, To: transFeeAcc, Amount: bizContent.Fee, Decimal: constants.TRADE_DECIMAIL,
				FromName: fromName, ToName: feeToName,
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

func CommitLVTCTrans(uidStr, txIdStr string) (retErr constants.Error) {
	txid := utils.Str2Int64(txIdStr)
	uid := utils.Str2Int64(uidStr)
	perPending, flag := FindAndModifyLVTCPending(txid, uid, constants.TX_STATUS_COMMIT)
	//未查到数据，返回处理中
	if !flag || perPending.Status != constants.TX_STATUS_DEFAULT {
		return constants.RC_TRANS_TIMEOUT
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
		fromName, err := GetCacheUserField(perPending.From, USER_CACHE_REDIS_FIELD_NAME_NICKNAME)
		if err != nil {
			logger.Info("get uid:", perPending.From, " nick name err,", err)
		}
		toName, err := GetCacheUserField(perPending.To, USER_CACHE_REDIS_FIELD_NAME_NICKNAME)
		if err != nil {
			logger.Info("get uid:", perPending.To, " nick name err,", err)
		}
		trade := TradeInfo{
			TradeNo: perPending.TradeNo, Txid: perPending.Id, Status: constants.TRADE_STATUS_SUCC,
			Type: constants.TRADE_TYPE_TRANSFER, SubType: perPending.Type, From: perPending.From,
			To: perPending.To, Amount: perPending.Value, Decimal: constants.TRADE_DECIMAIL,
			FromName: fromName, ToName: toName, Subject: bizContent.Remark,
			Currency: CURRENCY_LVTC, CreateTime: perPending.Ts, FinishTime: finishTime,
		}
		if feeTxid > 0 && len(feeTradeNo) > 0 {
			// 插入交易记录单：手续费
			trade.FeeTradeNo = feeTradeNo
			feeToName, err := GetCacheUserField(transFeeAcc, USER_CACHE_REDIS_FIELD_NAME_NICKNAME)
			if err != nil {
				logger.Info("get uid:", transFeeAcc, " nick name err,", err)
			}
			feeTrade := TradeInfo{
				TradeNo: feeTradeNo, OriginalTradeNo: perPending.TradeNo, Txid: feeTxid,
				Status: constants.TRADE_STATUS_SUCC, Type: constants.TRADE_TYPE_FEE, SubType: feeSubType,
				From: perPending.From, To: transFeeAcc, Amount: bizContent.Fee, Decimal: constants.TRADE_DECIMAIL,
				FromName: fromName, ToName: feeToName,
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

func PrepareTradePending(from, to int64, value int64, txTpye int, bizContent string) (
	string, string, constants.Error) {
	tradeNo := GenerateTradeNo(constants.TRADE_TYPE_TRANSFER, txTpye)
	txid := GenerateTxID()
	if txid == -1 {
		logger.Error("txid is -1  ")
		return "", "", constants.RC_SYSTEM_ERR
	}
	if err := InsertTradePending(txid, from, to, tradeNo, bizContent, value, txTpye); err != nil {
		logger.Error("insert trade pending error", err.Error())
		return "", "", constants.RC_SYSTEM_ERR
	}
	return utils.Int642Str(txid), tradeNo, constants.RC_OK
}

func CommitTransfer(uidStr, txidStr, currency string) (retErr constants.Error) {

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
	var intErr int
	var decimal, feeDecimal = constants.TRADE_DECIMAIL, constants.TRADE_DECIMAIL
	txid := utils.Str2Int64(txidStr)
	switch currency {
	case constants.TRADE_CURRENCY_EOS:
		decimal = constants.TRADE_EOS_DECIMAIL
		_, intErr = EosTransCommit(txid, tp.From, to, tp.Value, tp.TradeNo, tp.Type, tx)
	case constants.TRADE_CURRENCY_BTC:
		_, intErr = BtcTransCommit(txid, tp.From, to, tp.Value, tp.TradeNo, tp.Type, tx)
	case constants.TRADE_CURRENCY_ETH:
		_, intErr = EthTransCommit(txid, tp.From, to, tp.Value, tp.TradeNo, tp.Type, tx)
	}
	if bizContent.FeeCurrency == constants.TRADE_CURRENCY_EOS {
		feeDecimal = constants.TRADE_EOS_DECIMAIL
	}
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
		fromName, err := GetCacheUserField(tp.From, USER_CACHE_REDIS_FIELD_NAME_NICKNAME)
		if err != nil {
			logger.Info("get uid:", tp.From, " nick name err,", err)
		}
		toName, err := GetCacheUserField(tp.To, USER_CACHE_REDIS_FIELD_NAME_NICKNAME)
		if err != nil {
			logger.Info("get uid:", tp.To, " nick name err,", err)
		}
		trade := TradeInfo{
			TradeNo: tp.TradeNo, Txid: txid, Status: constants.TRADE_STATUS_SUCC,
			Currency: currency, Type: constants.TRADE_TYPE_TRANSFER,
			SubType: tp.Type, From: tp.From, To: tp.To, Decimal: decimal,
			FromName: fromName, ToName: toName, Subject: bizContent.Remark,
			Amount: tp.Value, CreateTime: tp.Ts, FinishTime: finishTime,
		}
		if feeTxid > 0 && len(feeTradeNo) > 0 {
			// 插入交易记录单：手续费
			trade.FeeTradeNo = feeTradeNo
			feeToName, err := GetCacheUserField(transFeeAcc, USER_CACHE_REDIS_FIELD_NAME_NICKNAME)
			if err != nil {
				logger.Info("get uid:", transFeeAcc, " nick name err,", err)
			}
			feeTrade := TradeInfo{
				TradeNo: feeTradeNo, OriginalTradeNo: tp.TradeNo, Txid: feeTxid,
				Status: constants.TRADE_STATUS_SUCC, Type: constants.TRADE_TYPE_FEE, SubType: feeSubType,
				From: tp.From, To: transFeeAcc, Amount: bizContent.Fee, Decimal: feeDecimal,
				FromName: fromName, ToName: feeToName,
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
		SetDailyTransAmount(uid, currency, tp.Value)
	}

	return constants.RC_OK
}

func TransFeeCommit(tx *sql.Tx, from, fee int64, currency string) (int64, string, int, int64, constants.Error) {
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
	case constants.TRADE_CURRENCY_EOS:
		_, intErr = EosTransCommit(feeTxid, from, transFeeAcc,
			fee, feeTradeNo, feeSubType, tx)
	case constants.TRADE_CURRENCY_BTC:
		_, intErr = BtcTransCommit(feeTxid, from, transFeeAcc,
			fee, feeTradeNo, feeSubType, tx)
	case constants.TRADE_CURRENCY_ETH:
		_, intErr = EthTransCommit(feeTxid, from, transFeeAcc,
			fee, feeTradeNo, feeSubType, tx)
	case constants.TRADE_CURRENCY_LVT:
		_, intErr = TransAccountLvt(tx, feeDth)
	case constants.TRADE_CURRENCY_LVTC:
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

func verifyTrans(uid, value int64, currency string) constants.Error {
	// 校验单笔最低限额
	if e := CheckSingleTransAmount(currency, value); e != constants.RC_OK {
		return e
	}
	// 校验转账日限额
	if f, e := CheckDailyTransAmount(uid, currency, value); !f {
		return e
	}
	return constants.RC_OK
}

func VerifyEthTrans(uid int64, valueStr string) constants.Error {
	value := utils.FloatStrToLVTint(valueStr)
	return verifyTrans(uid, value, constants.TRADE_CURRENCY_ETH)
}

func VerifyEosTrans(uid int64, valueStr string) constants.Error {
	value := utils.FloatStrToEOSint(valueStr)
	return verifyTrans(uid, value, constants.TRADE_CURRENCY_EOS)
}

func VerifyBtcTrans(uid int64, valueStr string) constants.Error {
	value := utils.FloatStrToLVTint(valueStr)
	return verifyTrans(uid, value, constants.TRADE_CURRENCY_BTC)
}

func CheckTransFee(value, fee, currency, feeCurrency string) constants.Error {
	transfee, err := QueryTransFee(currency, feeCurrency)
	if err != nil {
		return constants.RC_SYSTEM_ERR
	}
	if transfee == nil {
		return constants.RC_INVALID_CURRENCY
	}

	var feeInt64, realFeeInt64, feeMaxInt64, feeMinInt64 int64
	valueFloat := utils.Str2Float64(value)
	feeMaxStr := transfee.FeeMax
	feeMinStr := transfee.FeeMin
	feeRate := utils.Str2Float64(transfee.FeeRate)
	discount := utils.Str2Float64(transfee.Discount)
	realFee := valueFloat * feeRate * discount
	realFeeStr := strconv.FormatFloat(realFee, 'f', -1, 64)
	if feeCurrency == constants.TRADE_CURRENCY_EOS {
		feeInt64 = utils.FloatStrToEOSint(fee)
		feeMaxInt64 = utils.FloatStrToEOSint(feeMaxStr)
		feeMinInt64 = utils.FloatStrToEOSint(feeMinStr)
		realFeeInt64 = utils.FloatStrToEOSint(realFeeStr)
	} else {
		feeInt64 = utils.FloatStrToLVTint(fee)
		feeMaxInt64 = utils.FloatStrToLVTint(feeMaxStr)
		feeMinInt64 = utils.FloatStrToLVTint(feeMinStr)
		realFeeInt64 = utils.FloatStrToLVTint(realFeeStr)
	}

	if feeInt64 > feeMaxInt64 {
		return constants.RC_TRANSFER_FEE_ERROR
	}
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

func TransferPrepare(from, to int64, amount, fee, currency, feeCurrency, remark string) (string, string, constants.Error) {
	if !config.GetConfig().CheckSupportedCoin(currency) || !config.GetConfig().CheckSupportedCoin(feeCurrency) {
		return "", "", constants.RC_INVALID_CURRENCY
	}
	var currencyDecimal, feeCurrencyDecimal, feeDecimail int
	if strings.EqualFold(currency, CURRENCY_EOS) {
		currencyDecimal = utils.CONV_EOS
	} else {
		currencyDecimal = utils.CONV_LVT
	}
	if strings.EqualFold(feeCurrency, CURRENCY_EOS) {
		feeCurrencyDecimal = utils.CONV_EOS
		feeDecimail = constants.TRADE_EOS_DECIMAIL
	//} else if strings.EqualFold(feeCurrency, CURRENCY_LVTC) {
	//	feeDecimail = constants.TRADE_EOS_DECIMAIL
	//}
	} else {
		feeCurrencyDecimal = utils.CONV_LVT
		feeDecimail = constants.TRADE_DECIMAIL
	}
	realFee, err := calculationFeeAndCheckQuotaForTransfer(from, utils.Str2Float64(amount), currency, feeCurrency, currencyDecimal)
	if err.Rc != constants.RC_OK.Rc {
		return "", "", err
	}
	realFee = utils.Str2Float64(fmt.Sprintf("%."+utils.Int2Str(feeDecimail)+"f", realFee))
	if utils.Str2Float64(fee) != realFee {
		logger.Info("currency", currency, "feeCurrency", feeCurrency, "fee", utils.Str2Float64(fee), "realFee", realFee)
		return "", "", constants.RC_TRANSFER_FEE_ERROR
	}

	feeInt := utils.FloatStr2CoinsInt(fee, int64(feeCurrencyDecimal))
	amountInt := utils.FloatStr2CoinsInt(amount, int64(currencyDecimal))
	bizContent := TransBizContent{
		FeeCurrency: feeCurrency,
		Fee:         feeInt,
		Remark:      remark,
	}

	tradeNo := GenerateTradeNo(constants.TRADE_TYPE_TRANSFER, constants.TX_TYPE_TRANS)
	txid := GenerateTxID()
	if txid == -1 {
		logger.Error("generate txid err, txid is -1")
		return "", "", constants.RC_SYSTEM_ERR
	}

	if strings.EqualFold(currency, constants.TRADE_CURRENCY_LVTC) {
		txh := DTTXHistory{
			Id:         txid,
			TradeNo:    tradeNo,
			Status:     constants.TX_STATUS_DEFAULT,
			Type:       constants.TX_TYPE_TRANS,
			From:       from,
			To:         to,
			Value:      amountInt,
			Ts:         utils.TXIDToTimeStamp13(txid),
			Code:       constants.TX_CODE_SUCC,
			BizContent: utils.ToJSON(bizContent),
			Remark:     remark,
			Currency:   currency,
		}
		err := InsertLVTCPending(&txh)
		if err != nil {
			logger.Error("insert mongo db:dt_pending error ", err.Error())
			return "", "", constants.RC_SYSTEM_ERR
		}
	} else {
		if err := InsertTradePending(txid, from, to, tradeNo, utils.ToJSON(bizContent), amountInt, constants.TX_TYPE_TRANS); err != nil {
			logger.Error("insert trade pending error", err.Error())
			return "", "", constants.RC_SYSTEM_ERR
		}
	}
	return utils.Int642Str(txid), tradeNo, constants.RC_OK
}

func TransferCommit(uid, txId int64, currency string) constants.Error {
	if !config.GetConfig().CheckSupportedCoin(currency) {
		return constants.RC_INVALID_CURRENCY
	}
	txid_ts := utils.TXIDToTimeStamp13(txId)
	ts := utils.GetTimestamp13()
	//暂时写死10秒
	if ts-txid_ts > TRANS_TIMEOUT {
		//删除pending
		logger.Warn("transfer timeout, transfer time:", txid_ts, "current time", ts)
		DeletePendingByInfo(&DTTXHistory{Id: txId,})
		return constants.RC_TRANS_TIMEOUT
	}

	perPending, err := getPending(txId, uid, currency)
	//未查到数据，返回处理中
	if err.Rc != constants.RC_OK.Rc || perPending == nil {
		return constants.RC_TRANS_TIMEOUT
	}
	perPending.Status = constants.TX_STATUS_COMMIT

	var (
		to         int64
		bizContent TransBizContent
	)
	//识别类型进行操作
	switch perPending.Type {
	case constants.TX_TYPE_TRANS:
		to = perPending.To
	default:
		return constants.RC_PARAM_ERR
	}

	//存在就检测资产初始化状况，未初始化的用户给初始化
	CheckAndInitAsset(to)
	CheckAndInitAsset(perPending.From)
	if perPending.BizContent != "" {
		//解析业务数据，拿到具体数值
		if je := utils.FromJson(perPending.BizContent, &bizContent); je != nil {
			return constants.RC_PARAM_ERR
		}
	}

	tradeNo := GenerateTradeNo(constants.TX_SUB_TYPE_TRANS, perPending.Type)
	feeTradeNo := GenerateTradeNo(constants.TRADE_TYPE_FEE, constants.TX_SUB_TYPE_TRANSFER_FEE)
	timestamp := utils.GetTimestamp13()

	tx, _ := gDBAsset.Begin()

	//扣除转账资产
	err = transfer(txId, uid, to, perPending.Value, timestamp, currency, tradeNo, constants.TX_SUB_TYPE_TRANS, tx)
	if err.Rc != constants.RC_OK.Rc {
		tx.Rollback()
		return err
	}

	txIdFee := GenerateTxID()
	//扣除手续费资产
	err = transfer(txIdFee, uid, config.GetConfig().TransFeeAccountUid, bizContent.Fee, timestamp, bizContent.FeeCurrency, feeTradeNo, constants.TX_SUB_TYPE_TRANSFER_FEE, tx)
	if err.Rc != constants.RC_OK.Rc {
		tx.Rollback()
		return err
	}
	if strings.EqualFold(currency, CURRENCY_LVTC) || strings.EqualFold(currency, CURRENCY_LVT) {
		var eror error
		if strings.EqualFold(currency, CURRENCY_LVTC) {
			eror = InsertLVTCCommited(perPending)
		} else {
			eror = InsertCommited(perPending)
		}
		if !CheckDup(eror) {
			tx.Rollback()
			return constants.RC_SYSTEM_ERR
		}
		DeletePendingByInfo(&DTTXHistory{Id: txId,})
	}
	if strings.EqualFold(currency, CURRENCY_EOS) || strings.EqualFold(currency, CURRENCY_BTC) || strings.EqualFold(currency, CURRENCY_ETH) {
		error := DeleteTradePending(perPending.TradeNo, uid, tx)
		if error != nil {
			tx.Rollback()
			return constants.RC_SYSTEM_ERR
		}
	}
	tx.Commit()

	go func() {
		var currencyDecimal, feeCurrencyDecimal int
		if strings.EqualFold(currency, CURRENCY_EOS) {
			currencyDecimal = constants.TRADE_EOS_DECIMAIL
		} else {
			currencyDecimal = constants.TRADE_DECIMAIL
		}
		if strings.EqualFold(bizContent.FeeCurrency, CURRENCY_EOS) {
			feeCurrencyDecimal = constants.TRADE_EOS_DECIMAIL
		} else {
			feeCurrencyDecimal = constants.TRADE_DECIMAIL
		}
		var tradesArray []TradeInfo
		fromName, _ := GetCacheUserField(uid, USER_CACHE_REDIS_FIELD_NAME_NICKNAME)
		toName, _ := GetCacheUserField(config.GetConfig().TransFeeAccountUid, USER_CACHE_REDIS_FIELD_NAME_NICKNAME)
		feeTradeInfo := TradeInfo{
			TradeNo:         feeTradeNo,
			OriginalTradeNo: tradeNo,
			Type:            constants.TRADE_TYPE_FEE,
			SubType:         constants.TX_SUB_TYPE_TRANSFER_FEE,
			From:            uid,
			FromName:        fromName,
			To:              config.GetConfig().TransFeeAccountUid,
			ToName:          toName,
			Amount:          bizContent.Fee,
			Decimal:         feeCurrencyDecimal,
			Currency:        bizContent.FeeCurrency,
			CreateTime:      ts,
			FinishTime:      ts,
			Status:          constants.TRADE_STATUS_SUCC,
			Txid:            txIdFee,
		}
		tradesArray = append(tradesArray, feeTradeInfo)

		toName, _ = GetCacheUserField(to, USER_CACHE_REDIS_FIELD_NAME_NICKNAME)
		tradeInfo := TradeInfo{
			TradeNo:    tradeNo,
			Type:       constants.TX_SUB_TYPE_TRANS,
			SubType:    perPending.Type,
			Subject:    bizContent.Remark,
			From:       uid,
			FromName:   fromName,
			To:         to,
			ToName:     toName,
			Amount:     perPending.Value,
			Decimal:    currencyDecimal,
			Currency:   currency,
			CreateTime: ts,
			FinishTime: ts,
			Status:     constants.TRADE_STATUS_SUCC,
			Txid:       txId,
			FeeTradeNo: feeTradeNo,
		}
		tradesArray = append(tradesArray, tradeInfo)
		err := InsertTradeInfo(tradesArray...)
		if err != nil {
			logger.Error("transfer insert trade database error, error:", err.Error())
		}
	}()
	return constants.RC_OK
}

func getPending(txId, uid int64, currency string) (*DTTXHistory, constants.Error) {
	var dth *DTTXHistory
	switch strings.ToUpper(currency) {
	case CURRENCY_LVTC:
		dth, flag := FindAndModifyLVTCPending(txId, uid, constants.TX_STATUS_COMMIT)
		//未查到数据，返回处理中
		if !flag || dth.Status != constants.TX_STATUS_DEFAULT {
			return nil, constants.RC_TRANS_TIMEOUT
		}
	case CURRENCY_EOS:
		fallthrough
	case CURRENCY_BTC:
		fallthrough
	case CURRENCY_ETH:
		tp, err := GetTradePendingByTxid(utils.Int642Str(txId), uid)
		if err != nil {
			return nil, constants.RC_SYSTEM_ERR
		}
		if tp == nil {
			return nil, constants.RC_PARAM_ERR
		}
		dth = &DTTXHistory{
			Id:         utils.Str2Int64(tp.Txid),
			Status:     constants.TX_STATUS_DEFAULT,
			Type:       tp.Type,
			TradeNo:    tp.TradeNo,
			From:       tp.From,
			To:         tp.To,
			Value:      tp.Value,
			Ts:         tp.Ts,
			Code:       0,
			Remark:     nil,
			Miner:      nil,
			BizContent: tp.BizContent,
			Currency:   currency,
		}
	default:
		logger.Warn("unsupported currency:",currency)
		return nil, constants.RC_INVALID_CURRENCY
	}
	return dth, constants.RC_OK
}
