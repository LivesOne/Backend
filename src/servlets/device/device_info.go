package device

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"utils/logger"
	"gopkg.in/mgo.v2"
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

	device,err := common.QueryDeviceByDid(requestData.Base.Device.DID)
	if err != nil && err != mgo.ErrNotFound {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	response.Data = device

}
