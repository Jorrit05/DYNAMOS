package main

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcdHandler func() error

func watchEtcdDirectory(client *clientv3.Client, etcdHandler, pathName string) {

	// Watch for new entries in the /agents/ directory
	watchChan := client.Watch(context.Background(), pathName, clientv3.WithPrefix())

	// Loop to process watch events
	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			// Check if the event is a new entry
			if event.Type == clientv3.EventTypePut {
				log.Info("Entry added or updated: %s = %s\n", event.Kv.Key, event.Kv.Value)
			} else if event.Type == clientv3.EventTypeDelete {
				log.Info("Entry deleted: %s\n", event.Kv.Key)
			}
		}
	}

}
