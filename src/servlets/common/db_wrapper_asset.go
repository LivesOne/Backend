package common

import "utils"

type (
	Reward struct {
		Total      int64 `json:"total"`
		Yesterday  int64 `json:"yesterday"`
		Lastmodify int64 `json:"lastmodify"`
		Uid        int64 `json:"uid"`
	}

	AssetLock struct {
		Id       int64  `json:"-" bson:"id"`
		IdStr string 	`json:"id" bson:"-"`
		Uid      int64  `json:"uid" bson:"uid"`
		Value    string `json:"value" bson:"-"`
		Month    int    `json:"month" bson:"month"`
		Hashrate int    `json:"hashrate" bson:"hashrate"`
		Begin    int64  `json:"begin" bson:"begin"`
		End      int64  `json:"end" bson:"end"`
		ValueInt int64  `json:"-" bson:"value"`
	}
)


func (al *AssetLock)IsOk()bool{
	return al.Month>0&&
		   al.ValueInt >0 &&
		   al.End > utils.GetTimestamp13()

}