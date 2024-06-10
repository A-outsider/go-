package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	pb "hello_server/newpb"
	"io"
	"log"
	"net"
)

type server struct { //这里默认建议嵌入UnimplementedGreeterServer, 提供默认实现和向前兼容性
	pb.UnimplementedGreeterServer
}

// -------普通传输---------------------------------------------------------
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{Reply: "hello" + in.Name}, nil
}

// -------------------------- 流传输 ----------------------------------------

// 服务端流
func (s *server) LotsOfReplies(in *pb.HelloRequest, stream pb.Greeter_LotsOfRepliesServer) error {
	words := []string{
		"你好",
		"hello",
		"こんにちは",
		"안녕하세요",
	}
	for _, word := range words {
		data := &pb.HelloResponse{
			Reply: word + in.GetName(),
		}
		if err := stream.Send(data); err != nil {
			return err
		}
	}
	return nil
}

// 客户端流
func (s *server) LotsOfGreetings(stream pb.Greeter_LotsOfGreetingsServer) error {
	reply := "你好 : "
	for {
		// 接收客户端发送的数据
		in, err := stream.Recv()
		if err == io.EOF {
			// 接受完毕并返回数据
			return stream.SendAndClose(&pb.HelloResponse{Reply: reply})
		}
		if err != nil {
			return err
		}
		reply += in.GetName()
		//fmt.Printf("Greeting: %v", in.GetName())
	}
}

// 双向流
func (s *server) BidiHello(stream pb.Greeter_BidiHelloServer) error {
	// 在for循环中边接受流,边发送
	for {
		// 接收客户端发送的数据
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := stream.Send(&pb.HelloResponse{Reply: "你好 : " + in.GetName()}); err != nil {
			return err
		}
	}
}

// -------------------------- token认证拦截器 ----------------------------------------
// unaryInterceptor 服务端一元拦截器
func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// authentication (token verification)
	md, ok := metadata.FromIncomingContext(ctx) //获取元信息
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
	}
	if !valid(md["authorization"]) {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}
	//认证通过 , 用handler(ctx, req) 调用实际处理 gRPC 请求的处理函数。
	m, err := handler(ctx, req)

	if err != nil {
		fmt.Printf("RPC failed with error %v\n", err)
	}
	return m, err
}

func valid(strings []string) bool { //进行拦截器信息部分的校验
	if len(strings) == 0 {
		return false
	}
	if strings[0] == "111" {
		return true
	}
	return false
}

func main() {
	creds, err2 := credentials.NewServerTLSFromFile("../cert/server.pem", "../cert/server.key")
	if err2 != nil {
		log.Printf("failed to create credentials: %v", err2)
	}
	lis, err := net.Listen("tcp", ":8972") //创建一个连接
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
		return
	}
	var authInterceptor = unaryInterceptor // 创建一个拦截器
	//// 使用自己的单向证书certs并使用拦截器校验token信息来创建服务
	s := grpc.NewServer(grpc.Creds(creds), grpc.UnaryInterceptor(authInterceptor))

	pb.RegisterGreeterServer(s, &server{}) // 向grpc注册一个服务端的服务
	err = s.Serve(lis)                     // 启动服务
	if err != nil {
		fmt.Printf("failed to serve: %v", err)
		return
	}
}
