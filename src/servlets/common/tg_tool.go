package common

import (
	"utils"
	"utils/config"
	"utils/logger"
	"utils/lvthttp"
)

//valid protocol
//request body:
//{"lvt_uid":"123456789","code":"b36uf3"}
//response:
//{"rc":0,"msg":"ok","tg_id":"13810882398"}
//rc: 0 ok,1 not exists, 2 code error

type (
	tgReq struct {
		LvtUid string `json:"lvt"`
		Code   string `json:"code"`
	}
	tgResData struct {
		Lvt      string `json:"lvt"`
		Telegram string `json:"telegram"`
	}
	tgRes struct {
		Base *BaseResp  `json:"base"`
		Data *tgResData `json:"data"`
	}
)

func AuthTG(lvtUid, code string) (bool, *tgRes) {
	reqParam := &tgReq{
		LvtUid: lvtUid,
		Code:   code,
	}
	resStr, err := lvthttp.JsonPost(config.GetConfig().AuthTelegramUrl, reqParam)
	if err != nil {
		logger.Error("http request error", err.Error())
		return false, nil
	}
	res := new(tgRes)
	err = utils.FromJson(resStr, res)
	if err != nil {
		return false, nil
	}
	if res.Base == nil {
		return false, nil
	}
	return res.Base.RC == 0, res
}
