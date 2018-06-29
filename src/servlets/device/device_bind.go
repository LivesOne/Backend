package device

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/logger"
	"utils/config"
)

type deviceBindParam struct {
	Mid       int    `json:"mid"`
	Appid     int    `json:"appid"`
	Plat      int    `json:"plat"`
	Did       string `json:"did"`
	Dn        string `json:"dn"`
	OsVersion string `json:"os_ver"`
}

func (dbp *deviceBindParam) Validate() bool {
	return dbp.Mid > 0 && dbp.Appid > 0 && dbp.Plat > 0 &&
		len(dbp.Did) > 0 && len(dbp.Dn) > 0 && len(dbp.OsVersion) > 0
}

type deviceBindRequest struct {
	Base  *common.BaseInfo `json:"base"`
	Param *deviceBindParam `json:"param"`
}

// sendVCodeHandler
type deviceBindHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *deviceBindHandler) Method() string {
	return http.MethodPost
}

func (handler *deviceBindHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true)
	defer log.InfoAll()
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

	requestData := new(deviceBindRequest) // request body

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

	if common.CheckUnbindLimit(uid,param.Mid) {
		log.Error("uid",uid,"mid",param.Mid,"device unbind time too short")
		response.SetResponseBase(constants.RC_DEVICE_BIND_TOO_SHORT)
		return
	}

	query := bson.M{
		"uid": uid,
		"mid": param.Mid,
	}

	devicelist, err := common.QueryMinerBindDevice(query)

	if err != nil && err != mgo.ErrNotFound {
		log.Error("uid",uid,"mid",param.Mid,"query mongo error")
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	if err != mgo.ErrNotFound {
		// uid,mid,plat,appid check
		for _, v := range devicelist {
			if v.Plat != param.Plat {
				log.Error("uid",uid,"mid",param.Mid,"plat not match")
				response.SetResponseBase(constants.RC_DEVICE_PLAT_NOT_MATCH)
				return
			}
			if v.Appid == param.Appid {
				log.Error("uid",uid,"mid",param.Mid,"appid exists")
				response.SetResponseBase(constants.RC_DEVICE_DUP_APPID)
				return
			}
		}
	}
	//  DID check
	query = bson.M{
		"appid":param.Appid,
		"did": param.Did,
	}
	deviceCount, err := common.QueryMinerBindDeviceCount(query)
	if err != nil && err != mgo.ErrNotFound {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}
	if deviceCount > 0 {
		log.Error("uid",uid,"mid",param.Mid,"appid",param.Appid,"did appid already bind")
		response.SetResponseBase(constants.RC_DEVICE_DUP_BIND)
		return
	}


	ul := common.GetTransUserLevel(uid)
	ulc := config.GetLimitByLevel(ul)
	if param.Mid > 0 && param.Mid > ulc.MinerIndexSize() {
		log.Error("bind device mid index error",param.Mid,"mast <",ulc.MinerIndexSize())
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	// check and  lock uid,did
	userLockTs := common.DeviceUserLock(uid)
	if userLockTs == 0 {
		log.Error("unbind device uid",uid," in lock")
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}

	deviceLockTs := common.DeviceLock(param.Appid,param.Did)
	if deviceLockTs == 0 {
		log.Error("unbind device device in lock uid",uid,"mid",param.Mid,"appid",param.Appid,"did",param.Did)
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}


	// bind
	device := &common.DtDevice{
		Uid:    uid,
		Mid:    param.Mid,
		Plat:   param.Plat,
		Appid:  param.Appid,
		Did:    param.Did,
		Dn:     param.Dn,
		OsVer:  param.OsVersion,
		BindTs: utils.GetTimestamp13(),
	}
	if err := common.InsertDeviceBind(device); err != nil {
		log.Error("bind device error", err.Error())
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
	}
	//  unlock
	common.DeviceUnLockDid(param.Appid,param.Did,deviceLockTs)
	common.DeviceUnLockUid(uid,userLockTs)
	//common.DeviceUnLockUid(uid)
	//common.DeviceUnLockDid(param.Did)

}
