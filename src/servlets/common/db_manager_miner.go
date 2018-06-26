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
	mSession, err = mgo.Dial(connStr)
	if err != nil {
		logger.Error("connect failed ", err.Error())
		return
	}
	mSession.SetPoolLimit(minerdbc.MaxConn)
}

func minerCommonInsert(db, c string, p ...interface{}) error {
	session := mSession.Clone()
	defer session.Close()
	session.SetSafe(&mgo.Safe{WMode: "majority"})
	collection := session.DB(db).C(c)
	err := collection.Insert(p...)
	if err != nil {
		logger.Error("mongo_base method:Insert ", err.Error())
	}
	return err
}

func minerCommitDelete(db, c string, id bson.ObjectId) error {
	session := mSession.Clone()
	defer session.Close()
	collection := session.DB(db).C(c)
	return collection.RemoveId(id)
}

func QueryMinerBindDevice(query bson.M) ([]DtDevice, error) {
	session := mSession.Clone()
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
	session := mSession.Clone()
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

func QueryDevice(uid int64,mid int,did string) (*DtDevice ,error){
	session := mSession.Clone()
	defer session.Close()
	collection := session.DB(minerdbc.DBDatabase).C(DT_DEVICE)
	res := new(DtDevice)
	err := collection.Find(bson.M{"uid":uid,"did":did,"mid":mid}).One(res)
	if err != nil {
		logger.Error("query mongo db error", err.Error())
		return nil, err
	}
	return res, nil
}
func QueryAllDevice(uid int64,mid int) ([]DtDevice ,error){
	session := mSession.Clone()
	defer session.Close()
	collection := session.DB(minerdbc.DBDatabase).C(DT_DEVICE)
	res := []DtDevice{}
	err := collection.Find(bson.M{"uid":uid,"mid":mid}).All(&res)
	if err != nil {
		logger.Error("query mongo db error", err.Error())
		return nil, err
	}
	return res, nil
}

func DeleteDevice(uid int64,mid,appid int,did string) error{
	session := mSession.Clone()
	defer session.Close()
	collection := session.DB(minerdbc.DBDatabase).C(DT_DEVICE)
	return collection.Remove(bson.M{"uid":uid,"did":did,"mid":mid,"appid":appid})
}
func DeleteAllDevice(uid int64,mid,appid int) error{
	session := mSession.Clone()
	defer session.Close()
	collection := session.DB(minerdbc.DBDatabase).C(DT_DEVICE)
	return collection.Remove(bson.M{"uid":uid,"mid":mid,"appid":appid})
}

func InsertDeviceBindHistory(device *DtDevice) error {
	ddh := new(DtDeviceHistory)
	ddh.Build(device)
	return minerCommonInsert(minerdbc.DBDatabase, DT_DEVICE_HISTORY, ddh)
}

func InsertAllDeviceBindHistory(device []DtDevice) error {
	adds := make([]*DtDeviceHistory,0)
	for _,v := range device {
		ddh := new(DtDeviceHistory)
		ddh.Build(&v)
		adds = append(adds,ddh)
	}
	return minerCommonInsert(minerdbc.DBDatabase, DT_DEVICE_HISTORY, adds)
}


func QueryDeviceByDid(did string) (*DtDevice ,error){
	session := mSession.Clone()
	defer session.Close()
	collection := session.DB(minerdbc.DBDatabase).C(DT_DEVICE)
	res := new(DtDevice)
	err := collection.Find(bson.M{"did":did}).One(res)
	if err != nil {
		logger.Error("query mongo db error", err.Error())
		return nil, err
	}
	return res, nil
}

func GetLastUnbindDeviceTs(uid int64,mid int) (int64 ,error){
	session := mSession.Clone()
	defer session.Close()
	collection := session.DB(minerdbc.DBDatabase).C(DT_DEVICE_HISTORY)
	res := new(DtDeviceHistory)
	err := collection.Find(bson.M{"uid":uid,"mid":mid}).Sort("-unbind_ts").Limit(1).One(res)
	if err != nil {
		logger.Error("query mongo db error", err.Error())
		return 0, err
	}
	return res.UnbindTs, nil
}


func QueryUserAllDevice(uid int64) ([]DtDevice ,error){
	session := mSession.Clone()
	defer session.Close()
	collection := session.DB(minerdbc.DBDatabase).C(DT_DEVICE)
	res := []DtDevice{}
	err := collection.Find(bson.M{"uid":uid}).All(&res)
	if err != nil {
		logger.Error("query mongo db error", err.Error())
		return nil, err
	}
	return res, nil
}