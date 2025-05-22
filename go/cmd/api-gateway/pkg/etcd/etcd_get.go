package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	bo "github.com/cenkalti/backoff/v4"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type ErrKeyNotFound struct {
	Key string
}

func (e *ErrKeyNotFound) Error() string {
	return fmt.Sprintf("key %s not found in etcd", e.Key)
}

type ErrEtcdOperation struct {
	Key string
	Err error
}

func (e *ErrEtcdOperation) Error() string {
	return fmt.Sprintf("failed to get key %s from etcd: %v", e.Key, e.Err)
}

func retryWrapper(operation func() error, opts ...Option) (string, error) {
	var value string
	// Start with default options
	retryOpts := DefaultRetryOptions

	// Apply any specified options
	for _, opt := range opts {
		opt(&retryOpts)
	}

	// Create a new exponential backoff
	backoff := bo.NewExponentialBackOff()
	backoff.InitialInterval = retryOpts.InitialInterval
	backoff.MaxInterval = retryOpts.MaxInterval
	backoff.MaxElapsedTime = retryOpts.MaxElapsedTime

	// Run the operation with backoff
	err := bo.Retry(operation, backoff)

	if err != nil {
		logger.Sugar().Errorf("failed retrieving key from etcd: %v", err)
		return "", err
	}

	return value, nil

}
func GetKeysFromPrefix(etcdClient *clientv3.Client, key string, opts ...Option) ([]string, error) {
	var keys []string
	// Start with default options
	retryOpts := DefaultRetryOptions

	// Apply any specified options
	for _, opt := range opts {
		opt(&retryOpts)
	}
	operation := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := etcdClient.Get(ctx, key, clientv3.WithPrefix())
		if err != nil {
			return &ErrEtcdOperation{Key: key, Err: err}
			// return fmt.Errorf("failed to get key %s from etcd: %v", key, err)
		}

		if len(resp.Kvs) == 0 {
			// If key not found, return an error to trigger a retry
			return &ErrKeyNotFound{Key: key}
			// return fmt.Errorf("key %s not found in etcd", key)
		}

		for _, ev := range resp.Kvs {
			key := string(ev.Key)
			// Get last part of key
			parts := strings.Split(key, "/")
			value := parts[len(parts)-1]
			keys = append(keys, value)
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
		switch e := err.(type) {
		case *ErrKeyNotFound:
			logger.Sugar().Info("No active jobs for this user: %v", e)
			return []string{}, nil

		default:
			logger.Sugar().Errorf("failed retrieving key from etcd: %v", e)
			return []string{}, err
		}
	}

	return keys, nil
}

func GetValueFromEtcd(etcdClient *clientv3.Client, key string, opts ...Option) (string, error) {
	var value string
	// Start with default options
	retryOpts := DefaultRetryOptions

	// Apply any specified options
	for _, opt := range opts {
		opt(&retryOpts)
	}
	operation := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := etcdClient.Get(ctx, key)
		if err != nil {
			return &ErrEtcdOperation{Key: key, Err: err}
			// return fmt.Errorf("failed to get key %s from etcd: %v", key, err)
		}

		if len(resp.Kvs) == 0 {
			// If key not found, return an error to trigger a retry
			return &ErrKeyNotFound{Key: key}
			// return fmt.Errorf("key %s not found in etcd", key)
		}

		value = string(resp.Kvs[0].Value)
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
		logger.Sugar().Errorf("failed retrieving key from etcd: %v", err)
		return "", err
	}

	return value, nil
}

func GetByteValueFromEtcd(etcdClient *clientv3.Client, key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := etcdClient.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get key %s from etcd: %v", key, err)
	}

	if len(resp.Kvs) == 0 {
		fmt.Printf("key %s not found in etcd", key)
		return nil, nil
	}

	return resp.Kvs[0].Value, nil
}
func GetKeyValueMap(etcdClient *clientv3.Client, pathName string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := etcdClient.Get(ctx, pathName, clientv3.WithPrefix())
	if err != nil {
		logger.Sugar().Errorw("failed to get keys with prefix %s from etcd: %v", pathName, err)
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		logger.Sugar().Errorw("no keys with prefix %s found in etcd", pathName)
		return nil, err
	}

	values := make(map[string]string)
	for _, kv := range resp.Kvs {
		values[string(kv.Key)] = string(kv.Value)
	}
	return values, nil
}

// GetAndUnmarshalJSON retrieves a JSON value from etcd and unmarshals it into the target struct.
// - T should be a pointer to a struct type.
// - etcdClient is an instance of the etcd client.
// - key is the etcd key where the JSON value is stored.
// - target should be a pointer to an instance of the target struct.
func GetAndUnmarshalJSON[T any](etcdClient *clientv3.Client, key string, target *T) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get the value from etcd.
	resp, err := etcdClient.Get(ctx, key)
	if err != nil {
		logger.Sugar().Errorf("failed to get value from etcd: %v", err)
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		logger.Sugar().Warnw("no value found for", "key", key)
		return nil, nil //fmt.Errorf("no value found for key %v", key)
	}

	// Unmarshal the JSON value into the target struct.
	err = json.Unmarshal(resp.Kvs[0].Value, target)
	if err != nil {
		logger.Sugar().Errorf("failed to unmarshal JSON: %v", err)
		return nil, err
	}

	return resp.Kvs[0].Value, nil
}

// T should be a struct type.
// Pass a full path (like /microservices/) and get a Map back of all entries in that folder.
//
// See etcd_test.go for examples
func GetAndUnmarshalJSONMap[T any](etcdClient *clientv3.Client, prefix string) (map[string]T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Get all key-value pairs under the specified prefix.
	resp, err := etcdClient.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get values from etcd: %v", err)
	}

	// Initialize an empty map to store the unmarshaled structs.
	result := make(map[string]T)

	// Iterate through the key-value pairs and unmarshal the values into structs.
	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		// Extract the map key from the etcd key.
		mapKey := strings.TrimPrefix(key, prefix)
		if mapKey == "" {
			continue
		}

		// Unmarshal the JSON value into the target struct.
		var target T
		err = json.Unmarshal(kv.Value, &target)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON for key %s: %v", key, err)
		}

		// Add the unmarshaled struct to the result map.
		result[mapKey] = target
	}

	return result, nil
}

// Get all values from a certain prefix, convert these to a list of the given type.
func GetPrefixListEtcd[T any](client KVGetter, prefix string, target *T) ([]*T, error) {
	// func GetPrefixListEtcd[T any](client *clientv3.Client, prefix string, target *T) ([]T, error) {
	ctx := context.Background()
	resp, err := client.Get(ctx, prefix, clientv3.WithPrefix())

	var targets []*T
	if err != nil {
		logger.Sugar().Errorw("Failed to get from etcd: %v", err)
		return nil, err
	}

	for _, kv := range resp.Kvs {
		if err := json.Unmarshal(kv.Value, &target); err != nil {
			logger.Sugar().Errorw("Failed to unmarshal value for key %s: %v", kv.Key, err)
			continue
		}

		targets = append(targets, target)
		target = nil
	}

	return targets, err
}
