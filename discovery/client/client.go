package client

import (
	"context"
	"etcd/discovery"
	pb "etcd/discovery/proto"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

func main() {
	// 通过协程开启一个监听
	go discovery.WatchServiceName("hello.Greeter")
	for {
		sayHello()
		time.Sleep(time.Second * 200)
	}
}

func getServerAddress(svcName string) string {
	s := discovery.ServerDiscovery(svcName)
	if s == nil {
		return ""
	}
	return s.IP + ":" + s.Port
}

func sayHello() {
	addr := getServerAddress("hello.Greeter")
	if addr == "" {
		log.Fatalln("未发现可用服务")
	}
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	c := pb.NewGreeterClient(conn)
	in := &pb.HelloRequest{
		Msg: "hello server",
	}
	r, err := c.SayHello(context.Background(), in)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Recv server msg:", r.Msg)
}
