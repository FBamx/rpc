package main

import (
	"context"
	"flag"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"
	helloworld "gorpc/message"
	"log"
	"net"
	"time"
)

const (
	MyService = "lcb/demo"
	EtcdUrl   = "http://localhost:2379"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type Server struct {
	helloworld.UnimplementedGreeterServer
}

func (s *Server) SayHello(_ context.Context, req *helloworld.Request) (resp *helloworld.Response, err error) {
	log.Printf("Received: %v", req.GetName())
	return &helloworld.Response{
		Message: "Hello " + req.GetName(),
	}, nil
}

func registerEndPointToEtcd(ctx context.Context, addr string) {
	etcdClient, err := clientv3.NewFromURL(EtcdUrl)
	if err != nil {
		log.Fatalf("err")
	}
	etcdManager, err := endpoints.NewManager(etcdClient, MyService)
	if err != nil {
		log.Fatalf("err")
	}

	var ttl int64 = 10
	lease, _ := etcdClient.Grant(ctx, ttl)
	_ = etcdManager.AddEndpoint(ctx, fmt.Sprintf("%s/%s", MyService, addr), endpoints.Endpoint{Addr: addr}, clientv3.WithLease(lease.ID))
	for {
		select {
		case <-time.After(5 * time.Second):
			// 续约操作
			resp, _ := etcdClient.KeepAliveOnce(ctx, lease.ID)
			fmt.Printf("keep alive resp: %+v", resp)
			fmt.Println()
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	flag.Parse()
	addr := fmt.Sprintf("localhost:%d", *port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	helloworld.RegisterGreeterServer(s, &Server{})
	// 注册 grpc 服务节点到 etcd 中
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go registerEndPointToEtcd(ctx, addr)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
