package common

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"utils/config"
	"utils/logger"
)

var mSession *mgo.Session
var minerdbc config.MongoConfig

const (

	// MGDBPoolMax = 10
	DT_DEVICE         = "dt_device"
	DT_DEVICE_HISTORY = "dt_device_history"
)

func InitMinerRMongoDB() {
	config := config.GetConfig()
	minerdbc = config.Miner
	connStr := fmt.Sprintf("%s?maxPoolSize=%d", minerdbc.DBHost, minerdbc.MaxConn)
	logger.Info("conn mongo db ---> ", connStr)
	var err error
	tSession, err = mgo.Dial(connStr)
	if err != nil {
		logger.Error("connect failed ", err.Error())
		return
	}
	tSession.SetPoolLimit(minerdbc.MaxConn)
}

func minerCommonInsert(db, c string, p interface{}) error {
	session := tSession.Clone()
	defer session.Close()
	session.SetSafe(&mgo.Safe{WMode: "majority"})
	collection := session.DB(db).C(c)
	err := collection.Insert(p)
	if err != nil {
		logger.Error("mongo_base method:Insert ", err.Error())
	}
	return err
}

func minerCommitDelete(db, c string, id bson.ObjectId) error {
	session := tSession.Clone()
	defer session.Close()
	collection := session.DB(db).C(c)
	return collection.RemoveId(id)
}

func QueryMinerBindDevice(query bson.M) ([]DtDevice, error) {
	session := tSession.Clone()
	defer session.Close()
	collection := session.DB(minerdbc.DBDatabase).C(DT_DEVICE)
	res := []DtDevice{}
	err := collection.Find(query).All(&res)
	if err != nil {
		logger.Error("query mongo db error", err.Error())
		return nil, err
	}
	return res, nil
}

func QueryMinerBindDeviceCount(query bson.M) (int, error) {
	session := tSession.Clone()
	defer session.Close()
	collection := session.DB(minerdbc.DBDatabase).C(DT_DEVICE)
	count, err := collection.Find(query).Count()
	if err != nil {
		logger.Error("query mongo db error", err.Error())
		return 0, err
	}
	return count, nil
}

func InsertDeviceBind(device *DtDevice) error {
	return minerCommonInsert(minerdbc.DBDatabase, DT_DEVICE, device)
}
