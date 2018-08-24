package common

import (
	"utils"
)

const (
	ASSET_LOCK_TYPE_NOR  = 0
	ASSET_LOCK_TYPE_DRAW = 1
	CURRENCY_LVT         = "LVT"
	CURRENCY_ETH         = "ETH"
	CURRENCY_LVTC        = "LVTC"
	ASSET_INCOME_MINING = 1
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

	AssetLockLvtc struct {
		Id          int64  `json:"-" bson:"id"`
		IdStr       string `json:"id" bson:"-"`
		Uid         int64  `json:"-" bson:"uid"`
		Value       string `json:"value" bson:"-"`
		Month       int    `json:"month" bson:"month"`
		Hashrate    int    `json:"hashrate" bson:"hashrate"`
		Begin       int64  `json:"begin" bson:"begin"`
		End         int64  `json:"end" bson:"end"`
		ValueInt    int64  `json:"-" bson:"value"`
		Currency    string `json:"currency" bson:"currency"`
		AllowUnlock int    `json:"allow_unlock" bson:"allow_unlock"`
		Income int    `json:"-" bson:"income,omitempty"`
	}

	UserWithdrawalQuota struct {
		Day       int64 `json:"day"`
		Month     int64 `json:"month"`
		Casual    int64 `json:"casual"`
		DayExpend int64 `json:"day_expend"`
		LastLevel int   `json:"last_level"`
	}

	EthTxHistory struct {
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
		From        int64  `json:"from"`
		To        int64  `json:"to"`
		BizContent string `json:"biz_content"`
		Ts         int64  `json:"ts"`
		Value      int64  `json:"-"`
		ValueStr   string `json:"value"`
		Type       int    `json:"type"`
	}
	UserWithdrawalCardUse struct {
		Txid       int64  `json:"txid"`
		TradeNo    string `json:"trade_no"`
		Uid        int64  `json:"uid"`
		Quota      int64  `json:"-"`
		QuotaStr   string `json:"quota"`
		Cost       int64  `json:"-"`
		CostStr    string `json:"cost"`
		CreateTime int64  `json:"create_time"`
		Type       int    `json:"type"`
		Currency   string `json:"currency"`
	}

	UserWithdrawalRequest struct {
		Id            int64  `json:"id"`
		TradeNo       string `json:"trade_no"`
		Uid           int64  `json:"uid"`
		Address       string `json:"address"`
		Value         int64  `json:"value"`
		Currency      string `json:"currency"`
		Fee           int64  `json:"currency"`
		FeeCurrency   string `json:"currency"`
		Txid          int64  `json:"txid"`
		TxidReturn    int64  `json:"txid_return"`
		TxidFee       int64  `json:"txid_eth"`
		TxidFeeReturn int64  `json:"txid_eth_return"`
		CreateTime    int64  `json:"create_time"`
		UpdateTime    int64  `json:"update_time"`
		Status        int    `json:"status"`
	}

	UserWithdrawCard struct {
		Id         int64  `json:"id"`
		Password   string `json:"password"`
		TradeNo    string `json:"trade_no"`
		OwnerUid   int64  `json:"owner_uid"`
		Quota      int64  `json:"quota"`
		CreateTime int64  `json:"create_time"`
		ExpireTime int64  `json:"expire_time"`
		Cost       int64  `json:"cost"`
		GetTime    int64  `json:"get_time"`
		UseTime    int64  `json:"use_time"`
		Status     int    `json:"status"`
	}

	TransBizContent struct {
	FeeCurrency string `json:"fee_currency"`
	Fee         int64 `json:"fee"`
	Remark      string `json:"remark"`
	}

)

func (al *AssetLock) IsOk() bool {
	return al.Month > 0 &&
		al.ValueInt > 0 &&
		al.End > utils.GetTimestamp13()

}

func (al *AssetLockLvtc) IsOk() bool {
	return al.Month > 0 &&
		al.ValueInt > 0 &&
		al.End > utils.GetTimestamp13()

}
