package rpc

import (
	"gitlab.maxthon.net/cloud/base-sms-gateway/src/proto"
	"gitlab.maxthon.net/cloud/base-vcode/src/proto"
	"utils/config"
)

var (
	vcodeClient      vcodeproto.ImgEmailServiceClient
	grpcSmsClient    smspb.SmsServiceClient
)

func GetSmsClient() smspb.SmsServiceClient {
	if grpcSmsClient == nil {
		conn := getRpcConn(config.GetConfig().RegistryAddr,config.GetConfig().SmsSvrName)
		if conn == nil {
			return nil
		}
		grpcSmsClient = smspb.NewSmsServiceClient(conn)
	}
	return grpcSmsClient
}

func GetVcodeClient() vcodeproto.ImgEmailServiceClient {
	if vcodeClient == nil {
		conn := getRpcConn(config.GetConfig().RegistryAddr,config.GetConfig().ImgEmailSvrName)
		if conn == nil {
			return nil
		}
		vcodeClient = vcodeproto.NewImgEmailServiceClient(conn)
	}
	return vcodeClient
}
