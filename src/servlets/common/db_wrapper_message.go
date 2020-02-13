package common

import "gopkg.in/mgo.v2/bson"

const (
	MSG_TYPE_ADD_CONTACT = 1
)

type (
	DtMessage struct {
		Id         bson.ObjectId `json:"_id,omitempty" bson:"_id,omitempty"`
		To         int64         `json:"-" bson:"to,omitempty"`
		Type       int           `json:"type,omitempty" bson:"type,omitempty"`
		Status     int           `json:"status" bson:"status,omitempty"`
		Ts         int64         `json:"ts,omitempty" bson:"ts,omitempty"`
		NewContact *NewContact   `json:"new_contact,omitempty" bson:"new_contact,omitempty"`
	}
	NewContact struct {
		Uid      int64  `json:"uid,omitempty" bson:"uid,omitempty"`
		Nickname string `json:"nickname,omitempty" bson:"nickname,omitempty"`
	}
)
