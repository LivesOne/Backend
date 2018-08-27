package common

import (
	"utils/logger"
	"servlets/constants"
	"utils/config"
	"utils"
	"database/sql"
)

func Lvt2Lvtc(uid int64) (int64, int64, constants.Error) {

	tx, err := gDBAsset.Begin()
	if err != nil {
		logger.Error("begin trans error", err.Error())
	}

	lvt, lvtc, err := lvt2LvtcInMysql(uid, tx)

	if err != nil {
		logger.Error("begin trans error", err.Error())
		tx.Rollback()
		return 0, 0, constants.RC_SYSTEM_ERR
	}

	if lvt == 0 {
		tx.Rollback()
		return 0, 0, constants.RC_OK
	}
	systemUid := config.GetConfig().Lvt2LvtcSystemAccountUid

	if ok, e := commonConvTrans(uid, systemUid, lvt, lvtc, tx); !ok {
		tx.Rollback()
		return lvt, lvtc, e
	}

	err = tx.Commit()
	if err != nil {
		logger.Error("commit trans error", err.Error())
		return 0, 0, constants.RC_SYSTEM_ERR
	}

	return lvt, lvtc, constants.RC_OK

}

func Lvt2LvtcDelay(uid int64) (int64, int64, constants.Error) {

	tx, err := gDBAsset.Begin()
	if err != nil {
		logger.Error("begin trans error", err.Error())
	}

	lvt, lvtc, err := lvt2LvtcDelayInMysql(uid, tx)

	if err != nil {
		logger.Error("begin trans error", err.Error())
		tx.Rollback()
		return 0, 0, constants.RC_SYSTEM_ERR
	}

	if lvt == 0 {
		tx.Rollback()
		return 0, 0, constants.RC_OK
	}
	systemUid := config.GetConfig().Lvt2LvtcDelaySystemAccountUid
	if ok, e := commonConvTrans(uid, systemUid, lvt, lvtc, tx); !ok {
		tx.Rollback()
		return lvt, lvtc, e
	}

	err = tx.Commit()
	if err != nil {
		logger.Error("commit trans error", err.Error())
		return 0, 0, constants.RC_SYSTEM_ERR
	}

	return lvt, lvtc, constants.RC_OK

}

func commonConvTrans(uid, systemUid, lvt, lvtc int64, tx *sql.Tx) (bool, constants.Error) {
	if txid, e := buildLvtTxHistory(uid, systemUid, lvt, tx); txid < 0 {
		logger.Error("build lvt tx history failed ,rollback the tx")
		return false, e
	} else {
		lvtTradeNo := GenerateTradeNo(11, constants.TX_TYPE_ASSET_CONV)
		lvtcTradeNo := GenerateTradeNo(11, constants.TX_TYPE_ASSET_CONV)
		addTradeInfoOfLVT(lvtTradeNo, lvtcTradeNo, uid, systemUid, lvt, txid)
		if ok, e := buildLvtcTxHistory(uid, systemUid, lvt, lvtc, tx); !ok {
			logger.Error("build lvtc tx history failed ,rollback the tx")
			DeleteCommited(txid)
			return false, e
		}
		addTradeInfoOfLVTC(lvtcTradeNo, lvtTradeNo, systemUid, uid, lvtc, txid)
	}
	return true, constants.RC_OK
}

func buildLvtTxHistory(uid, systemUid, lvt int64, tx *sql.Tx) (int64, constants.Error) {
	txid := GenerateTxID()

	if txid == -1 {
		logger.Error("get txid error")
		return -1, constants.RC_SYSTEM_ERR
	}

	f, c := TransAccountLvtByTx(txid, uid, systemUid, lvt, tx)
	if f {
		//成功 插入commited lvt
		txh := &DTTXHistory{
			Id:     txid,
			Status: constants.TX_STATUS_DEFAULT,
			Type:   constants.TX_TYPE_ASSET_CONV,
			From:   uid,
			To:     systemUid,
			Value:  lvt,
			Ts:     utils.TXIDToTimeStamp13(txid),
			Code:   constants.TX_CODE_SUCC,
		}
		err := InsertCommited(txh)
		if !CheckDup(err) {
			logger.Error("insert mongo error", err.Error())
			return -1, constants.RC_SYSTEM_ERR
		}
	} else {
		return -1, getConvResCode(c)
	}
	return txid, constants.RC_OK
}

func buildLvtcTxHistory(uid, systemUid, lvt, lvtc int64, tx *sql.Tx) (bool, constants.Error) {
	txid := GenerateTxID()

	if txid == -1 {
		logger.Error("get txid error")
		return false, constants.RC_SYSTEM_ERR
	}

	f, c := ConvAccountLvtcByTx(txid, systemUid, uid, lvt, lvtc, tx)
	if f {
		//成功 插入commited lvtc
		txh := &DTTXHistory{
			Id:     txid,
			Status: constants.TX_STATUS_DEFAULT,
			Type:   constants.TX_TYPE_ASSET_CONV,
			From:   systemUid,
			To:     uid,
			Value:  lvtc,
			Ts:     utils.TXIDToTimeStamp13(txid),
			Code:   constants.TX_CODE_SUCC,
		}
		err := InsertLVTCCommited(txh)
		if !CheckDup(err) {
			logger.Error("insert mongo error", err.Error())
			return false, constants.RC_SYSTEM_ERR
		}
	} else {
		return false, getConvResCode(c)
	}
	return true, constants.RC_OK
}

//status为成功
func addTradeInfoOfLVT(lvtTradeNo, lvtcTradeNo string, from, to, amount, txid int64) {
	lvtTradeInfo := TradeInfo{
		TradeNo:         lvtTradeNo,
		Type:            11,
		SubType:         constants.TX_TYPE_ASSET_CONV,
		From:            from,
		To:              to,
		Amount:          amount,
		Decimal:         8,
		Currency:        "LVT",
		CreateTime:      utils.TXIDToTimeStamp13(txid),
		Status:          2,
		Txid:            txid,
		OriginalTradeNo: lvtcTradeNo,
	}
	InsertTradeInfo(lvtTradeInfo)
}

//status为成功
func addTradeInfoOfLVTC(lvtcTradeNo, lvtTradeNo string, from, to, amount, txid int64) {
	lvtcTradeInfo := TradeInfo{
		TradeNo:         lvtcTradeNo,
		Type:            11,
		SubType:         constants.TX_TYPE_ASSET_CONV,
		From:            from,
		To:              to,
		Amount:          amount,
		Decimal:         8,
		Currency:        "LVTC",
		CreateTime:      utils.TXIDToTimeStamp13(txid),
		Status:          2,
		Txid:            txid,
		OriginalTradeNo: lvtTradeNo,
	}
	InsertTradeInfo(lvtcTradeInfo)
}

func getConvResCode(code int) constants.Error {
	switch code {
	case constants.TRANS_ERR_INSUFFICIENT_BALANCE:
		return constants.RC_INSUFFICIENT_BALANCE
	case constants.TRANS_ERR_SYS:
		return constants.RC_TRANS_IN_PROGRESS
	case constants.TRANS_ERR_ASSET_LIMITED:
		return constants.RC_ACCOUNT_ACCESS_LIMITED
	}
	return constants.RC_SYSTEM_ERR

}
