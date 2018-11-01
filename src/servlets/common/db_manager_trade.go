package common

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"utils/config"
	"utils/logger"
	"utils"
)

const (
	TRADES = "dt_trade"
)


var tradeSession *mgo.Session
var tradeConfig config.MongoConfig

func InitTradeMongoDB() {
	config := config.GetConfig()
	tradeConfig = config.Trade
	connStr := fmt.Sprintf("%s?maxPoolSize=%d", tradeConfig.DBHost, tradeConfig.MaxConn)
	logger.Info("conn mongo db ---> ", connStr)
	var connErr error = nil
	tradeSession, connErr = mgo.Dial(connStr)
	if connErr != nil {
		logger.Error("connect mongo db error", connErr.Error())
		panic(connErr)
	}
	tradeSession.SetPoolLimit(tradeConfig.MaxConn)
}

func InsertTradeInfo(info ...TradeInfo) error {
	session := tradeSession.Clone()
	defer session.Close()
	session.SetSafe(sessionSafe)
	collection := session.DB(tradeConfig.DBDatabase).C(TRADES)
	for i:=0;i<len(info);i++ {
		err := collection.Insert(info[i])
		if err != nil {
			logger.Error("add trade info error, tradeNo:", info[i].TradeNo, "error:", err.Error())
		}
	}
	return nil
}

func QueryTrades(query interface{}, limit int) []TradeInfo {
	session := tradeSession.Clone()
	defer session.Close()
	logger.Debug("mongo query :", utils.ToJSON(query))
	collection := session.DB(tradeConfig.DBDatabase).C(TRADES)
	res := make([]TradeInfo,0)
	err := collection.Find(query).Sort("-finish_time").Limit(limit).All(&res)
	if err != nil && err != mgo.ErrNotFound {
		logger.Error("query mongo db error ", err.Error())
		return nil
	}
	logger.Debug("query res ", utils.ToJSON (res))
	return res
}