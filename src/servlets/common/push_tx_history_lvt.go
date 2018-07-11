package common

import (
	"github.com/garyburd/redigo/redis"
	"servlets/constants"
	"time"
	"utils"
	"utils/logger"
)

func ListenTxhistoryQueue() {

	for {
		if redisPool == nil || tSession == nil {
			logger.Info("push_tx_history_lvt redis/mongo pool not init")
			time.Sleep(10 * 1e9)
			continue
		}

		results, _ := redis.Strings(rdsDo("BLPOP", constants.PUSH_TX_HISTORY_LVT_QUEUE_NAME, 0))
		if results != nil && len(results) >= 2 {
			logger.Debug(len(results))
			logger.Debug("jsonstr:", results[0], results[1])
			txh := new(DTTXHistory)

			if err := utils.FromJson(results[1], txh); err == nil {
				currency := GetWithdrawalCurrency(txh.TradeNo)
				if currency == "LVTC" {
					err = InsertLVTCCommited(txh)
				} else {
					err = InsertCommited(txh)
				}
				if err != nil {
					logger.Error("tx_history_lvt_tmp insert mongo error ", err.Error())
					rdsDo("RPUSH", constants.PUSH_TX_HISTORY_LVT_QUEUE_NAME, utils.ToJSON(txh))
				}
			}
		}
		time.Sleep(10 * 1e9)
	}
}

func PushTxHistoryByTimer() {
	c := time.Tick(time.Hour * 4)
	for {
		if gDBAsset != nil && tSession != nil {
			hour, _ := time.ParseDuration("-1h")
			before4Hour := time.Now().Add(4 * hour)
			dTTXHistoryList := QueryTxhistoryLvtTmpByTimie(utils.GetTimestamp13ByTime(before4Hour))
			if dTTXHistoryList != nil {
				for _, dTTXHistory := range dTTXHistoryList {
					var err error
					if dTTXHistory.Currency == "LVTC" {
						err = InsertLVTCCommited(dTTXHistory)
					} else {
						err = InsertCommited(dTTXHistory)
					}

					if err == nil {
						DeleteTxhistoryLvtTmpByTxid(dTTXHistory.Id)
					}
				}
			}
			c = time.Tick(time.Hour * 4)
		} else {
			logger.Info("push_tx_history_lvt mysql pool not init")
			c = time.Tick(time.Second * 10)
		}

		<-c
	}
}
