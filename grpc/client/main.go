package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	pb "hello_client/newpb"
	"io"
	"log"
	"strconv"
	"time"
)

const (
	defaultName = "xxx"
)

var (
	//addr = flag.String("addr", "192.168.1.6:8972", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

type Authentication struct { //传递信息接口的实现
	authorization string
}

// -----------普通传输--------------------------
func run(c pb.GreeterClient) {
	// 执行RPC调用并打印收到的响应数据
	ctx, cancel := context.WithTimeout(context.Background(), time.Second) // 超时处理
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name}) //调用方法
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	fmt.Printf("Greeting: %s", r.GetReply())
}

// -------------------------- 流传输 ----------------------------------------
// 1. 服务端流
func runLostOfReplies(c pb.GreeterClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	stream, err := c.LotsOfReplies(ctx, &pb.HelloRequest{Name: *name}) //创建流 , 单个参数
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	// 利用for循环持续读取流的内容
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		fmt.Printf("Greeting: %s", in.GetReply())
	}
}

// 2. 客户端流
func runLotsOfGreetings(c pb.GreeterClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	stream, err := c.LotsOfGreetings(ctx) //创建流 , 多个参数
	name := []string{"1", "2", "3"}
	for _, name := range name { //在流里多次发送数据
		if err := stream.Send(&pb.HelloRequest{Name: name}); err != nil {
			log.Fatalf("could not send data: %v", err)
		}
	}
	res, err := stream.CloseAndRecv() // 关闭流并接受结果
	if err != nil {
		log.Fatalf("could not send data: %v", err)
	}
	fmt.Printf("Greeting: %s", res.GetReply())

}

// 3. 双向流
func runBidiHello(c pb.GreeterClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	name := []string{"1", "2", "3"}
	stream, err := c.BidiHello(ctx)
	if err != nil {
		log.Fatalf("c.BidiHello failed: %v", err)
	}
	waitch := make(chan struct{}) // 创建一个通道,监听是否接受完毕
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(waitch)
				return
			}
			if err != nil {
				log.Fatalf("stream.Recv failed: %v", err)
			}
			fmt.Printf("Greeting: %s\n", in.GetReply())
			time.Sleep(1 * time.Second)
		}
	}()

	for _, name := range name {
		if err := stream.Send(&pb.HelloRequest{Name: name}); err != nil {
			log.Fatalf("stream.Send failed: %v", err)
		}
	}
	stream.CloseSend()
	<-waitch
}

// ------------基于token和密钥进行的安全认证-----------------------
func (a *Authentication) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Printf("panic: %v", err)
		}
	}()
	ri, _ := credentials.RequestInfoFromContext(ctx) //用来实现获取服务端的认证级别并且进行校验是否达标
	if err := credentials.CheckSecurityLevel(ri.AuthInfo, credentials.PrivacyAndIntegrity); err != nil {
		return nil, fmt.Errorf("unable to transfer oauthAccess PerRPCCredentials: %v", err)
	}
	return map[string]string{
		"authorization": a.authorization,
	}, nil
}

// 用于指定是否需要传输层安全性,如果进行了TLS 或其他方式进行加密和证书验证 ,则返回true就好了
func (a *Authentication) RequireTransportSecurity() bool {
	return true
}

func main() {
	creds, err2 := credentials.NewClientTLSFromFile("../cert/server.pem", "*.mszlu .com")
	if err2 != nil {
		log.Printf("failed to create credentials: %v", err2)
	}

	flag.Parse()
	// 连接到server端
	token := &Authentication{
		authorization: "111",
	}

	//WithPerRPCCredentials用来传递认证信息,它里面定义了两个方法
	//GetRequestMetadata用来获取认证信息，RequireTransportSecurity用来判断是否需要安全传输
	//我们需要写一个结构体实现这个接口并调用

	//连接consul
	config := api.DefaultConfig()
	config.Address = "192.168.134.130:8500"
	cc, err := api.NewClient(config)
	if err != nil {
		fmt.Printf("err:%v\n", err)
		return
	}
	// 返回的是一个 map[string]*api.AgentService
	// 其中key是服务ID，值是注册的服务信息
	serviceMap, err := cc.Agent().ServicesWithFilter("Service==`my-grpc-service`")
	if err != nil {
		fmt.Printf("err:%v\n", err)
		return
	}
	// 选一个服务机（这里选最后一个）
	var addr string
	for k, v := range serviceMap {
		fmt.Printf("%s:%#v\n", k, v)
		addr = v.Address + ":" + strconv.Itoa(v.Port)
	}
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(token))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn) //创建grpc客户端对象创建存根
	//run(c)
	//runLostOfReplies(c)
	runLotsOfGreetings(c)
	//runBidiHello(c)
}
