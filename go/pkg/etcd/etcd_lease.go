package etcd

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func keepAlive(etcdClient *clientv3.Client, lease *clientv3.LeaseGrantResponse, ctx context.Context, key string) error {
	leaseKeepAlive, err := etcdClient.KeepAlive(ctx, lease.ID)
	if err != nil {
		logger.Sugar().Fatalw("Failed starting the keepalive for etcd: %s", err)
		return err
	}

	for range leaseKeepAlive {
		logger.Sugar().Debugf("Lease refreshed on key: " + key)
	}
	return nil
}

func CancelLeaseEtcdKey(key string) {
	etcdLeaseMap[key].cancel()
}

func UpdateEtcdKeywithLease(etcdClient *clientv3.Client, key string, value string) error {

	if leaseStruct, ok := etcdLeaseMap[key]; ok {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := etcdClient.Put(ctx, key, value, clientv3.WithLease(leaseStruct.leaseID))
		if err != nil {
			logger.Sugar().Fatalw("Failed updating item in etcd: %s", err)
			return err
		}
	} else {
		logger.Sugar().Infof("Couldn't find existing lease item, creating new item with key, %s", key)
		return PutEtcdWithLease(etcdClient, key, value)
	}
	return nil
}
