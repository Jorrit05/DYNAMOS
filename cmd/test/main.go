package main

import (
	"fmt"
	"sort"

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

	// sliceA := []string{"unl1_agent", "unl2_agent", "unl3"}
	// sliceB := []string{"unl2_agent", "unl5"}

	sliceA := []string{"apple", "banana", "cherry", "apple", "grape"}
	sliceB := []string{"banana", "cherry", "kiwi", "mango"}
	sort.Strings(sliceA)
	// sliceA := []string{"apple", "banana", "cherry", "apple", "grape"}
	// sliceB := []string{"banana", "cherry", "kiwi", "mango"}
	matched, notMatched := lib.SliceIntersectAndDifference(sliceA, sliceB)
	fmt.Println("Matched:", matched)
	fmt.Println("Not Matched:", notMatched)
	// for k, v := range service.Services {
	// 	fmt.Println(k)
	// 	fmt.Println(v)
	// 	fmt.Println("--------------------------")
	// }

	// registerJSONArray[lib.ArcheType](archetypesJSON, &lib.ArcheTypes{}, etcdClient, "/reasoner/archetype_config")

	// 	// var err error = nil
	// 	payload := CreateServicePayload{
	// 		ImageName: "my-image",
	// 		Tag:       "latest",
	// 		EnvVars:   map[string]string{"ENV1": "value1", "ENV2": "value2"},
	// 		Networks:  []string{"network1", "network2"},
	// 		Secrets:   []string{"secret1", "secret2"},
	// 		Volumes:   map[string]string{"volume1": "/path1", "volume2": "/path2"},
	// 		Ports:     map[string]string{"8080": "80"},
	// 		Deploy: Deploy{
	// 			Replicas:  2,
	// 			Placement: Placement{Constraints: []string{"node.role == worker"}},
	// 			Resources: Resources{
	// 				Reservations: Resource{Memory: "100M"},
	// 				Limits:       Resource{Memory: "200M"},
	// 			},
	// 		},
	// 	}

	// 	jsonData, err := json.Marshal(payload)
	// 	if err != nil {
	// 		fmt.Printf("Error marshaling payload to JSON:", err)
	// 	}

	// 	fmt.Printf(string(jsonData))

	// defer logFile.Close()
	// mux := http.NewServeMux()
	// mux.HandleFunc("/", handler)
	// go func() {
	// 	fmt.Println("ListenAndServe: 1")

	// 	if err := http.ListenAndServe(":3000", mux); err != nil {

	// 		log.Fatalf("Error starting HTTP server: %s", err)
	// 	}
	// }()
	// fmt.Println("3")
	// select {}
}
