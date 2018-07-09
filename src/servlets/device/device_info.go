package device

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils/logger"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)





type deviceInfoRequest struct {
	Base  *common.BaseInfo `json:"base"`
}

// sendVCodeHandler
type deviceInfoHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *deviceInfoHandler) Method() string {
	return http.MethodPost
}

func (handler *deviceInfoHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true)
	defer log.InfoAll()
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := new(deviceInfoRequest) // request body

	if !common.ParseHttpBodyParams(request, requestData) {
		response.SetResponseBase(constants.RC_PROTOCOL_ERR)
		return
	}

	//校验参数
	if requestData.Base == nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	app,device := requestData.Base.App,requestData.Base.Device
	if app == nil || device == nil || !app.IsValid(){
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	//组装查询
	query := bson.M{
		"appid":app.AppID,
		"did":device.DID,
	}
	deviceInfo,err := common.QueryDevice(query)
	if err != nil && err != mgo.ErrNotFound {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}
	response.Data = deviceInfo

}
