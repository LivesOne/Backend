package accounts

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
)

// request "param"
type getImageParam struct {
	Type   int   `json:"type"`
	Number int   `json:"number"`
	Width  int   `json:"width"`
	Height int   `json:"height"`
	Expire int64 `json:"expire"`
}

// full get image request data
type getImageRequest struct {
	Base  common.BaseInfo `json:"base"`
	Param getImageParam   `json:"param"`
}

type responseGetImgVCode struct {
	IMG_id   string `json:"img_id"`
	IMG_size int    `json:"img_size"`
	IMG_data string `json:"img_data"`
	Expire   int64  `json:"expire"`
}

// // getImageVCode get image verification code
// type getImageVCode struct {
// 	Base  common.BaseInfo `json:"base"`
// 	Param string          `json:"param"`
// }

// getImgVCodeHandler
type getImgVCodeHandler struct {
	header     *common.HeaderParams // request header param
	getImgData *getImageRequest     // request data
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

	handler.header = common.ParseHttpHeaderParams(request)
	common.ParseHttpBodyParams(request, &handler.getImgData)
}
