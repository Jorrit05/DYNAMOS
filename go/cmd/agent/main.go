package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/api"
	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	"github.com/gorilla/handlers"
	"go.opencensus.io/plugin/ochttp"
	batchv1 "k8s.io/api/batch/v1"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

var (
	logger                           = lib.InitLogger(logLevel)
	etcdClient      *clientv3.Client = etcd.GetEtcdClient(etcdEndpoints)
	c               pb.RabbitMQClient
	conn            *grpc.ClientConn
	agentConfig     lib.AgentDetails
	mutex           = &sync.Mutex{}
	ttpMutex        = &sync.Mutex{}
	jobMutex        = &sync.Mutex{}
	waitingJobMutex = &sync.Mutex{}
	queueInfoMutex  = &sync.Mutex{}

	responseMap   = make(map[string]chan dataResponse)
	thirdPartyMap = make(map[string]string)
	jobCounter    = make(map[string]int)
	waitingJobMap = make(map[string]*waitingJob)
	queueInfoMap  = make(map[string]*pb.QueueInfo)
	receiveMutex  = &sync.Mutex{}

	clientSet = getKubeClient()
)

type dataResponse struct {
	response     *pb.MicroserviceCommunication
	localContext context.Context
}

type waitingJob struct {
	job              *batchv1.Job
	nrOfDataStewards int
}

func main() {
	serviceName = os.Getenv("DATA_STEWARD_NAME")

	if local && serviceName == "SURF" {
		port = ":8083"
	}

	_, err := lib.InitTracer(serviceName)
	if err != nil {
		logger.Sugar().Fatalf("Failed to create ocagent-exporter: %v", err)
	}

	conn = lib.GetGrpcConnection(grpcAddr)
	defer conn.Close()
	c = lib.InitializeSidecarMessaging(conn, &pb.InitRequest{ServiceName: fmt.Sprintf("%s-in", serviceName), RoutingKey: fmt.Sprintf("%s-in", serviceName), QueueAutoDelete: false})

	registerAgent()

	// Define a WaitGroup
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		lib.StartConsumingWithRetry(serviceName, c, fmt.Sprintf("%s-in", serviceName), handleIncomingMessages, 5, 5*time.Second, receiveMutex)

		wg.Done() // Decrement the WaitGroup counter when the goroutine finishes
	}()

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	agentMux := http.NewServeMux()
	agentMux.Handle(fmt.Sprintf("/agent/v1/sqlDataRequest/%s", strings.ToLower(serviceName)), &ochttp.Handler{Handler: sqlDataRequestHandler()})

	// apiMux.Handle("/archetypes/", &ochttp.Handler{Handler: archetypesHandler(etcdClient, "/archetypes")})

	wrappedAgentMux := authMiddleware(agentMux)

	mux := http.NewServeMux()
	mux.Handle(fmt.Sprintf("/agent/v1/sqlDataRequest/%s", strings.ToLower(serviceName)), wrappedAgentMux)

	logger.Sugar().Infow("Starting http server on: ", "port", port)
	go func() {
		if err := http.ListenAndServe(port, api.LogMiddleware(handlers.CORS(originsOk, headersOk, methodsOk)(mux))); err != nil {
			logger.Sugar().Fatalw("Error starting HTTP server: %s", err)
		}
	}()

	wg.Wait()

}
