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

type deviceForceUnBindParam struct {
	Uid   int    `json:"uid"`
	Mid   int    `json:"mid"`
	Did   string `json:"did"`
	Appid int    `json:"appid"`
}

func (dbp *deviceForceUnBindParam) Validate() bool {
	return dbp.Uid > 0 && dbp.Mid > 0 && dbp.Appid > 0 && len(dbp.Did) > 0
}

type deviceForceUnBindRequest struct {
	Param *deviceForceUnBindParam `json:"param"`
}

type deviceForceUnBindHandler struct {
}

func (handler *deviceForceUnBindHandler) Method() string {
	return http.MethodPost
}

func (handler *deviceForceUnBindHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true)
	defer log.InfoAll()
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := new(deviceForceUnBindRequest) // request body

	if !common.ParseHttpBodyParams(request, requestData) {
		response.SetResponseBase(constants.RC_PROTOCOL_ERR)
		return
	}
	param := requestData.Param

	if param == nil || !param.Validate() {
		log.Info("device force unbind: request param invalid")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	httpHeader := common.ParseHttpHeaderParams(request)

	// if httpHeader.IsValid() == false {
	if !httpHeader.IsValidTimestamp() || !httpHeader.IsValidTokenhash() {
		log.Info("device force unbind: request header invalid")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	// 判断用户身份
	forceUidStr, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := common.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		log.Info("device force unbind: get info from redis error:", err)
		response.SetResponseBase(err)
		return
	}
	if len(aesKey) != constants.AES_totalLen {
		log.Info("device force unbind: get aeskey from redis error:", len(aesKey))
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		log.Info("device force unbind: signature parse invalid")
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	forceUid := utils.Str2Int64(forceUidStr)
	uid := int64(param.Uid)
	// 设备强制解绑24小时内禁止再次进行强制解绑
	if common.CheckForceUnbindLimit(param.Appid, param.Did) {
		log.Error("uid", uid, "mid", param.Mid, "device unbind interval too short")
		response.SetResponseBase(constants.RC_DEVICE_UNBIND_TOO_SHORT)
		return
	}

	ul := common.GetTransUserLevel(uid)
	ulc := config.GetLimitByLevel(ul)
	if param.Mid > ulc.MinerIndexSize() {
		log.Error("bind device mid index error mid:",param.Mid,"mast <",ulc.MinerIndexSize())
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	userLockTs := common.DeviceUserLock(uid)
	if userLockTs == 0 {
		log.Error("device force unbind: unbind device uid in lock")
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	resBase := execForceUnbind(uid, forceUid, param.Mid, param.Appid, param.Did, log)
	common.DeviceUnLockUid(uid,userLockTs)
	response.SetResponseBase(resBase)
}

func execForceUnbind(uid, forceUid int64, mid, appid int, did string, log *logger.LvtLogger) constants.Error {

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
		if execMongoAndReidsForceUnbind(device, forceUid,log){
			// set unbind time
			res = constants.RC_OK
			common.SetForceUnbindLimit(appid, did)
			common.ClearOnline(uid,mid,device.Sid)
		}
	case mgo.ErrNotFound:
		log.Error("force unbind device uid",uid,"mid",mid,"appid",appid,"did",did,"device not found")
		res = constants.RC_NOT_FOUND_DEVICE
	}

	return res
}

func execMongoAndReidsForceUnbind(device *common.DtDevice, forceUid int64,log *logger.LvtLogger)bool{
	deviceLockTs := common.DeviceLock(device.Appid,device.Did)
	if deviceLockTs == 0 {
		log.Error("unbind device in lock uid",device.Uid,"mid",device.Mid,"appid",device.Appid,"did",device.Did)
		return false
	}
	f := true
	if err := common.InsertDeviceForceUnBindHistory(device, forceUid); err != nil {
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
	common.DeviceUnLockDid(device.Appid,device.Did,deviceLockTs)
	return f
}