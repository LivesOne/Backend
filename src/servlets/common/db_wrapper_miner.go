package common

import "gopkg.in/mgo.v2/bson"

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
		Id     bson.ObjectId `bson:"_id,omitempty"`
		Uid    int64         `bson:"uid,omitempty"`
		Mid    int           `bson:"mid,omitempty"`
		Plat   int           `bson:"plat,omitempty"`
		Appid  int           `bson:"appid,omitempty"`
		Did    string        `bson:"did,omitempty"`
		Dn     string        `bson:"dn,omitempty"`
		OsVer  string        `bson:"os_ver,omitempty"`
		BindTs int64         `bson:"bind_ts,omitempty"`
	}
	DtDeviceHistory struct {
		DtDevice
		UnbindTs int64 `bson:"unbind_ts,omitempty"`
	}
)

func (ddh *DtDeviceHistory) Build(dd *DtDevice) {
	ddh.DtDevice = *dd
}
