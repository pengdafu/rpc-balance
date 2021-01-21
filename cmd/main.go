package main

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	api "rpc/api/v1/http"
	"rpc/api/v1/pb"
	"rpc/configs"
	"rpc/internal/service"
	"rpc/pkg/etcd"
	"syscall"
	"time"
)

func main() {
	g, _ := errgroup.WithContext(context.Background())
	ctx, exit := context.WithCancel(context.Background())
	g.Go(func() error {
		err := rpcServer(ctx, "8081")
		if err != nil {
			exit()
		}
		return err
	})

	g.Go(func() error {
		err := httpServer(ctx, "8080", "8081")
		if err != nil {
			exit()
		}
		return err
	})

	g.Go(func() error {
		err := register(ctx)
		if err != nil {
			exit()
		}
		return err
	})

	g.Go(func() error {
		err := handlerSignal(ctx)
		if err != nil {
			exit()
		}
		return err
	})

	if err := g.Wait(); err != nil {
		log.Println(err)
	}
	time.Sleep(time.Second * 4)
	log.Println("shutdown")
}

func rpcServer(ctx context.Context, port string) error {
	serve := grpc.NewServer()
	pb.RegisterPingServer(serve, service.NewService())
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	go func() {
		<- ctx.Done()
		serve.GracefulStop()
	}()
	return serve.Serve(lis)
}

func httpServer(ctx context.Context, httpPort, grpcPort string) error {
	e, err := api.NewEngine(grpcPort)
	if err != nil {
		return err
	}
	serve := http.Server{
		Addr:    ":" + httpPort,
		Handler: e,
	}
	go func() {
		<-ctx.Done()
		c, _ := context.WithTimeout(context.Background(), 4*time.Second)
		_ = serve.Shutdown(c)
	}()
	return serve.ListenAndServe()
}

func register(ctx context.Context) error {
	srvName, ok := os.LookupEnv("SERVICE")
	if !ok {
		srvName = "balance"
	}
	podIP, ok := os.LookupEnv("POD_IP")
	if !ok {
		return errors.New("could not get current service ip")
	}
	r, err := etcd.NewServiceRegister(configs.NewConfig().Etcd.Endpoints, srvName, podIP+":8081", 20)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		_ = r.Close()
	}()
	r.ListenLease()
	return errors.New("续租失败~")
}

func handlerSignal(ctx context.Context) error {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		return nil
	case ev := <-ch:
		return fmt.Errorf("%s", ev)
	}
}