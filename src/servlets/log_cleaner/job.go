package log_cleaner

import (
	"github.com/robfig/cron"
	"time"
	"utils/logger"
	"utils"
)

func StartJob(){
	c := cron.NewWithLocation(time.UTC)
	//每个整点执行
	c.AddFunc("0 0 * * * ?", func() {
		startTask()
	})
	//启动定时任务
	c.Start()
}


func startTask(){
	logger.Info("start cleaner task job time :",utils.GetFormatDateNow())
	for cleanerTxid() {}
	for cleanerPending() {}
	logger.Info("end cleaner task job time :",utils.GetFormatDateNow())
}