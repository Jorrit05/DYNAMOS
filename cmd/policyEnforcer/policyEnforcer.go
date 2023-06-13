package main

import (
	"net/http"

	"github.com/Jorrit05/micro-recomposer/pkg/lib"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	log, logFile                  = lib.InitLogger(logFileLocation, serviceName)
	etcdClient   *clientv3.Client = lib.GetEtcdClient()
)

type RequestType struct {
	Type             string   `json:"type"`
	RequiredServices []string `json:"requiredServices"`
	OptionalServices []string `json:"optionalServices"`
}

type Archetype struct {
	Name            string `json:"name"`
	ComputeProvider string `json:"computeProvider"`
	ResultRecipient string `json:"resultRecipient"`
}

type MicroserviceMetadata struct {
	Name           string   `json:"name"`
	Label          string   `json:"label"`
	AllowedOutputs []string `json:"allowedOutputs"`
}

type Named interface {
	GetName() string
}

func (a Archetype) GetName() string {
	return a.Name
}

func main() {

	defer logFile.Close()
	defer etcdClient.Close()
	defer lib.HandlePanicAndFlushLogs(log, logFile)

	registerPolicyEnforcerConfiguration()

	mux := http.NewServeMux()
	mux.HandleFunc("/archetypes/", archetypesHandler(etcdClient, "/archetypes"))

	// mux.HandleFunc("/archetypes/", genericGetHandler)
	log.Info("Starting http server on 8081/30011")
	go func() {
		if err := http.ListenAndServe(":8081", mux); err != nil {
			log.Fatalf("Error starting HTTP server: %s", err)
		}
	}()

	select {}

}
