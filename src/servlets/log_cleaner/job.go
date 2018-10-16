package log_cleaner

import (
	"math/rand"
	"servlets/common"
	"time"
	"utils"
	"utils/logger"
)

func StartJob() {
	go func() {
		logger.Info("start cleaner lvt pending job ---> ", utils.GetFormatDateNow())
		for {
			startTask()
		}
	}()
	go func() {
		logger.Info("start cleaner lvtc pending job ---> ", utils.GetFormatDateNow())
		for {
			startLvtcTask()
		}
	}()

	go func() {
		logger.Info("start Listen Txhistory Queue job ---> ", utils.GetFormatDateNow())
		common.ListenTxhistoryQueue()
	}()
	go func() {
		logger.Info("start Push TxHistory ByTimer job ---> ", utils.GetFormatDateNow())
		common.PushTxHistoryByTimer()
	}()
}

func startTask() {
	//循环至每月数据
	for cleanerPending() {
	}
	//随机3-5秒休眠
	s := random3To5()
	//logger.Debug("sleep task second ", s)
	time.Sleep(time.Duration(s) * time.Second)
}

func startLvtcTask() {
	//循环至每月数据
	for cleanerLVTCPending() {
	}
	//随机3-5秒休眠
	s := random3To5()
	//logger.Debug("sleep task second ", s)
	time.Sleep(time.Duration(s) * time.Second)
}

func random3To5() int {
	rand.Seed(time.Now().UnixNano())
	return 3 + rand.Intn(3)
}
