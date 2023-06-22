package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func GetEtcdClient(endpoints string) *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(endpoints, ","),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logger.Sugar().Fatalw("Error in creating ETCD client %v", err)
	}

	return cli
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

func CancelLeaseEtcdKey(key string) {
	etcdLeaseMap[key].cancel()
}

func GetValueFromEtcd(etcdClient *clientv3.Client, key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := etcdClient.Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to get key %s from etcd: %v", key, err)
	}

	if len(resp.Kvs) == 0 {
		fmt.Printf("key %s not found in etcd", key)
		return "", nil
	}

	value := string(resp.Kvs[0].Value)
	return value, nil
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
		logger.Sugar().Errorw("failed to get value from etcd: %v", err)
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		logger.Sugar().Errorw("no value found for key: %s", key)
		return nil, err
	}

	// Unmarshal the JSON value into the target struct.
	err = json.Unmarshal(resp.Kvs[0].Value, target)
	if err != nil {
		logger.Sugar().Errorw("failed to unmarshal JSON: %v", err)
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

// Get all values from a certain prefix, convert these to a list of the given type.
func GetPrefixListEtcd[T any](client KVGetter, prefix string, target *T) ([]T, error) {
	// func GetPrefixListEtcd[T any](client *clientv3.Client, prefix string, target *T) ([]T, error) {
	ctx := context.Background()
	resp, err := client.Get(ctx, prefix, clientv3.WithPrefix())

	var targets []T
	if err != nil {
		logger.Sugar().Errorw("Failed to get from etcd: %v", err)
		return nil, err
	}

	for _, kv := range resp.Kvs {
		if err := json.Unmarshal(kv.Value, &target); err != nil {
			logger.Sugar().Errorw("Failed to unmarshal value for key %s: %v", kv.Key, err)
			continue
		}

		targets = append(targets, *target)
		target = nil
	}

	return targets, err
}
