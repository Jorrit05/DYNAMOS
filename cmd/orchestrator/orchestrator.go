package main

import (
	"fmt"
	"net/http"

	"github.com/Jorrit05/micro-recomposer/pkg/lib"
	"github.com/gorilla/handlers"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	log, logFile                  = lib.InitLogger(logFileLocation, serviceName)
	etcdClient   *clientv3.Client = lib.GetEtcdClient(etcdEndpoints)
)

func main() {

	defer logFile.Close()
	defer etcdClient.Close()
	defer lib.HandlePanicAndFlushLogs(log, logFile)

	registerPolicyEnforcerConfiguration()
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	mux := http.NewServeMux()
	mux.HandleFunc("/archetypes/", archetypesHandler(etcdClient, "/archetypes"))
	mux.HandleFunc("/requesttypes/", requestTypesHandler(etcdClient, "/requestTypes"))
	mux.HandleFunc("/requestTypes/", requestTypesHandler(etcdClient, "/requestTypes"))
	mux.HandleFunc("/microservices/", microserviceMetadataHandler(etcdClient, "/microservices"))
	mux.HandleFunc("/policyEnforcer/", agreementsHandler(etcdClient, "/policyEnforcer"))

	mux.HandleFunc("/requestapproval", requestApprovalHandler())

	log.Info(fmt.Sprintf("Starting http server on %s/30011", port))
	go func() {
		if err := http.ListenAndServe(port, handlers.CORS(originsOk, headersOk, methodsOk)(mux)); err != nil {
			log.Fatalf("Error starting HTTP server: %s", err)
		}
	}()

	select {}

}
