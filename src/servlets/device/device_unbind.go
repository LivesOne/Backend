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
	//锁定矿机绑定时间
	common.SetUnbindLimt(uid, param.Mid)
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
		deviceLockTs := common.DeviceLock(appid,did)
		if deviceLockTs == 0 {
			log.Error("unbind device device in lock uid",uid,"mid",mid,"appid",appid,"did",did)
			return constants.RC_SYSTEM_ERR
		}
		// device bind history insert
		if err := common.InsertDeviceBindHistory(device); err != nil {
			log.Error("insert device history error", err.Error())
		} else {
			// delete device info
			if err := common.DeleteDevice(device.Uid, device.Mid, device.Appid, device.Did); err != nil {
				log.Error("delete device error", err.Error())
			} else {
				// set unbind time
				res = constants.RC_OK
			}
		}
		common.DeviceUnLockDid(appid,did,deviceLockTs)
	case mgo.ErrNotFound:
		res =  constants.RC_NOT_FOUND_DEVICE
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
			deviceLockTs := common.DeviceLock(v.Appid,v.Did)
			if deviceLockTs == 0 {
				log.Error("unbind device device in lock uid",v.Uid,"mid",v.Mid,"appid",v.Appid,"did",v.Did)
				continue
			}
			device := &v
			if err := common.InsertDeviceBindHistory(device); err != nil {
				log.Error("insert device history error", err.Error())
			} else {
				// delete device info
				if err := common.DeleteDevice(device.Uid, device.Mid, device.Appid, device.Did); err != nil {
					log.Error("delete device error", err.Error())
				}
			}
			common.DeviceUnLockDid(v.Appid,v.Did,deviceLockTs)
		}
		res = constants.RC_OK
	case mgo.ErrNotFound:
		res =  constants.RC_NOT_FOUND_DEVICE
	}
	return res
}
