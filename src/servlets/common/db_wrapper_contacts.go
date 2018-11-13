package common

import (
	"gopkg.in/mgo.v2/bson"
)

type (

	//"contact_id": 1,
	//"name": "张三",
	//"email": "zs@abc.com",
	//"country": 86,
	//"phone": "18888888888",
	//"livesone_uid": NumberLong(222222222),
	//"wallet_addrss": "0x988cbccfc7e26407b191282891b96c308de79947",

	DtContacts struct {
		Id            bson.ObjectId `json:"-" bson:"_id"`
		Uid           int64         `json:"-" bson:"uid,omitempty"`
		ContactId     int           `json:"contact_id,omitempty" bson:"contact_id,omitempty"`
		Name          string        `json:"name,omitempty" bson:"name,omitempty"`
		Remark          string        `json:"remark,omitempty" bson:"remark,omitempty"`
		Email         string        `json:"email,omitempty" bson:"email,omitempty"`
		Country       int           `json:"country,omitempty" bson:"country,omitempty"`
		Phone         string        `json:"phone,omitempty" bson:"phone,omitempty"`
		LivesoneUid   string        `json:"livesone_uid,omitempty" bson:"livesone_uid,omitempty"`
		WalletAddress string        `json:"wallet_address,omitempty" bson:"wallet_address,omitempty"`
		UpdateTime    int64         `json:"update_time,omitempty" bson:"update_time,omitempty"`
		CreateTime    int64         `json:"create_time,omitempty" bson:"create_time,omitempty"`
	}
)
