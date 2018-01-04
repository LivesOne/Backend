package common

import (
	"gopkg.in/mgo.v2"
	"utils/config"
	"fmt"
	"utils/logger"
	"gopkg.in/mgo.v2/bson"
)

var tSession *mgo.Session
var txdbc config.DBConfig

const (

	MGDBPoolMax = 10
	PENDING = "dt_pending"
	COMMITED = "dt_committed"
)

func InitTxHistoryMongoDB() {
	config := config.GetConfig()
	txdbc = config.TxHistory
	connStr := fmt.Sprintf("%s?maxPoolSize=%d", txdbc.DBHost, MGDBPoolMax)
	logger.Info("conn mongo db ---> ", connStr)
	tSession, _ = mgo.Dial(connStr)
	tSession.SetPoolLimit(MGDBPoolMax)
}



func txCommonInsert(db, c string, p interface{}) error {
	session := tSession.Clone()
	defer session.Close()
	collection := session.DB(db).C(c)
	err := collection.Insert(p)
	if err != nil {
		logger.Error("mongo_base method:Insert ", err.Error())
	}
	return err
}

func txCommitDelete(db,c string,txid int64)error{
	session := tSession.Clone()
	defer session.Close()
	collection := session.DB(db).C(c)
	return collection.Remove(bson.M{"_id":txid})
}



func InsertPending(pending *DTTXHistory) error {
	return txCommonInsert(txdbc.DBDatabase, PENDING, pending)
}

func InsertCommited(commited *DTTXHistory) error {
	return txCommonInsert(txdbc.DBDatabase, COMMITED, commited)
}

func DeletePending(txid int64)error{
	return txCommitDelete(txdbc.DBDatabase,PENDING,txid)
}

func FindPending(txid int64)*DTTXHistory{
	session := tSession.Clone()
	defer session.Close()
	collection := session.DB(txdbc.DBDatabase).C(PENDING)
	res := DTTXHistory{}
	err := collection.FindId(txid).One(&res)
	if err != nil {
		logger.Error("query mongo db error ",err.Error())
		return nil
	}
	return &res
}

func FindAndModify(txid int64,updDoc *DTTXHistory)*DTTXHistory{
	session := tSession.Clone()
	defer session.Close()
	db := session.DB(txdbc.DBDatabase)
	res := DTTXHistory{}
	cmd := bson.D{
		bson.DocElem{
			Name:"findAndModify",
			Value:PENDING,
		},
		bson.DocElem{
			Name:"query",
			Value:bson.M{"_id":txid},
		},
		bson.DocElem{
			Name:"update",
			Value:updDoc,
		},
	}
	db.Run(cmd,&res)
	return &res
}



