package rpc

import (
	"google.golang.org/grpc"
	"utils/consul"
	"utils/logger"
)

func getRpcConn(addr, servName string) *grpc.ClientConn {
	if len(addr) == 0 || len(servName) == 0 {
		logger.Info("consul addr or servName is empty")
		return nil
	}
	r := consul.NewResolver(servName)
	b := grpc.RoundRobin(r)
	conn, err := grpc.Dial(addr,
		grpc.WithBalancer(b),
		grpc.WithBlock(),
		grpc.WithInsecure())
	if err != nil {
		logger.Error("conn grpc server failed, addr: ", addr, "error info: ", err.Error())
		return nil
	}
	return conn
}
