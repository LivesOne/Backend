package common

type (
	Reward struct {
		Total      int64 `json:"total"`
		Yesterday  int64 `json:"yesterday"`
		Lastmodify int64 `json:"lastmodify"`
		Uid        int64 `json:"uid"`
	}

	AssetLock struct {
		Id       int64  `json:"id" bson:"id"`
		Uid      int64  `json:"uid" bson:"uid"`
		Value    string `json:"value" bson:"value"`
		Month    int    `json:"month" bson:"month"`
		Hashrate int    `json:"hashrate" bson:"hashrate"`
		Begin    int64  `json:"begin" bson:"begin"`
		End      int64  `json:"end" bson:"end"`
		ValueInt int64  `json:"-"`
	}
)
