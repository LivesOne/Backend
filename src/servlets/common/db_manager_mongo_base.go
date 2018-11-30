package common

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"utils/config"
	"utils/logger"
)

var (
	sessionSafe = &mgo.Safe{WMode: "majority"}
)


func mgoConn(c config.MongoConfig)*mgo.Session{
	connStr := fmt.Sprintf("%s?maxPoolSize=%d", c.DBHost, c.MaxConn)
	logger.Info("conn mongo db ---> ", connStr)
	session, err := mgo.Dial(connStr)
	if err != nil {
		logger.Error("conn mongodb error",err.Error())
		panic(err)
	}
	session.SetPoolLimit(c.MaxConn)
	return session
}



func mgoCommonInsert(cs *mgo.Session, db, c string, p interface{}) error {
	session := cs.Clone()
	defer session.Close()
	session.SetSafe(sessionSafe)
	collection := session.DB(db).C(c)
	err := collection.Insert(p)
	if err != nil {
		logger.Error("mongo_base method:Insert ", err.Error())
	}
	return err
}

func mgoCommonDelete(cs *mgo.Session, db, c string, id interface{}) error {
	session := cs.Clone()
	defer session.Close()
	collection := session.DB(db).C(c)
	return collection.RemoveId(id)
}

func mgoCommonCheckExists(cs *mgo.Session, db, tb string, id interface{}) bool {
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
