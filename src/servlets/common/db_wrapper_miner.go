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
		Didi   int           `bson:"didi" json:"didi"`
		Sid    int           `bson:"sid" json:"sid"`
		Dn     string        `bson:"dn,omitempty" json:"dn"`
		OsVer  string        `bson:"os_ver,omitempty" json:"os_ver"`
		BindTs int64         `bson:"bind_ts,omitempty" json:"bind_ts"`
	}
	DtDeviceHistory struct {
		Id       bson.ObjectId `bson:"_id,omitempty" json:"-"`
		Uid      int64         `bson:"uid,omitempty" json:"uid"`
		Mid      int           `bson:"mid,omitempty" json:"mid"`
		Plat     int           `bson:"plat,omitempty" json:"plat"`
		Appid    int           `bson:"appid,omitempty" json:"appid"`
		Did      string        `bson:"did,omitempty" json:"-"`
		Dn       string        `bson:"dn,omitempty" json:"dn"`
		OsVer    string        `bson:"os_ver,omitempty" json:"os_ver"`
		BindTs   int64         `bson:"bind_ts,omitempty" json:"bind_ts"`
		UnbindTs int64         `bson:"unbind_ts,omitempty"`
		ForceUid int64         `bson:"force_uid,omitempty"`
	}

)

func (ddh *DtDeviceHistory) Build(dd *DtDevice) {
	ddh.UnbindTs = utils.GetTimestamp13()
	ddh.Id = dd.Id
	ddh.Uid = dd.Uid
	ddh.Mid = dd.Mid
	ddh.Plat = dd.Plat
	ddh.Appid = dd.Appid
	ddh.Did = dd.Did
	ddh.Dn = dd.Dn
	ddh.OsVer = dd.OsVer
	ddh.BindTs = dd.BindTs
}

func (ddh *DtDeviceHistory) BuildForceUnBind(dd *DtDevice, forceUid int64) {
	ddh.ForceUid = forceUid
	ddh.Build(dd)
}
