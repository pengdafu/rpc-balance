package etcd

import (
  "context"
  "fmt"
  "github.com/coreos/etcd/clientv3"
  "log"
  "time"
)

type serviceRegister struct {
	cli     *clientv3.Client
	leaseID clientv3.LeaseID

	keepalive <-chan *clientv3.LeaseKeepAliveResponse
	key       string
	val       string
}

func NewServiceRegister(endpoints []string, srvName, srvAddr string, lease int64) (*serviceRegister, error) {
  cli, err := clientv3.New(clientv3.Config{
    Endpoints: endpoints,
    DialTimeout: 5 * time.Second,
  })
  if err != nil {
    return nil, err
  }
  srv := &serviceRegister{
    cli: cli,
    key: fmt.Sprintf("/%s/%s/%s", schema, srvName, srvAddr),
    val: srvAddr,
  }
  if err := srv.putKeyWithLease(lease); err != nil {
    return nil, err
  }
  return srv, nil
}

func (s *serviceRegister) putKeyWithLease(lease int64) error {
  // 设置租约时间
  resp, err := s.cli.Grant(context.Background(), lease)
  if err != nil {
    return err
  }

  // 注册服务
  _, err = s.cli.Put(context.Background(), s.key, s.val, clientv3.WithLease(resp.ID))
  if err != nil {
    return err
  }

  // 设置心跳，定期续期
  keepalive, err := s.cli.KeepAlive(context.Background(), resp.ID)
  if err != nil {
    return err
  }

  s.keepalive = keepalive
  s.leaseID = resp.ID
  return nil
}

func (s *serviceRegister) ListenLease() {
  for res := range s.keepalive {
    log.Println("续约成功", res)
  }
  log.Println("关闭续租")
}

// Close 注销服务
func (s *serviceRegister) Close() error {
  //撤销租约
  if _, err := s.cli.Revoke(context.Background(), s.leaseID); err != nil {
    return err
  }
  log.Println("撤销租约")
  return s.cli.Close()
}