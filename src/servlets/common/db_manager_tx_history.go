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
var ntSession *mgo.Session
var ntxdbc config.MongoConfig

const (

	// MGDBPoolMax = 10
	PENDING  = "dt_pending"
	COMMITED = "dt_committed"
	FAILED   = "dt_failed"
)

func InitTxHistoryMongoDB() {
	initOldTxHistoryMongoDB()
	initNewTxHistoryMongoDB()
}

func initOldTxHistoryMongoDB(){
	config := config.GetConfig()
	txdbc = config.TxHistory
	connStr := fmt.Sprintf("%s?maxPoolSize=%d", txdbc.DBHost, txdbc.MaxConn)
	logger.Info("conn mongo db ---> ", connStr)
	tSession, _ = mgo.Dial(connStr)
	tSession.SetPoolLimit(txdbc.MaxConn)
}

func initNewTxHistoryMongoDB(){
	config := config.GetConfig()
	ntxdbc = config.NewTxHistory
	connStr := fmt.Sprintf("%s?maxPoolSize=%d", ntxdbc.DBHost, ntxdbc.MaxConn)
	logger.Info("conn mongo db ---> ", connStr)
	ntSession, _ = mgo.Dial(connStr)
	ntSession.SetPoolLimit(ntxdbc.MaxConn)
}


func txCommonInsert(cs *mgo.Session,db, c string, p interface{}) error {
	session := cs.Clone()
	defer session.Close()
	session.SetSafe(&mgo.Safe{WMode: "majority"})
	collection := session.DB(db).C(c)
	err := collection.Insert(p)
	if err != nil {
		logger.Error("mongo_base method:Insert ", err.Error())
	}
	return err
}


func txCommitDelete(cs *mgo.Session,db, c string, txid int64) error {
	session := tSession.Clone()
	defer session.Close()
	collection := session.DB(db).C(c)
	return collection.Remove(bson.M{"_id": txid})
}

func txCommonCheckExists(cs *mgo.Session,db, tb string, id interface{}) bool {
	session := cs.Clone()
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
	return txCommonInsert(tSession,txdbc.DBDatabase, PENDING, pending)
}


func InsertCommited(commited *DTTXHistory) error {
	return txCommonInsert(tSession,txdbc.DBDatabase, COMMITED, commited)
}
func InsertFailed(failed *DTTXHistory) error {
	return txCommonInsert(tSession,txdbc.DBDatabase, FAILED, failed)
}
func InsertLVTCFailed(failed *DTTXHistory) error {
	return txCommonInsert(ntSession,ntxdbc.DBDatabase, FAILED, failed)
}

func InsertLVTCPending(pending *DTTXHistory) error {
	return txCommonInsert(ntSession,ntxdbc.DBDatabase, PENDING, pending)
}

func InsertLVTCCommited(commited *DTTXHistory) error {
	return txCommonInsert(ntSession,ntxdbc.DBDatabase, COMMITED, commited)
}

func DeletePending(txid int64) error {
	logger.Info("DELETE PENDING :", FindPending(txid))
	return txCommitDelete(tSession,txdbc.DBDatabase, PENDING, txid)
}

func DeletePendingByInfo(tx *DTTXHistory) error {
	logger.Info("DELETE PENDING :", *tx)
	return txCommitDelete(tSession,txdbc.DBDatabase, PENDING, tx.Id)
}

func DeleteLVTCPendingByInfo(tx *DTTXHistory) error {
	logger.Info("DELETE PENDING :", *tx)
	return txCommitDelete(ntSession,ntxdbc.DBDatabase, PENDING, tx.Id)
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

func FindLVTCPending(txid int64) *DTTXHistory {
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
	return txCommonCheckExists(tSession,txdbc.DBDatabase, COMMITED, txid)
}

func CheckPending(txid int64) bool {
	return txCommonCheckExists(tSession,txdbc.DBDatabase, PENDING, txid)
}

func CheckLVTCPending(txid int64) bool {
	return txCommonCheckExists(ntSession,ntxdbc.DBDatabase, PENDING, txid)
}

func CheckLVTCCommited(txid int64) bool {
	return txCommonCheckExists(ntSession,ntxdbc.DBDatabase, COMMITED, txid)
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

func FindAndModifyLVTCPending(txid, from, status int64) (*DTTXHistory, bool) {
	session := ntSession.Clone()
	defer session.Close()
	coll := session.DB(ntxdbc.DBDatabase).C(PENDING)
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

func ExistsLVTCPending(txid int64) bool {
	session := ntSession.Clone()
	defer session.Close()
	collection := session.DB(ntxdbc.DBDatabase).C(PENDING)
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

func FindTopLVTCPending(query interface{}, top int) *DTTXHistory {
	session := ntSession.Clone()
	defer session.Close()
	collection := session.DB(ntxdbc.DBDatabase).C(PENDING)
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

func QueryLVTCCommitted(query interface{}, limit int) []DTTXHistory {
	session := ntSession.Clone()
	defer session.Close()
	logger.Debug("mongo query :", utils.ToJSONIndent(query))
	collection := session.DB(ntxdbc.DBDatabase).C(COMMITED)
	res := []DTTXHistory{}
	err := collection.Find(query).Sort("-_id").Limit(limit).All(&res)
	if err != nil && err != mgo.ErrNotFound {
		logger.Error("query mongo db error ", err.Error())
		return nil
	}
	logger.Debug("query res ", utils.ToJSONIndent(res))
	return res
}

