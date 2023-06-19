package main

import (
	"context"
	"fmt"

	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func watchKeyChanges(client *clientv3.Client, key string) {
	// Create a context with no timeout
	ctx := context.Background()

	// Start a new watch using the client
	watchChan := client.Watch(ctx, key)

	// Process events received on the watch channel
	for watchResponse := range watchChan {
		for _, event := range watchResponse.Events {
			switch event.Type {
			case clientv3.EventTypePut:
				fmt.Printf("Key %q was updated to value %q\n", event.Kv.Key, event.Kv.Value)
			case clientv3.EventTypeDelete:
				fmt.Printf("Key %q was deleted\n", event.Kv.Key)
			}
		}
	}
}

func test() {
	// Connect to etcd
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	defer client.Close()

	// The key to watch
	watchKey := "/config/myKey"

	// Use a wait group to block the main goroutine
	var wg sync.WaitGroup
	wg.Add(1)

	// Start the watch in a separate goroutine
	go func() {
		watchKeyChanges(client, watchKey)
		wg.Done()
	}()

	// Block main goroutine until watchKeyChanges is done
	wg.Wait()
}
