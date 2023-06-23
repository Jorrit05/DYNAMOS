package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/api"
	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	"github.com/gorilla/handlers"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

var (
	logger                       = lib.InitLogger(logLevel)
	etcdClient  *clientv3.Client = etcd.GetEtcdClient(etcdEndpoints)
	c           pb.SideCarClient
	conn        *grpc.ClientConn
	agentConfig lib.AgentDetails
)

func main() {
	if !local {
		serviceName = os.Getenv("DATA_STEWARD_NAME")
	}

	c, conn = lib.InitializeRabbit(grpcAddr, &pb.ServiceRequest{ServiceName: fmt.Sprintf("%s-in", serviceName), RoutingKey: fmt.Sprintf("%s-in", serviceName), QueueAutoDelete: false})
	defer conn.Close()
	registerAgent()

	// Define a WaitGroup
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		startConsumingWithRetry(c, fmt.Sprintf("%s-in", serviceName), 5, 5*time.Second)

		wg.Done() // Decrement the WaitGroup counter when the goroutine finishes
	}()

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	agentMux := http.NewServeMux()
	agentMux.HandleFunc("/agent/v1/sqlDataRequest", sqlDataRequestHandler())

	wrappedAgentMux := authMiddleware(agentMux)

	mux := http.NewServeMux()
	mux.Handle("/agent/v1/sqlDataRequest", wrappedAgentMux)

	logger.Sugar().Infow("Starting http server on: ", "port", port)
	go func() {
		if err := http.ListenAndServe(port, api.LogMiddleware(handlers.CORS(originsOk, headersOk, methodsOk)(mux))); err != nil {
			logger.Sugar().Fatalw("Error starting HTTP server: %s", err)
		}
	}()

	wg.Wait()

}
