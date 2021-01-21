package etcd

import (
  "context"
  "fmt"
  "github.com/coreos/etcd/clientv3"
  "github.com/coreos/etcd/mvcc/mvccpb"
  "google.golang.org/grpc/resolver"
  "log"
  "sync/atomic"
  "time"
)

const schema = "grpclb"

type serviceResolver struct {
	cli  *clientv3.Client
	cc   resolver.ClientConn
	srvs atomic.Value
}

func (s *serviceResolver) ResolveNow(options resolver.ResolveNowOptions) {
  // todo
}

func (s *serviceResolver) Close() {
  _ = s.cli.Close()
}

func NewServiceResolver(endpoints []string) resolver.Builder {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	return &serviceResolver{
		cli: cli,
	}
}

func (s *serviceResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	s.cc = cc

	prefix := fmt.Sprintf("/%s/%s/", target.Scheme, target.Endpoint)
	// 根据前缀获取所有的kv
	res, err := s.cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	for _, v := range res.Kvs {
    s.setSrcList(string(v.Key), string(v.Value))
	}
	s.updateState()
	go s.watcher(prefix)
  return s, nil
}

func (s *serviceResolver) Scheme() string {
	return schema
}

func (s *serviceResolver) getSrvList() []resolver.Address {
  srvs, _ := s.srvs.Load().(map[string]resolver.Address)
  addr := make([]resolver.Address, len(srvs))

  for _, v := range srvs {
    addr = append(addr, v)
  }
  return addr
}

func (s *serviceResolver) setSrcList(key, val string) {
  srvs, ok := s.srvs.Load().(map[string]resolver.Address)
  if !ok {
    srvs = make(map[string]resolver.Address)
  }
  tmp := make(map[string]resolver.Address, len(srvs) + 1)
  for k, v := range srvs {
    tmp[k] = v
  }
  tmp[key] = resolver.Address{Addr: val}
  s.srvs.Store(tmp)
  s.updateState()
}

func (s *serviceResolver) updateState() {
  s.cc.UpdateState(resolver.State{Addresses: s.getSrvList()})
}

func (s *serviceResolver) watcher(prefix string) {
  watch := s.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())

  for event := range watch {
    for _, ev := range event.Events {
      switch ev.Type {
      case mvccpb.DELETE:
        s.delSrvList(string(ev.Kv.Key))
      case mvccpb.PUT:
        s.setSrcList(string(ev.Kv.Key), string(ev.Kv.Value))
      }
    }
  }
  log.Println("watch close")
}

func (s *serviceResolver) delSrvList(key string) {
  srvs, ok := s.srvs.Load().(map[string]resolver.Address)
  if !ok {
    srvs = make(map[string]resolver.Address)
  }
  tmp := make(map[string]resolver.Address, len(srvs) + 1)
  for k, v := range srvs {
    tmp[k] = v
  }
  delete(tmp, key)
  s.srvs.Store(tmp)
  s.updateState()
}

