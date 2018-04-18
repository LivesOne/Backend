package common

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"servlets/constants"
	"utils"
	"utils/config"
	"utils/logger"
)

var tSession *mgo.Session
var txdbc config.MongoConfig

const (

	// MGDBPoolMax = 10
	PENDING  = "dt_pending"
	COMMITED = "dt_committed"
	FAILED   = "dt_failed"
)

func InitTxHistoryMongoDB() {
	config := config.GetConfig()
	txdbc = config.TxHistory
	connStr := fmt.Sprintf("%s?maxPoolSize=%d", txdbc.DBHost, txdbc.MaxConn)
	logger.Info("conn mongo db ---> ", connStr)
	tSession, _ = mgo.Dial(connStr)
	tSession.SetPoolLimit(txdbc.MaxConn)
}

func txCommonInsert(db, c string, p interface{}) error {
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

func txCommitDelete(db, c string, txid int64) error {
	session := tSession.Clone()
	defer session.Close()
	collection := session.DB(db).C(c)
	return collection.Remove(bson.M{"_id": txid})
}

func txCommonCheckExists(db, tb string, id interface{}) bool {
	session := tSession.Clone()
	defer session.Close()
	collection := session.DB(db).C(tb)
	c, e := collection.FindId(id).Count()
	if e != nil {
		logger.Error("query mongo err ", e.Error())
		return false
	}
	return c > 0
}

func InsertPending(pending *DTTXHistory) error {
	return txCommonInsert(txdbc.DBDatabase, PENDING, pending)
}

func InsertCommited(commited *DTTXHistory) error {
	return txCommonInsert(txdbc.DBDatabase, COMMITED, commited)
}
func InsertFailed(failed *DTTXHistory) error {
	return txCommonInsert(txdbc.DBDatabase, FAILED, failed)
}

func DeletePending(txid int64) error {
	logger.Info("DELETE PENDING :", FindPending(txid))
	return txCommitDelete(txdbc.DBDatabase, PENDING, txid)
}

func DeletePendingByInfo(tx *DTTXHistory) error {
	logger.Info("DELETE PENDING :", *tx)
	return txCommitDelete(txdbc.DBDatabase, PENDING, tx.Id)
}

func FindPending(txid int64) *DTTXHistory {
	session := tSession.Clone()
	defer session.Close()
	collection := session.DB(txdbc.DBDatabase).C(PENDING)
	res := DTTXHistory{}
	err := collection.FindId(txid).One(&res)
	if err != nil {
		logger.Error("query mongo db error ", err.Error())
		return nil
	}
	return &res
}

func CheckCommited(txid int64) bool {
	return txCommonCheckExists(txdbc.DBDatabase, COMMITED, txid)
}

func CheckPending(txid int64) bool {
	return txCommonCheckExists(txdbc.DBDatabase, PENDING, txid)
}

func FindAndModifyPending(txid, from, status int64) (*DTTXHistory, bool) {
	session := tSession.Clone()
	defer session.Close()
	coll := session.DB(txdbc.DBDatabase).C(PENDING)
	res := DTTXHistory{}
	query := bson.M{
		"_id":  txid,
		"from": from,
	}
	change := mgo.Change{
		Update: bson.M{
			"$bit": bson.M{
				"status": bson.M{
					"or": constants.TX_STATUS_COMMIT,
				},
			},
		},
		ReturnNew: false,
	}
	info, err := coll.Find(query).Apply(change, &res)
	if err != nil {
		logger.Error("findAndModify error ", err.Error())
	}
	f := true
	if info == nil || info.Matched == 0 {
		f = false
	}
	return &res, f
}

func ExistsPending(txid int64) bool {
	session := tSession.Clone()
	defer session.Close()
	collection := session.DB(txdbc.DBDatabase).C(PENDING)
	c, err := collection.FindId(txid).Count()
	if err != nil {
		logger.Error("query mongo db error ", err.Error())
		return false
	}
	return c > 0
}

func FindTopPending(query interface{}, top int) *DTTXHistory {
	session := tSession.Clone()
	defer session.Close()
	collection := session.DB(txdbc.DBDatabase).C(PENDING)
	var res DTTXHistory
	err := collection.Find(query).Sort("+_id").One(&res)
	if err != nil && err != mgo.ErrNotFound {
		logger.Error("query mongo error ", err.Error())
		return nil
	}
	return &res
}

func CheckDup(err error) bool {
	if err != nil {
		return mgo.IsDup(err)
	}
	return true
}

func QueryCommitted(query interface{}, limit int) []DTTXHistory {
	session := tSession.Clone()
	defer session.Close()
	logger.Debug("mongo query :", utils.ToJSONIndent(query))
	collection := session.DB(txdbc.DBDatabase).C(COMMITED)
	res := []DTTXHistory{}
	err := collection.Find(query).Sort("-_id").Limit(limit).All(&res)
	if err != nil && err != mgo.ErrNotFound {
		logger.Error("query mongo db error ", err.Error())
		return nil
	}
	logger.Debug("query res ", utils.ToJSONIndent(res))
	return res
}

func QueryCountMinerByTs(ts int64)int{
	session := tSession.Clone()
	defer session.Close()
	collection := session.DB(txdbc.DBDatabase).C(COMMITED)
	query := bson.M{
		"type":constants.TX_TYPE_REWARD,
		"_id":bson.M{
			"&gt":utils.TimestampToTxid(ts,0),
		},
	}
	count,err := collection.Find(query).Count()
	if err != nil {
		logger.Error("query mongo tx error",err.Error())
	}
	return count
}