package log_cleaner

import (
	"servlets/common"
	"gopkg.in/mgo.v2/bson"
	"utils"
	"servlets/constants"
	"utils/logger"
)

const (

	T_S_120 = 120 * 1000

	T_S_60 = 60 * 1000

)

func cleanerPending()bool{
	mts := utils.GetTimestamp13()
	query := bson.M{
		"_id":bson.M{
			"$lt":getTimeOutTS(mts),
		},
	}
	pd := common.FindTopPending(query,1)
	//拿到数据
	if pd != nil && pd.Id > 0 {
		//反解析txid里面的时间戳
		pdts := utils.TXIDToTimeStamp13(pd.Id)
		//根据status分开处理
		if pd.Status == constants.TX_STATUS_COMMIT {
			//超时处理
			if (mts - pdts) > T_S_120 {
				if common.CheckTXID(pd.Id) {
					err := common.InsertCommited(pd)
					if common.CheckDup(err) {
						common.DeletePendingByInfo(pd)
					}
				} else {
					err := common.InsertFailed(pd)
					if common.CheckDup(err) {
						common.DeletePending(pd.Id)
					} else {
						logger.Error("insert mongo failed error ",err.Error())
					}
				}
			}
		} else {
			if (mts - pdts) > T_S_60 {
				common.DeletePendingByInfo(pd)
			}
		}
		return true
	}
	return false
}

func getTimeOutTS(ts int64) int64{
	//当前时间减去60秒 在减去txid生成时需要减去的时间戳
	delta := ts - T_S_60 - constants.BASE_TIMESTAMP
	timeout := (delta << 22) & 0x7FFFFFFFFFC00000
	return timeout
}