package rpc

import (
	"errors"
	"gitlab.maxthon.net/cloud/livesone-user-micro/src/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"servlets/constants"
	"utils/config"
	"utils/logger"
)

var (
	conn            *grpc.ClientConn
	loginClient     microuser.UserLoginServiceClient
	userCacheClient microuser.UserServiceClient
	walletClient    microuser.UserWalletServiceClient
)

func getUserClientConn() *grpc.ClientConn {
	if conn == nil {
		conn = getRpcConn(config.GetConfig().RegistryAddr, config.GetConfig().UserServiceName)
		if conn != nil {
			logger.Info("get user service rpc conn success")
		} else {
			logger.Error("cannot conn to user service")
		}
	}
	return conn
}

func GetLoginClient() microuser.UserLoginServiceClient {
	if loginClient == nil {
		loginClient = microuser.NewUserLoginServiceClient(getUserClientConn())
	}
	return loginClient
}

func GetUserCacheClient() microuser.UserServiceClient {
	if userCacheClient == nil {
		userCacheClient = microuser.NewUserServiceClient(getUserClientConn())
	}
	return userCacheClient
}

func GetWalletClient() microuser.UserWalletServiceClient {
	if walletClient == nil {
		walletClient = microuser.NewUserWalletServiceClient(getUserClientConn())
	}
	return walletClient
}

func GetUserField(uid int64, field microuser.UserField) (string, error) {
	cli := GetUserCacheClient()
	if cli != nil {
		req := &microuser.GetUserInfoReq{
			Uid:   uid,
			Field: field,
		}
		resp, err := cli.GetUserInfo(context.Background(), req)
		if err != nil {
			logger.Error("grpc GetUserField request error: ", err.Error())
			return "", err
		}
		if resp.Result != microuser.ResCode_OK {
			return "", errors.New(resp.Result.String())
		}
		return resp.Value, nil
	}
	return "", errors.New("can not get rpc client")
}

func GetUserInfo(uid int64) (*microuser.GetUserAllInfoRes, error) {
	cli := GetUserCacheClient()
	if cli != nil {
		req := &microuser.UserIdReq{
			Uid: uid,
		}
		resp, err := cli.GetUserAllInfo(context.Background(), req)
		if err != nil {
			logger.Error("grpc GetUserInfo request error: ", err)
			return nil, err
		}
		if resp.Result != microuser.ResCode_OK {
			return nil, errors.New(resp.Result.String())
		}
		return resp, nil
	}
	return nil, errors.New("can not get rpc client")
}
func SetUserField(uid int64, field microuser.UserField, value string) (bool, error) {
	cli := GetUserCacheClient()
	if cli != nil {
		req := &microuser.SetUserInfoReq{
			Uid:   uid,
			Field: field,
			Value: value,
		}
		resp, err := cli.SetUserInfo(context.Background(), req)
		if err != nil {
			logger.Error("grpc SetUserField request error: ", err)
			return false, err
		}
		if resp.Result != microuser.ResCode_OK {
			return false, errors.New(resp.Msg)
		}
		return true, nil
	}
	return false, errors.New("can not get rpc client")
}

func CheckPwd(uid int64, pwdHash string, cType microuser.PwdCheckType) (bool, error) {
	cli := GetLoginClient()
	if cli != nil {
		req := &microuser.CheckPwdReq{
			Uid:     uid,
			Type:    cType,
			PwdHash: pwdHash,
		}
		resp, err := cli.CheckPwd(context.Background(), req)
		if err != nil {
			logger.Error("grpc CheckPwd request error: ", err)
			return false, err
		}
		if resp.Result != microuser.ResCode_OK {
			return false, errors.New(resp.Msg)
		}
		return true, nil

	}
	return false, errors.New("can not get rpc client")

}

func ActiveUser(uid int64) {
	cli := GetUserCacheClient()
	if cli != nil {
		req := &microuser.UserIdReq{
			Uid: uid,
		}
		resp, err := cli.UserActive(context.Background(), req)
		if err != nil {
			logger.Error("grpc UserActive request error: ", err)
		}
		if resp.Result != microuser.ResCode_OK {
			logger.Error("active user failed", resp.Msg)
		}
	}
}

func GetTokenInfo(tkHash string) (string, string, string, microuser.ResCode) {
	cli := GetLoginClient()
	if cli != nil {
		req := &microuser.GetLoginInfoReq{
			TokenHash: tkHash,
		}
		resp, err := cli.GetLoginInfo(context.Background(), req)
		if err == nil {
			return resp.Uid, resp.Key, resp.Token, resp.Result
		}
		logger.Error("grpc GetTokenInfo request error: ", err)

	}
	return "", "", "", microuser.ResCode_ERR_SYSTEM
}

func TokenErr2RcErr(tokenErr microuser.ResCode) constants.Error {
	switch tokenErr {
	case microuser.ResCode_OK:
		return constants.RC_OK
	case microuser.ResCode_ERR_PARAM:
		return constants.RC_PARAM_ERR
	case microuser.ResCode_ERR_NOTFOUND:
		return constants.RC_INVALID_TOKEN
	case microuser.ResCode_ERR_INVALID_TOKEN:
		return constants.RC_INVALID_TOKEN
	case microuser.ResCode_ERR_FAILED:
		return constants.RC_PARAM_ERR
	default:
		return constants.RC_SYSTEM_ERR
	}
}

func UserExists(uid int64) bool {
	cli := GetUserCacheClient()
	if cli != nil {
		req := &microuser.UserIdReq{
			Uid: uid,
		}
		resp, err := cli.UserExists(context.Background(), req)
		if err != nil {
			logger.Error("grpc UserExists request error: ", err)
			return false
		}
		if resp.Result == microuser.ResCode_OK {
			return true
		}
	}
	return false
}
