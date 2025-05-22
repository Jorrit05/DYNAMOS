package etcd

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func PutEtcdWithGrant(ctx context.Context, etcdClient *clientv3.Client, key string, value string, timeout int64) error {

	grantResp, err := etcdClient.Grant(ctx, timeout)
	if err != nil {
		logger.Sugar().Errorf("error granting etcd lease %v", err)
		return err
	}
	_, err = etcdClient.Put(ctx, key, value, clientv3.WithLease(grantResp.ID))
	if err != nil {
		logger.Sugar().Errorf("error putting key in  etcd %v", err)
		return err
	}
	return nil
}
