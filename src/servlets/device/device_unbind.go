package device

import (
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
)

type deviceUnBindParam struct {
	Mid       int    `json:"mid"`
	Did       string `json:"did"`
	Pwd        string `json:"pwd"`
}

func (dbp *deviceUnBindParam) Validate() bool {
	return dbp.Mid > 0 && len(dbp.Did) > 0 && len(dbp.Pwd) > 0
}

type deviceUnBindRequest struct {
	Base  *common.BaseInfo `json:"base"`
	Param *deviceUnBindParam `json:"param"`
}

// sendVCodeHandler
type deviceUnBindHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *deviceUnBindHandler) Method() string {
	return http.MethodPost
}

func (handler *deviceUnBindHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true)
	defer log.InfoAll()
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := new(deviceUnBindRequest) // request body

	if !common.ParseHttpBodyParams(request, requestData) {
		response.SetResponseBase(constants.RC_PROTOCOL_ERR)
		return
	}
	param := requestData.Param

	if param == nil || !param.Validate() {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	httpHeader := common.ParseHttpHeaderParams(request)

	// if httpHeader.IsValid() == false {
	if !httpHeader.IsValidTimestamp() || !httpHeader.IsValidTokenhash() {
		log.Info("asset trans commited: request param error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	uidStr, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := common.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("asset trans commited: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if len(aesKey) != constants.AES_totalLen {
		log.Info("asset trans commited: get aeskey from cache error:", len(aesKey))
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	uid := utils.Str2Int64(uidStr)

	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]

	password, err := utils.AesDecrypt(param.Pwd, key, iv)
	if err != nil {
		log.Error("aes decrypt error ", err.Error())
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if !common.CheckLoginPwd(uid, password) {
		response.SetResponseBase(constants.RC_INVALID_LOGIN_PWD)
		return
	}


	if common.CheckDeviceLockUid(uid) || common.CheckDeviceLockDid(param.Did) {
		log.Error("unbind device uid or did in lock")
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	//  lock uid,did
	common.DeviceLockUid(uid)
	if len(param.Did) > 0 {
		if !execUnbind(uid,param.Mid,param.Did,log) {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
		}
	} else {
		if !execUnbindAll(uid,param.Mid,log) {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
		}
	}
	//锁定矿机绑定时间
	common.SetUnbindLimt(uid,param.Mid)
	common.DeviceUnLockUid(uid)

}


func execUnbind(uid int64,mid int,did string,log *logger.LvtLogger)bool{
	f := false
	common.DeviceLockDid(did)
	// query device info
	device,err := common.QueryDevice(uid,mid,did)
	if err != nil {
		log.Error("query device error",err.Error())
	}else{
		// device bind history insert
		if err := common.InsertDeviceBindHistory(device);err != nil {
			log.Error("insert device history error",err.Error())
		}else{
			// delete device info
			if err := common.DeleteDevice(device.Uid,device.Mid,device.Appid,device.Did);err != nil {
				log.Error("delete device error",err.Error())
			}else {
				// set unbind time
				f = true
			}
		}
	}
	//  unlock
	common.DeviceUnLockDid(did)
	return f
}

func execUnbindAll(uid int64,mid int,log *logger.LvtLogger)bool{
	f := false
	// query device info
	device,err := common.QueryAllDevice(uid,mid)
	if err != nil {
		log.Error("query device error",err.Error())
	}else{
		// device bind history insert
		if err := common.InsertAllDeviceBindHistory(device);err != nil {
			log.Error("insert device history error",err.Error())
		}else{
			// delete device info
			for _,v := range device {
				common.DeleteDevice(v.Uid,v.Mid,v.Appid,v.Did)
			}
			f = true
		}
	}
	return f
}