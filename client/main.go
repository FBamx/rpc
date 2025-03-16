package main

import (
	"context"
	"flag"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	helloworld "gorpc/message"
	"time"
)

const (
	defaultName = "lichubin"
	MyService   = "lcb/demo"
	EtcdUrl     = "http://localhost:2379"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

func main() {
	flag.Parse()

	// 创建 etcd 客户端
	etcdClient, _ := clientv3.NewFromURL(EtcdUrl)

	// 创建 etcd 实现的 grpc 服务注册发现模块 resolver
	etcdResolverBuilder, _ := resolver.NewBuilder(etcdClient)

	// 拼接服务名称，需要固定义 etcd:/// 作为前缀
	etcdTarget := fmt.Sprintf("etcd:///%s", MyService)

	// 创建 grpc 连接代理
	conn, _ := grpc.Dial(
		// 服务名称
		etcdTarget,
		// 注入 etcd resolver
		grpc.WithResolvers(etcdResolverBuilder),
		// 声明使用的负载均衡策略为 roundrobin     grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	defer conn.Close()

	// 创建 grpc 客户端
	client := helloworld.NewGreeterClient(conn)
	for {
		// 发起 grpc 请求
		resp, _ := client.SayHello(context.Background(), &helloworld.Request{Name: *name})
		fmt.Printf("resp: %+v \n", resp)
		// 每隔 1s 发起一轮请求
		<-time.After(time.Second)
	}
}
