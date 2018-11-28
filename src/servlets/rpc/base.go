package rpc

import (
	"context"
	"google.golang.org/grpc"
	"time"
	"utils/consul"
	"utils/logger"
)

func getRpcConn(addr, servName string) *grpc.ClientConn {
	if len(addr) == 0 || len(servName) == 0 {
		logger.Error("consul addr or servName is empty")
		return nil
	}
	r := consul.NewResolver(servName)
	b := grpc.RoundRobin(r)
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	conn, err := grpc.DialContext(
		ctx, addr,
		grpc.WithBalancer(b),
		grpc.WithBlock(),
		grpc.WithInsecure())
	if err != nil {
		logger.Error("conn grpc server failed, addr: ", addr, "error info: ", err.Error())
		return nil
	}
	return conn
}
