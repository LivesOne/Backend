package log_cleaner

import (
	"servlets/common"
	"utils"
	"servlets/constants"
	"utils/logger"
)

func cleanerTxid()bool{
	rows := common.FindTopTxid(1)
	if rows != nil && len(rows) > 0 {
		txid := utils.Str2Int64(rows[0]["txid"])
		if !common.ExistsPending(txid) {
			common.RemoveTXID(txid)
		}
		return true
	}
	return false
}

func cleanerPending()bool{
	pd := common.FindTopPending(1)
	if pd != nil && pd.Id > 0 {
		perPending,flag := common.FindAndModifyPending(pd.Id,pd.From,constants.TX_STATUS_COMMIT)
		if flag {
			if perPending.Status == constants.TX_STATUS_DEFAULT {
				if common.CheckTXID(pd.Id) {
					err := common.InsertCommited(perPending)
					if err == nil {
						common.DeletePending(perPending.Id)
						common.RemoveTXID(perPending.Id)
					}else{
						logger.Error("insert commited error ",err.Error())
					}
				}
			}
		}
		return true
	}
	return false
}

