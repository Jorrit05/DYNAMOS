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

func main() {

	defer logFile.Close()
	defer etcdClient.Close()
	defer lib.HandlePanicAndFlushLogs(log, logFile)

	registerPolicyEnforcerConfiguration()

	mux := http.NewServeMux()
	mux.HandleFunc("/archetypes/", archetypesHandler(etcdClient, "/archetypes"))
	mux.HandleFunc("/requesttypes/", requestTypesHandler(etcdClient, "/requestTypes"))
	mux.HandleFunc("/requestTypes/", requestTypesHandler(etcdClient, "/requestTypes"))
	mux.HandleFunc("/microservices/", microserviceMetadataHandler(etcdClient, "/microservices"))

	// mux.HandleFunc("/archetypes/", genericGetHandler)
	log.Info("Starting http server on 8081/30011")
	go func() {
		if err := http.ListenAndServe(":8081", mux); err != nil {
			log.Fatalf("Error starting HTTP server: %s", err)
		}
	}()

	select {}

}
