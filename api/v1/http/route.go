package http

import (
  "github.com/gin-gonic/gin"
  "rpc/internal/service"
)

func NewEngine(grpcPort string) (*gin.Engine, error) {
  e := gin.Default()
  //gateway := runtime.NewServeMux()
  //err := pb.RegisterPingHandlerFromEndpoint(
  //  context.Background(),
  //  gateway,
  //  "localhost:" + grpcPort,
  //  []grpc.DialOption{grpc.WithInsecure()},
  //)
  //if err != nil {
  //  return nil, err
  //}
  {
    e.GET("/alive", service.Alive)
  }
  //e.NoRoute(func(c *gin.Context) {
  //  gateway.ServeHTTP(c.Writer, c.Request)
  //})
  return e, nil
}