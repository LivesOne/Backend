package device

import (
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
	"gopkg.in/mgo.v2"
	"utils/config"
)

type deviceUnBindParam struct {
	Mid   int    `json:"mid"`
	Did   string `json:"did"`
	Appid int    `json:"appid"`
	Pwd   string `json:"pwd"`
}

func (dbp *deviceUnBindParam) Validate() bool {
	return dbp.Mid > 0 && len(dbp.Pwd) > 0
}

type deviceUnBindRequest struct {
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


	ul := common.GetTransUserLevel(uid)
	ulc := config.GetLimitByLevel(ul)
	if param.Mid > 0 && param.Mid > ulc.MinerIndexSize() {
		log.Error("bind device mid index error mid:",param.Mid,"mast <",ulc.MinerIndexSize())
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

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


	userLockTs := common.DeviceUserLock(uid)
	if userLockTs == 0 {
		log.Error("unbind device uid in lock")
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}



	switch {
	case len(param.Did) == 0 && param.Appid == 0:
		resBase := execUnbindAll(uid, param.Mid, log)
		response.SetResponseBase(resBase)
	case len(param.Did) > 0 && param.Appid > 0:
		resBase := execUnbind(uid, param.Mid, param.Appid, param.Did, log)
		response.SetResponseBase(resBase)
	default:
		log.Error("unkonw unbind type uid",uid,"did",param.Did,"appid",param.Appid)
		response.SetResponseBase(constants.RC_PARAM_ERR)
	}
	common.DeviceUnLockUid(uid,userLockTs)

}

func execUnbind(uid int64, mid, appid int, did string, log *logger.LvtLogger) constants.Error {

	res := constants.RC_SYSTEM_ERR
	// query device info
	query := bson.M{
		"uid":   uid,
		"mid":   mid,
		"appid": appid,
		"did":   did,
	}
	device, err := common.QueryDevice(query)
	switch err {
	case nil:
		if execMongoAndReidsUnbind(device,log){
				// set unbind time
				res = constants.RC_OK
				//清理sid对应下所有心跳

				common.SetUnbindLimt(uid, mid)

		}
	case mgo.ErrNotFound:
		log.Error("unbind device uid",uid,"mid",mid,"appid",appid,"did",did,"device not found")
		res = constants.RC_NOT_FOUND_DEVICE
	}

	return res
}

func execUnbindAll(uid int64, mid int, log *logger.LvtLogger) constants.Error {
	res := constants.RC_SYSTEM_ERR
	// query device info
	device, err := common.QueryAllDevice(uid, mid)
	switch err {
	case nil:
		for _, v := range device {
			execMongoAndReidsUnbind(&v,log)
		}
		//锁定矿机绑定时间
		common.SetUnbindLimt(uid, mid)
		res = constants.RC_OK
	case mgo.ErrNotFound:
		log.Error("unbind all device uid",uid,"mid",mid,"device not found")
		res = constants.RC_NOT_FOUND_DEVICE
	}
	return res
}

func execMongoAndReidsUnbind(device *common.DtDevice,log *logger.LvtLogger)bool{
	deviceLockTs := common.DeviceLock(device.Appid,device.Did)
	if deviceLockTs == 0 {
		log.Error("unbind device device in lock uid",device.Uid,"mid",device.Mid,"appid",device.Appid,"did",device.Did)
		return false
	}
	f := true
	if err := common.InsertDeviceBindHistory(device); err != nil {
		log.Error("insert device history error", err.Error())
		f = false
	} else {
		// delete device info
		if err := common.DeleteDevice(device.Uid, device.Mid, device.Appid, device.Did); err != nil && err != mgo.ErrNotFound  {
			log.Error("delete device error", err.Error())
			f = false
		}

		if err := common.DelDtActive(device.Uid,device.Mid,device.Sid);err != nil && err != mgo.ErrNotFound {
			log.Error("delete dt active error", err.Error())
			f = false
		}
	}

	err := common.ClearOnline(device.Uid,device.Mid,device.Sid)
	if err != nil {
		logger.Error("remove online error",err.Error())
	}
	common.DeviceUnLockDid(device.Appid,device.Did,deviceLockTs)
	return f
}