package main

import (
	"fmt"
	"strings"

	"github.com/Jorrit05/micro-recomposer/pkg/lib"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	serviceName string = "test_service"

	log, logFile = lib.InitLogger(serviceName)
	etcdClient   *clientv3.Client
)

func main() {
	// etcdClient, err := clientv3.New(clientv3.Config{
	// 	Endpoints:   []string{"localhost:2379"},
	// 	DialTimeout: 5 * time.Second,
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// hostname := "unl1_agent"

	fileLocation := "/Users/jorrit/Documents/master-software-engineering/thesis/micro-recomposer/stack/agents.yaml"

	var service lib.MicroServiceData = lib.UnmarshalStackFile(fileLocation)

	for _, v := range service.Services {
		// Map of network result would be:
		// network: core_network and list of aliases: unl1_agent
		// network: unl_1 list of aliases: unl1_agent
		for key, value := range v.Networks {
			fmt.Println("network: " + key)
			fmt.Println("list of aliases: " + strings.Join(value.Aliases, ","))
		}
		break
	}
}
