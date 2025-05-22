package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/api"
	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"github.com/gorilla/handlers"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.opencensus.io/plugin/ochttp"
	"google.golang.org/grpc"
)

var (
	logger                             = lib.InitLogger(logLevel)
	etcdClient        *clientv3.Client = etcd.GetEtcdClient(etcdEndpoints)
	conn              *grpc.ClientConn
	receiveMutex      = &sync.Mutex{}
	policyUpdateMutex = &sync.Mutex{}
	policyUpdateMap   = make(map[string]map[string]*pb.CompositionRequest)
	c                 pb.RabbitMQClient
)

type validation struct {
	response     *pb.ValidationResponse
	localContext context.Context
}

func main() {
	defer logger.Sync() // flushes buffer, if any
	defer etcdClient.Close()

	_, err := lib.InitTracer(serviceName)
	if err != nil {
		logger.Sugar().Fatalf("Failed to create ocagent-exporter: %v", err)
	}

	conn = lib.GetGrpcConnection(grpcAddr)
	defer conn.Close()
	c = lib.InitializeSidecarMessaging(conn, &pb.InitRequest{ServiceName: fmt.Sprintf("%s-in", serviceName), RoutingKey: fmt.Sprintf("%s-in", serviceName), QueueAutoDelete: false})

	// Define a WaitGroup
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		lib.StartConsumingWithRetry(serviceName, c, fmt.Sprintf("%s-in", serviceName), handleIncomingMessages, 5, 5*time.Second, receiveMutex)
		wg.Done() // Decrement the WaitGroup counter when the goroutine finishes
	}()

	go registerPolicyEnforcerConfiguration()
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	mux := http.NewServeMux()

	apiMux := http.NewServeMux()
	// apiMux.HandleFunc("/archetypes", ochttp.Handler{Handler: http.HandlerFunc(archetypesHandler(etcdClient, "/archetypes"))})
	// apiMux.Handle("/archetypes", ochttp.Handler{Handler: http.HandlerFunc(archetypesHandler(etcdClient, "/archetypes"))})
	apiMux.Handle("/archetypes", &ochttp.Handler{Handler: archetypesHandler(etcdClient, "/archetypes")})
	apiMux.Handle("/archetypes/", &ochttp.Handler{Handler: archetypesHandler(etcdClient, "/archetypes")})

	apiMux.Handle("/requesttypes", &ochttp.Handler{Handler: requestTypesHandler(etcdClient, "/requestTypes")})
	apiMux.Handle("/requestTypes", &ochttp.Handler{Handler: requestTypesHandler(etcdClient, "/requestTypes")})

	apiMux.Handle("/requesttypes/", &ochttp.Handler{Handler: requestTypesHandler(etcdClient, "/requestTypes")})
	apiMux.Handle("/requestTypes/", &ochttp.Handler{Handler: requestTypesHandler(etcdClient, "/requestTypes")})

	apiMux.Handle("/microservices", &ochttp.Handler{Handler: microserviceMetadataHandler(etcdClient, "/microservices")})
	apiMux.Handle("/microservices/", &ochttp.Handler{Handler: microserviceMetadataHandler(etcdClient, "/microservices")})

	apiMux.Handle("/updateEtc", &ochttp.Handler{Handler: updateEtc()})

	apiMux.Handle("/policyEnforcer", &ochttp.Handler{Handler: agreementsHandler(etcdClient, "/policyEnforcer")})
	apiMux.Handle("/policyEnforcer/", &ochttp.Handler{Handler: agreementsHandler(etcdClient, "/policyEnforcer")})

	logger.Info(apiVersion) // prints /api/v1

	mux.Handle(apiVersion+"/", http.StripPrefix(apiVersion, apiMux))

	logger.Sugar().Infow("Starting http server on: ", "port", port)
	go func() {
		if err := http.ListenAndServe(port, api.LogMiddleware(handlers.CORS(originsOk, headersOk, methodsOk)(mux))); err != nil {
			logger.Sugar().Fatalw("Error starting HTTP server: %s", err)
		}
	}()

	wg.Wait()
}
