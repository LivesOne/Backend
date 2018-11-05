package accounts

import (
	"errors"
	"net/http"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"strings"
	"utils"
	"utils/config"
	"utils/logger"
	"utils/lvthttp"
)

type (
	reChargeAddrParam struct {
		Currency string `json:"currency"`
	}

	reChargeAddrRequest struct {
		// Base  common.BaseInfo `json:"base"`
		Param reChargeAddrParam `json:"param"`
	}

	reChargeAddrRespData struct {
		Currency string `json:"currency,omitempty"`
		Address  string `json:"address,omitempty"`
	}

	hotWalletResult struct {
		Address string `json:"address,omitempty"`
	}
	hotWalletRes struct {
		Code   int             `json:"code"`
		Result hotWalletResult `json:"result,omitempty"`
	}
)

// bindEMailHandler
type reChargeAddrHandler struct {
}

func (handler *reChargeAddrHandler) Method() string {
	return http.MethodPost
}

func (handler *reChargeAddrHandler) Handle(
	request *http.Request, writer http.ResponseWriter) {

	response := common.NewResponseData()
	defer common.FlushJSONData2Client(response, writer)

	httpHeader := common.ParseHttpHeaderParams(request)
	requestData := new(reChargeAddrRequest)
	common.ParseHttpBodyParams(request, requestData)

	if httpHeader.Timestamp < 1 || requestData.Param.Currency == "" {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	currency := strings.ToUpper(requestData.Param.Currency)

	// 判断用户身份
	uidStr, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		response.SetResponseBase(err)
		return
	}

	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	uid := utils.Str2Int64(uidStr)
	// 返回充值地址
	var addr string
	for _, rechargeAddr := range config.GetConfig().ReChargeAddress {
		if rechargeAddr.Currency == currency {
			addr = rechargeAddr.Address
			addr = strings.Trim(addr, " ")
			break
		}
	}
	if addr == "" {
		if currency != "BTC" {
			response.SetResponseBase(constants.RC_INVALID_CURRENCY)
			return
		}
		// 从user_recharge_address查询
		rechAddr, err := common.GetRechargeAddrList(uid, currency)
		if err != nil {
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
			return
		}
		if rechAddr == "" {
			// http Get config.GetConfig().ChainHotWalletAddr
			rechAddr, err = GenerateBtcWalletAddr(uidStr)
			if err != nil || rechAddr == "" {
				response.SetResponseBase(constants.RC_SYSTEM_ERR)
				return
			}
			go func(uid int64, currency, addr string) {
				// insert btc addr
				err = common.InsertRechargeAddr(uid, currency, addr)
				if err != nil {
					logger.Error("insert recharge addr error,", err)
				}
			}(uid, currency, rechAddr)
		}
		addr = rechAddr
	}
	respData := new(reChargeAddrRespData)
	respData.Currency = currency
	respData.Address = addr
	response.Data = respData
	// send response
	response.SetResponseBase(constants.RC_OK)
	return
}

// 获取hotwallet address
func GenerateBtcWalletAddr(uid string) (string, error) {
	url := config.GetConfig().ChainHotWalletAddr
	url = strings.Replace(url, ":coin", "btc", 1)
	url = strings.Replace(url, ":uid", uid, 1)
	resBody, err := lvthttp.Get(url, nil)
	if err != nil {
		logger.Error("hotwallet http req error", err.Error())
		return "", err
	}
	logger.Info("wx http res ", resBody)

	//校验是否是错误返回的格式
	res := new(hotWalletRes)
	if err = utils.FromJson(resBody, res); err != nil {
		logger.Error("hotwallet response: json parse error", err.Error(), "res body", resBody)
		return "", err
	}

	if res.Code > 0 {
		logger.Error("hotwallet response error, code:", res.Code)
		return "", errors.New("hotwallet response code error")
	}

	return res.Result.Address, nil
}
