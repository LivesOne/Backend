package common

import "utils"

const (
	ASSET_LOCK_TYPE_NOR  = 0
	ASSET_LOCK_TYPE_DRAW = 1
)

type (
	Reward struct {
		Total      int64 `json:"total"`
		Yesterday  int64 `json:"yesterday"`
		Lastmodify int64 `json:"lastmodify"`
		Uid        int64 `json:"uid"`
		Days       int   `json:"days"`
	}

	AssetLock struct {
		Id       int64  `json:"-" bson:"id"`
		IdStr    string `json:"id" bson:"-"`
		Type     int    `json:"type" bson:"type"`
		Uid      int64  `json:"-" bson:"uid"`
		Value    string `json:"value" bson:"-"`
		Month    int    `json:"month" bson:"month"`
		Hashrate int    `json:"hashrate" bson:"hashrate"`
		Begin    int64  `json:"begin" bson:"begin"`
		End      int64  `json:"end" bson:"end"`
		ValueInt int64  `json:"-" bson:"value"`
	}

	UserWithdrawalQuota struct {
		Day       int64 `json:"day"`
		Month     int64 `json:"month"`
		Casual    int64 `json:"casual"`
		DayExpend int64 `json:"dayExpend"`
	}

	EthTxHistyr struct {
		Txid    int64  `json:"txid"`
		Type    int    `json:"type"`
		TradeNo string `json:"trade_no"`
		From    int64  `json:"from"`
		To      int64  `json:"to"`
		Value   int64  `json:"value"`
		Ts      int64  `json:"ts"`
	}
	TradePending struct {
		TradeNo    string `json:"trade_no"`
		Uid        int64  `json:"uid"`
		BizContent string `json:"biz_content"`
		Ts         int64  `json:"ts"`
	}
)

func (al *AssetLock) IsOk() bool {
	return al.Month > 0 &&
		al.ValueInt > 0 &&
		al.End > utils.GetTimestamp13()

}
