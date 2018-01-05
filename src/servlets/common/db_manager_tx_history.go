package common

import (
	"gopkg.in/mgo.v2"
	"utils/config"
	"fmt"
	"utils/logger"
	"gopkg.in/mgo.v2/bson"
	"servlets/constants"
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

func txCommonCheckExists(db,tb string,id interface{})bool{
	session := tSession.Clone()
	defer session.Close()
	collection := session.DB(db).C(tb)
	c,e := collection.FindId(id).Count()
	if e != nil {
		logger.Error("query mongo err ",e.Error())
		return false
	}
	return c>0
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

func CheckCommited(txid int64)bool{
	return txCommonCheckExists(txdbc.DBDatabase,COMMITED,txid)
}

func CheckPending(txid int64)bool{
	return txCommonCheckExists(txdbc.DBDatabase,PENDING,txid)
}


func FindAndModifyPending(txid,from,status int64)(*DTTXHistory){
	session := tSession.Clone()
	defer session.Close()
	coll := session.DB(txdbc.DBDatabase).C(PENDING)
	res := DTTXHistory{}
	query := bson.M{
		"_id":txid,
		"from":from,
		"status":bson.M{
			"$bitsAllClear": []int{0}  ,
		},
	}
	change := mgo.Change{
		Update: bson.M{
			"$bit":bson.M{
				"status":bson.M{
					"and":constants.TX_STATUS_COMMIT,
				},
			},
		},
		ReturnNew: false,
	}
	_,err := coll.Find(query).Apply(change,&res)
	if err!=nil {
		logger.Error("findAndModify error ",err.Error())
	}
	//logger.Info("findAndModify matched doc record num ",info.Matched)
	return &res
}


