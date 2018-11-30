package common

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"utils"
	"utils/config"
	"utils/logger"
)

const (
	DT_MESSAGE = "dt_message"
)

var msgSession *mgo.Session
var msgConfig config.MongoConfig

func InitMsgMongoDB() {
	config := config.GetConfig()
	msgConfig = config.Msg
	msgSession = mgoConn(msgConfig)
}

func AddMsg(msg *DtMessage) error {
	return mgoCommonInsert(msgSession, msgConfig.DBDatabase, DT_MESSAGE, msg)
}

func GetMsgByUidAndType(uid int64, mtype int) []DtMessage {
	session := msgSession.Clone()
	defer session.Close()
	collection := session.DB(msgConfig.DBDatabase).C(DT_MESSAGE)
	res := []DtMessage{}
	query := bson.M{"to": uid}
	if mtype > 0 {
		query["type"] = mtype
	}
	err := collection.Find(query).Sort("-ts").All(&res)
	if err != nil && err != mgo.ErrNotFound {
		logger.Error("get msg by ", utils.ToJSON(query),"query mongo db error ", err.Error())
		return nil
	}
	logger.Debug("get msg by ", utils.ToJSON(query)," res ", utils.ToJSON(res))
	return res
}

func DelReadMsg(ids []bson.ObjectId) error {
	if len(ids) > 0 {
		session := msgSession.Clone()
		defer session.Close()
		collection := session.DB(tradeConfig.DBDatabase).C(DT_MESSAGE)

		orlist := make([]bson.M, len(ids))
		for i, v := range ids {
			orlist[i] = bson.M{
				"_id": v,
			}
		}

		selector := bson.M{
			"$or": orlist,
		}
		logger.Debug("del msg by ", utils.ToJSON(selector))
		return collection.Remove(selector)
	}
	return nil
}
