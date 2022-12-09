package main

import (
	"context"
	"etcd/discovery"
	pb "etcd/discovery/proto"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"strconv"
)

var port = flag.Int("port", 8081, "")

type Server struct {
	pb.UnimplementedGreeterServer
}

func (s *Server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Printf("Recv Client Msg:%v \n", in.Msg)
	return &pb.HelloReply{
		Msg: "hello client",
	}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalln(err)
	}
	s := grpc.NewServer()
	// 向rpc注册，这里还没有向etcd注册
	pb.RegisterGreeterServer(s, &Server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalln(err)
	}
}

func serverRegister(s grpc.ServiceRegistrar, srv pb.GreeterServer) {
	pb.RegisterGreeterServer(s, srv)
	s1 := &discovery.Service{
		Name: "hello.Greet",
		Port: strconv.Itoa(*port),
		IP: "localhost",
		Protocol: "grpc",
	}
	// 这里我们必须要开启一个协程，因为我的ServiceRegister里面有一个死循环，在这个协程结束之前是不会停止循环的
	go discovery.ServiceRegister(s1)
}