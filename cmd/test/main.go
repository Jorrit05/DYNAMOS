package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	logger                      = lib.InitLogger()
	etcdClient *clientv3.Client = lib.GetEtcdClient(etcdEndpoints)
)

var serviceName = "test"
var etcdEndpoints = "http://localhost:30005"
var etcdLeaseMap = make(map[string]leaseStruct)

type leaseStruct struct {
	cancel  context.CancelFunc
	leaseID clientv3.LeaseID
}

func putEtcdWithLease(key string, value string) error {
	// Create context
	ctx, cancel := context.WithCancel(context.Background())

	lease, err := etcdClient.Lease.Grant(ctx, 5)
	if err != nil {
		cancel()
		return err
	}

	_, err = etcdClient.Put(ctx, key, value, clientv3.WithLease(lease.ID))
	if err != nil {
		cancel()
		return err
	}

	etcdLeaseMap[key] = leaseStruct{
		cancel:  cancel,
		leaseID: lease.ID,
	}

	go keepAlive(lease, ctx, key)

	return nil
}

func keepAlive(lease *clientv3.LeaseGrantResponse, ctx context.Context, key string) error {
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

func updateEtcdKeywithLease(key string, value string) error {

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
		return putEtcdWithLease(key, value)
	}
	return nil
}

func wrapper() {
	key := "test_key"
	value := "test_value"

	// defer cancel()

	err := putWithLease(key, value)
	if err != nil {
		log.Fatalf("failed to put with lease: %v", err)
	}
	fmt.Println("no more wrapper")
}

func main() {

	defer etcdClient.Close()
	wrapper()

	key := "test_key"

	// Use the cancel function to cancel the lease renewal.
	// This is just an example, replace it with your own cancellation condition.
	time.Sleep(4 * time.Second)
	fmt.Println("Updating lease")
	updateKey(key, "updated key")
	time.Sleep(6 * time.Second)

	// str := etcdLeaseMap["test_key"]
	// str()
	fmt.Println("cancelling lease")
	etcdLeaseMap["test_key"].cancel()
	// Give the lease renewal goroutine some time to stop.
	time.Sleep(10 * time.Second)
}
