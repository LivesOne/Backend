package consul

import (
	"context"
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

type ConsulRegistry struct {
	client    *consul.Client
	serviceId string
}

func NewConsulRegistry() *ConsulRegistry {
	return new(ConsulRegistry)
}

// Register is the helper function to self-register service into Etcd/Consul server
// name - service name
// host - service host
// port - service port
// target - consul dial address, for example: "127.0.0.1:8500"
// interval - interval of self-register to etcd
// ttl - ttl of the register information
func (cr *ConsulRegistry) Register(name string, host string, port int, target string, ttl int, server *grpc.Server) error {
	conf := &consul.Config{Scheme: "http", Address: target}
	client, err := consul.NewClient(conf)
	if err != nil {
		return fmt.Errorf("consul: create consul client error: %v", err)
	}
	cr.client = client
	serviceID := fmt.Sprintf("%s-%s-%d", name, host, port)
	cr.serviceId = serviceID
	//de-register if meet signhup
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
		x := <-ch
		log.Println("consul: receive signal: ", x)
		cr.UnRegister()
		s, _ := strconv.Atoi(fmt.Sprintf("%d", x))
		os.Exit(s)
	}()

	grpc_health_v1.RegisterHealthServer(server, &HealthImpl{})
	// initial register service
	check := consul.AgentServiceCheck{GRPC: fmt.Sprintf("%v:%v/%v", host, port, name),
		Interval: strconv.Itoa(ttl) + "s", Timeout: "5s"}
	regis := &consul.AgentServiceRegistration{
		ID:      serviceID,
		Name:    name,
		Address: host,
		Port:    port,
		Check:   &check,
	}
	err = client.Agent().ServiceRegister(regis)
	if err != nil {
		return fmt.Errorf("consul: initial register service '%s' host to consul error: %s", name, err.Error())
	}
	return nil
}

func (cr *ConsulRegistry) UnRegister() error {
	err := cr.client.Agent().ServiceDeregister(cr.serviceId)
	if err != nil {
		log.Println("consul: deregister service error: ", err.Error())
	} else {
		log.Println("consul: deregistered service from consul server.")
	}

	err = cr.client.Agent().CheckDeregister(cr.serviceId)
	if err != nil {
		log.Println("consul: deregister check error: ", err.Error())
	}
	return err
}

// HealthImpl 健康检查实现
type HealthImpl struct{}

// Check 实现健康检查接口，这里直接返回健康状态，这里也可以有更复杂的健康检查策略，比如根据服务器负载来返回
func (h *HealthImpl) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}
