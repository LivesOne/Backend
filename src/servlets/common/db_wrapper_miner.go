package common

import (
	"gopkg.in/mgo.v2/bson"
	"utils"
)

//"collection": "dt_device",
//{
//	"_id": ObjectId(5778879145ce3b47118d99af),
//	"uid": NumberLong(123456789),
//	"mid": NumberInt(3),
//	"plat": NumberInt(1),
//	"appid": NumberInt(1),
//	"did": "asdf12345ABCDEF",
//	"dn": "Thinkpad",
//	"os_ver": "windows10",
//	"bind_ts": NumberLong(1501234567890)
//}

type (
	DtDevice struct {
		Id     bson.ObjectId `bson:"_id,omitempty" json:"-"`
		Uid    int64         `bson:"uid,omitempty" json:"uid"`
		Mid    int           `bson:"mid,omitempty" json:"mid"`
		Plat   int           `bson:"plat,omitempty" json:"plat"`
		Appid  int           `bson:"appid,omitempty" json:"appid"`
		Did    string        `bson:"did,omitempty" json:"-"`
		Dn     string        `bson:"dn,omitempty" json:"dn"`
		OsVer  string        `bson:"os_ver,omitempty" json:"os_version"`
		BindTs int64         `bson:"bind_ts,omitempty" json:"bind_ts"`
	}
	DtDeviceHistory struct {
		DtDevice
		UnbindTs int64 `bson:"unbind_ts,omitempty"`
	}
)

func (ddh *DtDeviceHistory) Build(dd *DtDevice) {
	ddh.UnbindTs = utils.GetTimestamp13()
	ddh.DtDevice = *dd
}
