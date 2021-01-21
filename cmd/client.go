package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/resolver"
	"log"
	"rpc/api/v1/pb"
	"rpc/configs"
	"rpc/pkg/etcd"
	"time"
)

func main() {
	r := etcd.NewServiceResolver(configs.NewConfig().Etcd.Endpoints)
  resolver.Register(r)
	conn, err := grpc.Dial(
		r.Scheme() + "://8.8.8.8/balance",
		grpc.WithInsecure(),
		grpc.WithBlock(),
    grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)))
	if err != nil {
		log.Fatal("启动失败", err)
	}
	defer conn.Close()
	c := pb.NewPingClient(conn)
	for range time.Tick(time.Second) {
		fmt.Println(c.Ping(context.Background(), &pb.PingRequest{P: "Ping"}))
	}
}
