package common

import (
	"utils"
	"utils/lvthttp"
	"utils/config"
)

//valid protocol
//request body:
//{"lvt_uid":"123456789","code":"b36uf3"}
//response:
//{"rc":0,"msg":"ok","tg_id":"13810882398"}
//rc: 0 ok,1 not exists, 2 code error

type (
	tgReq struct {
		LvtUid string `json:"lvt_uid"`
		Code   string `json:"code"`
	}
	tgRes struct {
		Rc   int    `json:"rc"`
		Msg  string `json:"msg"`
		TgId string `json:"tg_id"`
	}
)

func AuthTG(lvtUid, code string) (bool, *tgRes) {
	reqParam := &tgReq{
		LvtUid: lvtUid,
		Code:   code,
	}
	resStr, err := lvthttp.JsonPost(config.GetConfig().AuthTelegramUrl, reqParam)
	if err != nil {
		return false, nil
	}
	res := new(tgRes)
	err = utils.FromJson(resStr, res)
	if err != nil {
		return false, nil
	}
	return res.Rc == 0, res
}
