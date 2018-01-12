package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils/vcode"
	"utils/logger"
)

type imgParam struct {
	Type   int `json:"type,omitempty"`
	Length int `json:"length,omitempty"`
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
	Expire int `json:"expire,omitempty"`
}

type imgRequest struct {
	Base  *common.BaseInfo `json:"base,omitempty"`
	Param *imgParam        `json:"param,omitempty"`
}

type responseImg struct {
	ImgId   string `json:"img_id,omitempty"`
	ImgSize int    `json:"img_size,omitempty"`
	ImgData string `json:"img_data,omitempty"`
	Expire  int    `json:"expire,omitempty"`
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

	header := common.ParseHttpHeaderParams(request)

	params := imgRequest{}
	common.ParseHttpBodyParams(request, &params)
	if params.Base == nil || params.Param == nil || 
		(handler.checkRequestParams(header, &params) == false) {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	vcodeRes := vcode.GetImgVCode(params.Param.Width,
		params.Param.Height,
		params.Param.Length,
		params.Param.Expire)

	if vcodeRes != nil && vcodeRes.Ret == 0 {
		response.Data = &responseImg{
			ImgId:   vcodeRes.Data.VCode.Id,
			ImgSize: vcodeRes.Data.VCode.Size,
			ImgData: vcodeRes.Data.ImgBase,
			Expire:  vcodeRes.Data.VCode.Expire,
		}
	} else {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
	}

}


func (handler *getImgVCodeHandler) checkRequestParams(header *common.HeaderParams, data *imgRequest) bool {
	if header == nil || (data == nil) {
		return false
	}

	if (header.IsValidTimestamp() == false)  {
		logger.Info("get image verify code: some header param missed")
		return false
	}

	if (data.Base.App == nil) || (data.Base.App.IsValid() == false) {
		logger.Info("get image verify code: app info invalid")
		return false
	}

	return true
}
