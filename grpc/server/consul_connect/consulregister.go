package Consul_connect

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"log"
	"net"
)

type Consul struct {
	client *api.Client
}

// ---------------consul注销服务------------------------
func (c *Consul) Deregister(serviceID string) error {
	return c.client.Agent().ServiceDeregister(serviceID)
}

// --------------consul注册服务------------------------
func NewConsul(addr string) (*Consul, error) {
	config := api.DefaultConfig()
	config.Address = addr
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &Consul{client: client}, nil
}

// GetOutboundIP 获取本机的出口IP
func GetOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}

// RegisterService -------------将gRPC服务注册到consul------------------
func (c *Consul) RegisterService(serviceName string, ip string, port int) (string, error) {
	check := &api.AgentServiceCheck{
		TCP:      fmt.Sprintf("%s:%d", ip, port), // 这里一定是外部可以访问的地址
		Timeout:  "30s",                          //表示Consul健康检查在等待服务响应的时间可以长达30秒。
		Interval: "60s",                          //表示Consul每60秒进行一次健康检查
		// 指定时间后自动注销不健康的服务节点
		// 最小超时时间为1分钟，收获不健康服务的进程每30秒运行一次，因此触发注销的时间可能略长于配置的超时时间。
		DeregisterCriticalServiceAfter: "1m",
	}
	id := fmt.Sprintf("%s-%s-%d", serviceName, ip, port) // 服务唯一ID
	srv := &api.AgentServiceRegistration{
		ID:      id,
		Name:    serviceName,               // 服务名称
		Tags:    []string{"q1mi", "hello"}, // 为服务打标签
		Address: ip,
		Port:    port,
		Check:   check,
	}
	return id, c.client.Agent().ServiceRegister(srv)
}

// 服务注册具体流程
func Register() (consulClient *Consul, err error, id string) {
	//------------------------------ Consul注册-----------------------------------------
	// 将gRPC服务注册到Consul

	consulClient, err = NewConsul("192.168.134.130:8500") // 替换为你的Consul地址
	if err != nil {
		log.Fatalf("连接到Consul失败: %v", err)
	}
	// 获取本机的出口IP
	ip, err := GetOutboundIP()
	if err != nil {
		log.Fatalf("获取本机出口IP失败: %v", err)
	}
	serviceName := "my-grpc-service"
	servicePort := 8972 // 替换为你的服务端口
	id, err = consulClient.RegisterService(serviceName, ip.String(), servicePort)
	if err != nil {
		log.Fatalf("注册服务到Consul失败: %v", err)
	}
	fmt.Println("服务已成功注册到Consul")
	return
}
