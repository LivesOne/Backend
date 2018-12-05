package common

import (
	"github.com/thanhpk/randstr"
	"gitlab.maxthon.net/cloud/livesone-micro-user/src/proto"
	"net/url"
	"servlets/rpc"
	"strings"
	"utils"
	"utils/config"
	"utils/logger"
)

const (
	QR_CODE_PRFIX_PAY                  = "upx"
	QR_CODE_PRFIX_TRANS                = "fkx"
	QR_CODE_PRFIX_UINFO                = "uix"
	QR_CODE_LEN                        = 20
	QR_CODE_EXPIRE                     = 180
	QR_CODE_RDS_PRFIX                  = "cache:qrcode:"
	QR_CODE_CACHE_FIELD_UID            = "uid"
	QR_CODE_CACHE_FIELD_NICKNAME       = "nickname"
	QR_CODE_CACHE_FIELD_TRANS_TYPE     = "type"
	QR_CODE_CACHE_FIELD_TRANS_TO       = "to"
	QR_CODE_CACHE_FIELD_TRANS_CURRENCY = "currency"
	QR_CODE_CACHE_FIELD_TRANS_AMOUNT   = "amount"
)

func BuildUserInfoQrCodeCache(uid string) (string, int) {
	qrCode := QR_CODE_PRFIX_UINFO + strings.ToLower(randstr.String(QR_CODE_LEN))
	key := buildQrCodeRdsKey(qrCode)
	t, e := QrCodeExists(key)
	if e != nil {
		logger.Error("get qrcode from redis error", e.Error())
		return "", 0
	}
	if t {
		return BuildUserInfoQrCodeCache(uid)
	}
	nickname, _ := rpc.GetUserField(utils.Str2Int64(uid), microuser.UserField_NICKNAME)
	cache := map[string]string{
		QR_CODE_CACHE_FIELD_UID:      uid,
		QR_CODE_CACHE_FIELD_NICKNAME: nickname,
	}
	hmset(key, cache)
	rdsExpire(key, QR_CODE_EXPIRE)
	return qrCode, QR_CODE_EXPIRE
}

func BuildTransQrCodeCache(transType, to, currency, amount string) (string, int) {
	qrCode := QR_CODE_PRFIX_TRANS + strings.ToLower(randstr.String(QR_CODE_LEN))
	key := buildQrCodeRdsKey(qrCode)
	t, e := QrCodeExists(key)
	if e != nil {
		logger.Error("get qrcode from redis error", e.Error())
		return "", 0
	}
	if t {
		return BuildTransQrCodeCache(transType, to, currency, amount)
	}
	cache := map[string]string{
		QR_CODE_CACHE_FIELD_TRANS_TYPE:     transType,
		QR_CODE_CACHE_FIELD_TRANS_TO:       to,
		QR_CODE_CACHE_FIELD_TRANS_CURRENCY: currency,
		QR_CODE_CACHE_FIELD_TRANS_AMOUNT:   amount,
	}
	hmset(key, cache)
	rdsExpire(key, QR_CODE_EXPIRE)
	return qrCode, QR_CODE_EXPIRE
}

func BuildQrCodeContent(qrCode string) string {
	u, _ := url.Parse(config.GetConfig().QrCodeContentUrl)
	u.Path += qrCode
	q := u.Query()
	q.Add("t", utils.Int642Str(utils.GetTimestamp13()))
	u.RawQuery = q.Encode()
	return u.String()
}

func QrCodeExists(qrCode string) (bool, error) {
	return rdsExist(buildQrCodeRdsKey(qrCode))
}

func buildQrCodeRdsKey(qrCode string) string {
	return QR_CODE_RDS_PRFIX + qrCode
}
