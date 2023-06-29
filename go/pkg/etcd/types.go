package etcd

import (
	"context"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	logger       = lib.InitLogger(logLevel)
	etcdLeaseMap = make(map[string]leaseStruct)
)

type leaseStruct struct {
	cancel  context.CancelFunc
	leaseID clientv3.LeaseID
}

type EtcdClient interface {
	Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error)
}

type EtcdClientWrapper struct {
	*clientv3.Client
}

func (w *EtcdClientWrapper) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return w.Client.Put(ctx, key, val, opts...)
}

type KVGetter interface {
	Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
}

type TestType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Iterable interface {
	Len() int
	Get(index int) interface{}
}

type NameGetter interface {
	GetName() string
}
