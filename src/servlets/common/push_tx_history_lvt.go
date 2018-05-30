package common

import (
	"servlets/constants"
	"utils/logger"
	"bytes"
	"encoding/json"
	"time"
	"utils"
)

func ListenTxhistoryQueue()  {
	for  {
		reply, err := rdsDo("BLPOP", constants.PUSH_TX_HISTORY_LVT_QUEUE_NAME)
		if err == nil {
			rdsDo("RPUSH", constants.PUSH_TX_HISTORY_LVT_QUEUE_NAME, reply)
		} else {
			time.Sleep(10 * 1e9)
		}

		var txh DTTXHistory
		decoder := json.NewDecoder(bytes.NewReader(reply.([]byte)))
		if err := decoder.Decode(&txh); err != nil {
			logger.Error("redis tx history lvt parse error ", err.Error())
			rdsDo("RPUSH", constants.PUSH_TX_HISTORY_LVT_QUEUE_NAME, reply)
		}

		err = InsertCommited(&txh)
		if err != nil {
			logger.Error("tx_history_lv_tmp insert mongo error ", err.Error())
			rdsDo("RPUSH", constants.PUSH_TX_HISTORY_LVT_QUEUE_NAME, reply)
		}
	}
}

func PushTxHistoryByTimer()  {
	c := time.Tick(time.Hour * 4)
	for {
		hour, _ := time.ParseDuration("-1h")
		before4Hour := time.Now().Add(4 * hour)
		dTTXHistoryList := QueryTxhistoryLvtTmpByTimie(utils.GetTimestamp13ByTime(before4Hour))
		for _,dTTXHistory := range dTTXHistoryList {
			err := InsertCommited(dTTXHistory)
			if err == nil {
				DeleteTxhistoryLvtTmpByTxid(dTTXHistory.Id)
			}
		}
		<- c
	}
}
