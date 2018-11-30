package common

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"utils/config"
	"utils/logger"
)

const (
	DT_CONTACTS = "dt_contacts"
)

var cSession *mgo.Session
var cConfig config.MongoConfig

func InitContactsMongoDB() {
	cConfig = config.GetConfig().Contacts
	cSession = mgoConn(cConfig)
}

func GetContactsListByUid(uid int64) []DtContacts {
	session := cSession.Clone()
	defer session.Close()
	collection := session.DB(cConfig.DBDatabase).C(DT_CONTACTS)
	res := []DtContacts{}
	err := collection.Find(bson.M{"uid": uid}).All(&res)
	if err != nil && err != mgo.ErrNotFound {
		logger.Error("query user contacts error", err.Error())
		return nil
	}
	return res
}

func CreateContact(p map[string]interface{}) error {
	return mgoCommonInsert(cSession, cConfig.DBDatabase, DT_CONTACTS, p)
}

func ModifyContact(p map[string]interface{}, uid, contactId int64) error {
	session := cSession.Clone()
	defer session.Close()
	collection := session.DB(cConfig.DBDatabase).C(DT_CONTACTS)
	selector := bson.M{"uid": uid, "contact_id": contactId}
	md := bson.M{"$set": p}
	return collection.Update(selector, md)
}

func DeleteContact(uid int64, contactId int64) error {
	session := tSession.Clone()
	defer session.Close()
	collection := session.DB(cConfig.DBDatabase).C(DT_CONTACTS)
	selector := bson.M{"uid": uid, "contact_id": contactId}
	return collection.Remove(selector)
}
