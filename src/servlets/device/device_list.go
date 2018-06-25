package device

import (
	"gopkg.in/mgo.v2"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"utils"
	"utils/config"
	"utils/logger"
)

type minerDevice struct {
	Appid    int    `json:"appid"`
	Did      string `json:"did"`
	Dn       string `json:"dn"`
	OsServer string `json:"os_server"`
	BindTime int64  `json:"bind_time"`
}

type miners struct {
	Mid     int           `json:"mid"`
	Plat    int           `json:"plat"`
	IsValid bool          `json:"is_valid"`
	Devices []minerDevice `json:"devices"`
}

// sendVCodeHandler
type deviceListHandler struct {
	//header      *common.HeaderParams // request header param
	//requestData *sendVCodeRequest    // request body
}

func (handler *deviceListHandler) Method() string {
	return http.MethodPost
}

func (handler *deviceListHandler) Handle(request *http.Request, writer http.ResponseWriter) {
	log := logger.NewLvtLogger(true)
	defer log.InfoAll()
	response := &common.ResponseData{
		Base: &common.BaseResp{
			RC:  constants.RC_OK.Rc,
			Msg: constants.RC_OK.Msg,
		},
	}
	defer common.FlushJSONData2Client(response, writer)

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

	deviceAllList, err := common.QueryUserAllDevice(uid)
	if err != nil && err != mgo.ErrNotFound {
		response.SetResponseBase(constants.RC_SYSTEM_ERR)
		return
	}
	response.Data = convDevicelistToMiners(deviceAllList, uid)

}

func convDevicelistToMiners(deviceList []common.DtDevice, uid int64) []miners {

	cache := make(map[int]miners, 0)

	for _, v := range deviceList {
		m, ok := cache[v.Mid]
		if !ok {
			m = miners{
				Mid:            v.Mid,
				Plat:           v.Plat,
				IsValid: 		!common.CheckUnbindLimit(uid),
				Devices:        make([]minerDevice, 0),
			}
		}

		mm := minerDevice{
			Appid:    v.Appid,
			Did:      v.Did,
			Dn:       v.Dn,
			OsServer: v.OsVer,
			BindTime: v.BindTs,
		}
		m.Devices = append(m.Devices, mm)
		cache[v.Mid] = m
	}

	ul := common.GetTransUserLevel(uid)

	ulc := config.GetLimitByLevel(ul)

	res := make([]miners, 0)

	for i := 0; i < ulc.MinerIndexSize(); i++ {
		m, ok := cache[i]
		if !ok {
			m = miners{
				Mid: i,
			}
		}
		res = append(res, m)
	}

	return res
}
