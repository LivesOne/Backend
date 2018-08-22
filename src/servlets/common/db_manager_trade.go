package common

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"utils/config"
	"utils/logger"
)

const (
	TRADES = "dt_trades"
)

var sessionSafe = &mgo.Safe{WMode: "majority"}
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

func InsertTradeInfo(info TradeInfo) error {
	session := tradeSession.Clone()
	defer session.Close()
	session.SetSafe(sessionSafe)
	collection := session.DB(tradeConfig.DBDatabase).C(TRADES)
	err := collection.Insert(info)
	if err != nil {
		logger.Error("add trade info error, error:", err.Error())
		return err
	}
	return nil
}

