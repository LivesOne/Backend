package accounts

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"strings"
	log "utils/logger"
	"utils/config"
)

type imgParam struct {
	Type   int `json:"type,omitempty"`
	Length int `json:"length,omitempty"`
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
	Expire int `json:"expire,omitempty"`
}

type imgRequest struct {
	Base  common.BaseInfo `json:"base,omitempty"`
	Param imgParam        `json:"param,omitempty"`
}

type responseImg struct {
	ImgId   string `json:"img_id,omitempty"`
	ImgSize int    `json:"img_size,omitempty"`
	ImgData string `json:"img_data,omitempty"`
	Expire  int  `json:"expire,omitempty"`
}

type httpReqParam struct {
	Len    int `json:"len,omitempty"`
	W      int `json:"w,omitempty"`
	H      int `json:"h,omitempty"`
	Expire int `json:"expire,omitempty"`
}

type httpReqVCode struct {
	Expire int    `json:"expire,omitempty"`
	Size   int    `json:"size,omitempty"`
	Id     string `json:"id,omitempty"`
}

type httpReqVCodeData struct {
	ImgBase string `json:"imgBase,omitempty"`
	VCode *httpReqVCode	`json:"vCode,omitempty"`
}

type httpResParam struct {
	Ret int `json:"ret,omitempty"`
	Msg string `json:"msg,omitempty"`
	Data *httpReqVCodeData `json:"data,omitempty"`
}

// loginHandler implements the "Echo message" interface
type getImgVCodeHandler struct {

	//header      *common.HeaderParams // request header param
	//requestData *imgRequest    // request body

}

func (handler *getImgVCodeHandler) Method() string {
	return http.MethodPost
}

func (handler *getImgVCodeHandler) Handle(request *http.Request, writer http.ResponseWriter) {

	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	//handler.header = common.ParseHttpHeaderParams(request)

	params := imgRequest{}
	common.ParseHttpBodyParams(request, &params)

	typeData := httpReqParam{
		W:      params.Param.Width,
		H:      params.Param.Height,
		Len:    params.Param.Length,
		Expire: params.Param.Expire,
	}
	reqParam, _ := json.Marshal(typeData)
	url := config.GetConfig().ImgSvrAddr + "/img/v1/getCode"
	svrResStr, err := common.Post(url, string(reqParam))
	if err != nil {
		log.Error("url ---> ", url," http send error " ,err.Error())
		response.Base = &common.BaseResp{
			RC:  constants.RC_SYSTEM_ERR.Rc,
			Msg: constants.RC_SYSTEM_ERR.Msg,
		}
	} else {
		svrRes := httpResParam{}
		err := json.Unmarshal([]byte(svrResStr), &svrRes)
		if err != nil {
			log.Info("ParseHttpBodyParams, parse body param error: ", err)
			response.Base = &common.BaseResp{
				RC:  constants.RC_SYSTEM_ERR.Rc,
				Msg: constants.RC_SYSTEM_ERR.Msg,
			}
		}
		if svrRes.Ret == 0 {
			response.Data = &responseImg{
				ImgId:svrRes.Data.VCode.Id,
				ImgSize:svrRes.Data.VCode.Size,
				ImgData:svrRes.Data.ImgBase,
				Expire:svrRes.Data.VCode.Expire,
			}
		}
	}

}

