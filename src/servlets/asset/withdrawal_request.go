package asset

import (
	"database/sql"
	"net/http"
	"regexp"
	"servlets/common"
	"servlets/constants"
	"servlets/token"
	"servlets/vcode"
	"strings"
	"utils"
	"utils/config"
	"utils/logger"
	"utils/lvthttp"
)

const (
	ERR_SUCCESS         = 0
	ERR_REQ_PARAM       = 1
	ERR_SERVER_INTERNAL = 2

	ERR_ACCOUNT_NOT_EXISTS = 1012
)

type EOSAccountResponse struct {
	Code   int                          `json:"code"`
	Result *EOSAccountInformationResult `json:"result"`
}

type EOSAccountInformationResult struct {
	RamQuota  int64  `json:"ram_quota"`
	RamUsage  int64  `json:"ram_usage"`
	NetLimit  *Limit `json:"net_limit"`
	CpuLimit  *Limit `json:"cpu_limit"`
	NetWeight int64  `json:"net_weight"`
	CpuWeight int64  `json:"cpu_weight"`
}

type Limit struct {
	Used      int64 `json:"used"`
	available int64 `json:"available"`
	max       int64 `json:"max"`
}

type withdrawRequestParams struct {
	AuthType  int    `json:"auth_type"`
	QuotaType int    `json:"quota_type"`
	VcodeType int    `json:"vcode_type"`
	VcodeId   string `json:"vcode_id"`
	Vcode     string `json:"vcode"`
	Remark    string `json:"remark"`
	Secret    string `json:"secret"`
}

type withdrawRequest struct {
	Param *withdrawRequestParams `json:"param"`
}

type withdrawRequestSecret struct {
	Address  string `json:"address"`
	Currency string `json:"currency"`
	Value    string `json:"value"`
	Pwd      string `json:"pwd"`
}

type withdrawRequestResponseData struct {
	TradeNo string `json:"trade_no"`
}

func (wqs *withdrawRequestSecret) isValid() bool {
	return len(wqs.Address) > 0 && len(wqs.Value) > 0 && len(wqs.Pwd) > 0 && len(wqs.Currency) > 0
}

type withdrawRequestHandler struct {
}

func (handler *withdrawRequestHandler) Method() string {
	return http.MethodPost
}

func (handler *withdrawRequestHandler) Handle(request *http.Request, writer http.ResponseWriter) {
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

	if !httpHeader.IsValidTokenhash() {
		response.SetResponseBase(constants.RC_INVALID_TOKEN)
		return
	}

	// 判断用户身份
	uidString, aesKey, _, tokenErr := token.GetAll(httpHeader.TokenHash)
	if err := common.TokenErr2RcErr(tokenErr); err != constants.RC_OK {
		logger.Info("asset lockList: get info from cache error:", err)
		response.SetResponseBase(err)
		return
	}
	if !utils.SignValid(aesKey, httpHeader.Signature, httpHeader.Timestamp) {
		response.SetResponseBase(constants.RC_INVALID_SIGN)
		return
	}

	uid := utils.Str2Int64(uidString)

	requestData := withdrawRequest{} // request body

	parseFlag := common.ParseHttpBodyParams(request, &requestData)
	if !parseFlag {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if requestData.Param.VcodeType > 0 {
		acc, err := common.GetAccountByUID(uidString)
		if err != nil && err != sql.ErrNoRows {
			logger.Info("query account by uid  error", err.Error())
			response.SetResponseBase(constants.RC_SYSTEM_ERR)
			return
		}
		switch requestData.Param.VcodeType {
		case 1:
			if ok, errCode := vcode.ValidateSmsAndCallVCode(acc.Phone, acc.Country, requestData.Param.Vcode, 3600, vcode.FLAG_DEF); !ok {
				logger.Info("validate sms code failed")
				response.SetResponseBase(vcode.ConvSmsErr(errCode))
				return
			}
		case 2:
			if ok, resErr := vcode.ValidateSmsUpVCode(acc.Country, acc.Phone, requestData.Param.Vcode); !ok {
				logger.Info("validate up sms code failed")
				response.SetResponseBase(resErr)
				return
			}
		default:
			response.SetResponseBase(constants.RC_PARAM_ERR)
			return
		}
	}

	iv, key := aesKey[:constants.AES_ivLen], aesKey[constants.AES_ivLen:]

	secret := new(withdrawRequestSecret)

	if err := utils.DecodeSecret(requestData.Param.Secret, key, iv, secret); err != nil {
		logger.Info("secret decode error", err.Error())
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if !secret.isValid() {
		logger.Info("withdrawal secret valid failed")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if !validateWithdrawalValue(secret.Value) {
		logger.Info("withdrawal value error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	if !validateWithdrawalAddress(secret.Address, secret.Currency) {
		logger.Info("withdrawal address format error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	pwd := secret.Pwd
	switch requestData.Param.AuthType {
	case constants.AUTH_TYPE_LOGIN_PWD:
		if !common.CheckLoginPwd(uid, pwd) {
			logger.Info("login password error")
			response.SetResponseBase(constants.RC_INVALID_LOGIN_PWD)
			return
		}
	case constants.AUTH_TYPE_PAYMENT_PWD:
		if !common.CheckPaymentPwd(uid, pwd) {
			logger.Info("trade password error")
			response.SetResponseBase(constants.RC_INVALID_PAYMENT_PWD)
			return
		}
	default:
		logger.Info("auth type parameter error")
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}

	address := strings.ToLower(secret.Address)
	if strings.EqualFold(secret.Currency, constants.TRADE_CURRENCY_LVTC) || strings.EqualFold(secret.Currency, constants.TRADE_CURRENCY_LVT) || strings.EqualFold(secret.Currency, constants.TRADE_CURRENCY_ETH) {
		if !strings.HasPrefix(address, "0x") {
			address = "0x" + address
		}
	}

	var currencyDecimal, feeCurrencyDecimal int
	if strings.EqualFold(secret.Currency, "eos") {
		if len(requestData.Param.Remark) > config.GetConfig().EOSRemarkLengthLimit {
			response.SetResponseBase(constants.RC_REMARK_TOO_LONG)
			return
		}
		if err := validateEosAccount(address); err.Rc != constants.RC_OK.Rc {
			response.SetResponseBase(err)
			return
		}
		currencyDecimal = utils.CONV_EOS
		feeCurrencyDecimal = utils.CONV_EOS
	} else {
		currencyDecimal = utils.CONV_LVT
		feeCurrencyDecimal = utils.CONV_LVT
	}
	feeCurrency, error := common.GetFeeCurrencyByCurrency(strings.ToUpper(secret.Currency))
	if error != nil {
		response.SetResponseBase(constants.RC_PARAM_ERR)
		return
	}
	if strings.EqualFold(secret.Currency, "lvtc") {
		feeCurrency = "ETH"
		feeCurrencyDecimal = utils.CONV_LVT
	}

	tradeNo, err := common.Withdraw(uid, secret.Value, address, strings.ToUpper(secret.Currency), feeCurrency, requestData.Param.Remark, currencyDecimal, feeCurrencyDecimal)
	//tradeNo, err := common.Withdraw(uid, secret.Value, address, strings.ToUpper(secret.Currency))
	if err.Rc == constants.RC_OK.Rc {
		response.Data = withdrawRequestResponseData{
			TradeNo: tradeNo,
		}
	} else {
		response.SetResponseBase(err)
	}
}

/*
 * 验证提币数额
 */
func validateWithdrawalValue(value string) bool {
	if utils.Str2Float64(value) > 0 {
		index := strings.Index(value, ".")
		if index == -1 {
			return true
		}
		last := value[index+1:]
		if len(last) <= 8 {
			return true
		}
	}
	return false
}

/*
 * 验证提币目标地址
 */
func validateWithdrawalAddress(walletAddress, currency string) bool {
	ret := false
	switch strings.ToUpper(currency) {
	case constants.TRADE_CURRENCY_LVT:
		fallthrough
	case constants.TRADE_CURRENCY_LVTC:
		fallthrough
	case constants.TRADE_CURRENCY_ETH:
		reg := "^(0x)?[0-9a-fA-F]{40}$"
		ret, _ = regexp.MatchString(reg, strings.ToLower(walletAddress))
	case constants.TRADE_CURRENCY_BTC:
		ret = len(walletAddress) > 0
	case constants.TRADE_CURRENCY_EOS:
		ret = len(walletAddress) > 0
	}
	return ret
}

func validateEosAccount(account string) constants.Error {
	url := config.GetConfig().ChainApiAddress
	if strings.HasSuffix(url, "/") {
		url += "v2/eos/account/" + account
	} else {
		url += "/v2/eos/account/" + account
	}
	response, err := lvthttp.Get(url, nil)
	if err != nil {
		logger.Error("send transcation to chain error ", err.Error())
		return constants.RC_SYSTEM_ERR
	}
	accountResponse := new(EOSAccountResponse)
	if err := utils.FromJson(response, accountResponse); err != nil {
		logger.Error("json parse error", err.Error())
		return constants.RC_SYSTEM_ERR
	}

	switch accountResponse.Code {
	case ERR_SUCCESS:
		if accountResponse.Result != nil {
			return constants.RC_OK
		} else {
			return constants.RC_INVALID_ACCOUNT
		}
	case ERR_REQ_PARAM:
		fallthrough
	case ERR_SERVER_INTERNAL:
		return constants.RC_SYSTEM_ERR
	case ERR_ACCOUNT_NOT_EXISTS:
		return constants.RC_INVALID_ACCOUNT
	default:
		return constants.RC_SYSTEM_ERR
	}
}
