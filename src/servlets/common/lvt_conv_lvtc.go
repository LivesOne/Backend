package common

import (
	"utils/logger"
	"servlets/constants"
	"utils/config"
	"utils"
	"database/sql"
)

func Lvt2Lvtc(uid int64)(int64,int64,constants.Error){

	tx ,err := gDBAsset.Begin()
	if err != nil {
		logger.Error("begin trans error",err.Error())
	}

	lvt,lvtc,err := lvt2LvtcInMysql(uid,tx)

	if err != nil{
		logger.Error("begin trans error",err.Error())
		tx.Rollback()
		return 0,0,constants.RC_SYSTEM_ERR
	}

	if lvt == 0 {
		tx.Rollback()
		return 0,0,constants.RC_OK
	}



	if ok,e := buildLvtTxHistory(uid,lvt,tx);!ok {
		logger.Error("build lvt tx history failed ,rollback the tx")
		tx.Rollback()
		return lvt,lvtc,e
	}

	if ok,e := buildLvtcTxHistory(uid,lvt,tx);!ok {
		logger.Error("build lvtc tx history failed ,rollback the tx")
		tx.Rollback()
		return lvt,lvtc,e
	}


	err = tx.Commit()
	if err != nil{
		logger.Error("commit trans error",err.Error())
		return 0,0,constants.RC_SYSTEM_ERR
	}

	return lvt,lvtc,constants.RC_OK


}

func buildLvtTxHistory(uid,lvt int64,tx *sql.Tx)(bool,constants.Error){
	txid := GenerateTxID()

	if txid == -1 {
		logger.Error("get txid error")
		return false,constants.RC_SYSTEM_ERR
	}
	systemUid := config.GetConfig().Lvt2LvtcSystemAccountUid


	f, c := TransAccountLvtByTx(txid,uid, systemUid, lvt,tx)
	if f {
		//成功 插入commited lvt
		txh := &DTTXHistory{
			Id:         txid,
			Status:     constants.TX_STATUS_DEFAULT,
			Type:       constants.TX_TYPE_ASSET_CONV,
			From:       uid,
			To:         systemUid,
			Value:      lvt,
			Ts:         utils.TXIDToTimeStamp13(txid),
			Code:       constants.TX_CODE_SUCC,
		}
		err := InsertCommited(txh)
		if !CheckDup(err) {
			logger.Error("insert mongo error",err.Error())
			return false,constants.RC_SYSTEM_ERR
		}
	} else {
		return false,getConvResCode(c)
	}
	return true,constants.RC_OK
}


func buildLvtcTxHistory(uid,lvtc int64,tx *sql.Tx)(bool,constants.Error){
	txid := GenerateTxID()

	if txid == -1 {
		logger.Error("get txid error")
		return false,constants.RC_SYSTEM_ERR
	}
	systemUid := config.GetConfig().Lvt2LvtcSystemAccountUid


	f, c := TransAccountLvtcByTx(txid,systemUid, uid, lvtc,tx)
	if f {
		//成功 插入commited lvtc
		txh := &DTTXHistory{
			Id:         txid,
			Status:     constants.TX_STATUS_DEFAULT,
			Type:       constants.TX_TYPE_ASSET_CONV,
			From:       systemUid,
			To:         uid,
			Value:      lvtc,
			Ts:         utils.TXIDToTimeStamp13(txid),
			Code:       constants.TX_CODE_SUCC,
		}
		err := InsertLVTCCommited(txh)
		if !CheckDup(err) {
			logger.Error("insert mongo error",err.Error())
			return false,constants.RC_SYSTEM_ERR
		}
	} else {
		return false,getConvResCode(c)
	}
	return true,constants.RC_OK
}

func getConvResCode(code int)constants.Error{
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