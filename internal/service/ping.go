package service

import (
	"context"
  "log"
  "os"
  "rpc/api/v1/pb"
)

type service struct {
}

func NewService() *service {
	return &service{}
}

func (s *service) Ping(ctx context.Context, request *pb.PingRequest) (*pb.PingResponse, error) {
  h, err := os.Hostname()
  if err != nil {
    return nil, err
  }
  if h == "" {
    h = "local-custom"
  }
  log.Println("received: ", request)
	return &pb.PingResponse{
		P: "pong: " + h,
	}, nil
}
