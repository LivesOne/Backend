package log_cleaner

import (
	"time"
	"utils/logger"
	"utils"
	"math/rand"
)

func StartJob(){
	go func(){
		logger.Info("start cleaner pending job ---> ",utils.GetFormatDateNow())
		for {
			startTask()
		}
	}()
}


func startTask(){
	//循环至每月数据
	for cleanerPending() {}
	//随机3-5秒休眠
	s := random3To5()
	logger.Info("sleep task second ",s)
	time.Sleep(time.Duration(s) * time.Second)
}

func random3To5()int{
	rand.Seed(time.Now().UnixNano())
	return 3+rand.Intn(3)
}