package lib

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdClient interface {
	Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error)
}

type EtcdClientWrapper struct {
	*clientv3.Client
}

func (w *EtcdClientWrapper) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return w.Client.Put(ctx, key, val, opts...)
}
