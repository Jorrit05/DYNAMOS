package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	// use the exponential backoff package
	bo "github.com/cenkalti/backoff/v4"
)

func PutValueToEtcd(etcdClient *clientv3.Client, key, value string, opts ...Option) error {
	// Start with default options
	retryOpts := DefaultRetryOptions

	// Apply any specified options
	for _, opt := range opts {
		opt(&retryOpts)
	}

	operation := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := etcdClient.Put(ctx, key, value)
		if err != nil {
			return fmt.Errorf("failed to put key-value %s-%s to etcd: %v", key, value, err)
		}

		return nil
	}

	// Create a new exponential backoff
	backoff := bo.NewExponentialBackOff()
	backoff.InitialInterval = retryOpts.InitialInterval
	backoff.MaxInterval = retryOpts.MaxInterval
	backoff.MaxElapsedTime = retryOpts.MaxElapsedTime

	// Run the operation with backoff
	err := bo.Retry(operation, backoff)

	if err != nil {
		fmt.Printf("failed after retries: %v", err)
		return err
	}

	return nil
}

func PutEtcdWithLease(etcdClient *clientv3.Client, key string, value string) error {
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

	go keepAlive(etcdClient, lease, ctx, key)

	return nil
}

// RegisterJSONArray takes a JSON array, unmarshals it into the target Iterable,
// and stores each element in the etcd key-value store.
//   - T is the underlying struct type of the target.
//   - jsonContent is the byte array containing the JSON content.
//   - target should be an instance of a struct that implements the Iterable and NameGetter interfaces.
//   - etcdClient is an instance of the etcd client.
//   - key is the etcd key prefix where the elements will be stored.
//
// Add Get(), .GetName() interfaces to struct that uses this. See archetypes/requestor as an example
func RegisterJSONArray[T any](jsonContent []byte, target Iterable, etcdClient *clientv3.Client, key string) error {

	err := json.Unmarshal(jsonContent, &target)
	if err != nil {
		logger.Sugar().Errorw("failed to unmarshal JSON content: %v", err)
		return err
	}

	for i := 0; i < target.Len(); i++ {
		element := target.Get(i).(NameGetter) // Assert that element implements NameGetter

		jsonRep, err := json.Marshal(element)
		if err != nil {
			logger.Sugar().Errorw("Failed to Marshal config: %v", err)
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = etcdClient.Put(ctx, fmt.Sprintf("%s/%s", key, string(element.GetName())), string(jsonRep))
		if err != nil {
			logger.Sugar().Errorw("Failed creating archetypesJSON in etcd: %s", err)
			return err
		}
	}

	return nil
}

// T is the struct type to be saved.
// target is an instance of the struct.
// etcdClient is an instance of the etcd client.
// key is the etcd key where the value will be stored.
func SaveStructToEtcd[T any](etcdClient *clientv3.Client, key string, target T) error {
	// Marshal the target struct into a JSON representation
	jsonRep, err := json.Marshal(target)
	if err != nil {
		logger.Sugar().Fatalw("failed to marshal struct: %v", err)
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Save the JSON representation to the etcd key-value store
	_, err = etcdClient.Put(ctx, key, string(jsonRep))

	if err != nil {
		logger.Sugar().Fatalw("failed to save struct to etcd: %v", err)
		return err
	}

	return nil
}

// T is the struct type to be saved.
// target is an instance of the struct.
// etcdClient is an instance of the etcd client.
// key is the etcd key where the value will be stored.
func SaveStructToEtcdTimeout[T any](ctx context.Context, etcdClient *clientv3.Client, key string, target T, timeout int64) error {
	// Marshal the target struct into a JSON representation
	jsonRep, err := json.Marshal(target)
	if err != nil {
		logger.Sugar().Fatalw("failed to marshal struct: %v", err)
		return err
	}

	grantResp, err := etcdClient.Grant(ctx, timeout)
	if err != nil {
		logger.Sugar().Errorf("error granting etcd lease %v", err)
		return err
	}
	_, err = etcdClient.Put(ctx, key, string(jsonRep), clientv3.WithLease(grantResp.ID))
	if err != nil {
		logger.Sugar().Errorf("error putting key in  etcd %v", err)
		return err
	}

	return nil
}
