package common

import (
	"database/sql"
	"gitlab.maxthon.net/cloud/livesone-user-micro/src/proto"
	"servlets/constants"
	"servlets/rpc"
	"strconv"
	"utils"
	"utils/config"
	"utils/logger"
)

func Lvtc2Bsv(uid int64, lvtc_nums string, bsv_nums string) (int64, int64,
	constants.Error) {
	lvtc_num, err := strconv.ParseInt(lvtc_nums, 10, 64)
	bsv_num, err := strconv.ParseInt(bsv_nums, 10, 64)
	tx, err := gDBAsset.Begin()
	if err != nil {
		logger.Error("begin trans error", err.Error())
	}

	lvtc, err := lvtc2BsvInMysql(uid, tx)

	if err != nil {
		logger.Error("begin bsv trans error", err.Error())
		tx.Rollback()
		return 0, 0, constants.RC_SYSTEM_ERR
	}

	if lvtc == 0 {
		tx.Rollback()
		return 0, 0, constants.RC_OK
	}
	if lvtc != lvtc_num {
		tx.Rollback()
		return 0, 0, constants.RC_TRANS_LVTC_NUM_ERROR
	}
	systemUid := config.GetConfig().Lvtc2BsvSystemAccountUid

	if ok, e := commonConvTransBsv(uid, systemUid, lvtc, bsv_num, tx); !ok {
		tx.Rollback()
		return lvtc, bsv_num, e
	}

	err = tx.Commit()
	if err != nil {
		logger.Error("commit trans error", err.Error())
		return 0, 0, constants.RC_SYSTEM_ERR
	}

	return lvtc, bsv_num, constants.RC_OK

}
func commonConvTransBsv(uid, systemUid, lvtc, bsv int64, tx *sql.Tx) (bool, constants.Error) {
	if txid, e := buildLvtcBsvTxHistory(uid, systemUid, lvtc, tx); txid < 0 {
		logger.Error("build lvtc tx history failed ,rollback the tx")
		return false, e
	} else {
		lvtcTradeNo := GenerateTradeNo(constants.TRADE_TYPE_CONVERSION,
			constants.TX_SUB_TYPE_ASSET_CONV)
		bsvTradeNo := GenerateTradeNo(constants.TRADE_TYPE_CONVERSION,
			constants.TX_SUB_TYPE_ASSET_CONV)
		addTradeInfoOfLVTCS(lvtcTradeNo, bsvTradeNo, uid, systemUid, lvtc, txid)
		txidLvtc, e := buildBsvLvtcTxHistory(uid, systemUid, lvtc, bsv, tx)
		if txidLvtc < 0 {
			logger.Error("build bsv tx history failed ,rollback the tx")
			DeleteCommited(txid)
			return false, e
		}
		addTradeInfoOfBSV(bsvTradeNo, lvtcTradeNo, systemUid, uid, bsv, txidLvtc)
	}
	return true, constants.RC_OK
}

func buildLvtcBsvTxHistory(uid, systemUid, lvtc int64, tx *sql.Tx) (int64, constants.Error) {
	txid := GenerateTxID()
	if txid == -1 {
		logger.Error("get txid error")
		return -1, constants.RC_SYSTEM_ERR
	}
	//插入commited lvtc
	txh := &DTTXHistory{
		Id:     txid,
		Status: constants.TX_STATUS_DEFAULT,
		Type:   constants.TX_SUB_TYPE_ASSET_CONV,
		From:   uid,
		To:     systemUid,
		Value:  lvtc,
		Ts:     utils.TXIDToTimeStamp13(txid),
		Code:   constants.TX_CODE_SUCC,
	}
	err := InsertLVTCCommited(txh)
	if !CheckDup(err) {
		logger.Error("insert mongo error", err.Error())
		return -1, constants.RC_SYSTEM_ERR
	}
	return txid, constants.RC_OK
}
func buildBsvLvtcTxHistory(uid, systemUid, lvtc, bsv int64, tx *sql.Tx) (int64, constants.Error) {
	txid := GenerateTxID()

	if txid == -1 {
		logger.Error("get txid error")
		return txid, constants.RC_SYSTEM_ERR
	}

	f, c := ConvAccountLvtcBsvByTx(txid, systemUid, uid, lvtc, bsv, tx)
	if f {
		//成功 插入commited bsv
		txh := &DTTXHistory{
			Id:     txid,
			Status: constants.TX_STATUS_DEFAULT,
			Type:   constants.TX_TYPE_ASSET_CONV,
			From:   systemUid,
			To:     uid,
			Value:  bsv,
			Ts:     utils.TXIDToTimeStamp13(txid),
			Code:   constants.TX_CODE_SUCC,
		}
		err := InsertBSVCommited(txh)
		if !CheckDup(err) {
			logger.Error("insert mongo error", err.Error())
			return -1, constants.RC_SYSTEM_ERR
		}
	} else {
		return -1, getConvResCode(c)
	}
	return txid, constants.RC_OK
}

//status为成功
func addTradeInfoOfLVTCS(lvtcTradeNo, bsvTradeNo string, from, to, amount, txid int64) {
	conversion := TradeConversion{
		OriginalCurrency: "LVTC",
		TargetCurrency:   "BSV",
	}
	fromName, _ := rpc.GetUserField(from, microuser.UserField_NICKNAME)
	toName, _ := rpc.GetUserField(to, microuser.UserField_NICKNAME)
	lvtcTradeInfo := TradeInfo{
		TradeNo:         lvtcTradeNo,
		Type:            constants.TRADE_TYPE_CONVERSION,
		SubType:         constants.TX_SUB_TYPE_ASSET_CONV,
		From:            from,
		FromName:        fromName,
		To:              to,
		ToName:          toName,
		Amount:          amount,
		Decimal:         8,
		Currency:        "LVTC",
		CreateTime:      utils.TXIDToTimeStamp13(txid),
		Status:          constants.TRADE_STATUS_SUCC,
		Txid:            txid,
		OriginalTradeNo: bsvTradeNo,
		Conversion:      &conversion,
	}
	InsertTradeInfo(lvtcTradeInfo)
}

//status为成功
func addTradeInfoOfBSV(bsvTradeNo, lvtcTradeNo string, from, to, amount, txid int64) {
	conversion := TradeConversion{
		OriginalCurrency: "LVTC",
		TargetCurrency:   "BSV",
	}
	fromName, _ := rpc.GetUserField(from, microuser.UserField_NICKNAME)
	toName, _ := rpc.GetUserField(to, microuser.UserField_NICKNAME)
	bsvTradeInfo := TradeInfo{
		TradeNo:         bsvTradeNo,
		Type:            constants.TRADE_TYPE_CONVERSION,
		SubType:         constants.TX_SUB_TYPE_ASSET_CONV,
		From:            from,
		FromName:        fromName,
		To:              to,
		ToName:          toName,
		Amount:          amount,
		Decimal:         8,
		Currency:        "BSV",
		CreateTime:      utils.TXIDToTimeStamp13(txid),
		Status:          constants.TRADE_STATUS_SUCC,
		Txid:            txid,
		OriginalTradeNo: lvtcTradeNo,
		Conversion:      &conversion,
	}
	InsertTradeInfo(bsvTradeInfo)
}

