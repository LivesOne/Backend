package common

import "gopkg.in/mgo.v2/bson"

const (
	MSG_TYPE_ADD_CONTACT = 1
)

type (
	DtMessage struct {
		Id         bson.ObjectId `bosn:"id,omitempty" bosn:"id,omitempty"`
		To         int64         `bosn:"to,omitempty" bosn:"to,omitempty"`
		Type       int           `bosn:"type,omitempty" bosn:"type,omitempty"`
		Status     int           `bosn:"status,omitempty" bosn:"status,omitempty"`
		Ts         int64         `bosn:"ts,omitempty" bosn:"ts,omitempty"`
		NewContact *NewContact   `bosn:"new_contact,omitempty" bosn:"new_contact,omitempty"`
	}
	NewContact struct {
		Uid      int64  `bosn:"uid,omitempty" bosn:"uid,omitempty"`
		Nickname string `bosn:"nickname,omitempty" bosn:"nickname,omitempty"`
	}
)
