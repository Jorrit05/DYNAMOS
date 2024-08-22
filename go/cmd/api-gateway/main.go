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
	socketio "github.com/googollee/go-socket.io"
	"github.com/gorilla/handlers"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.opencensus.io/plugin/ochttp"
	"google.golang.org/grpc"
)

var (
	logger                                = lib.InitLogger(logLevel)
	etcdClient           *clientv3.Client = etcd.GetEtcdClient(etcdEndpoints)
	conn                 *grpc.ClientConn
	receiveMutex         = &sync.Mutex{}
	requestApprovalMap   = make(map[string]chan validation)
	requestApprovalMutex = &sync.Mutex{}
	policyUpdateMap      = make(map[string]map[string]*pb.CompositionRequest)
	c                    pb.RabbitMQClient
)

type validation struct {
	response     *pb.RequestApprovalResponse
	localContext context.Context
}

// Sequence of steps:
// 1. StartWebSocketServer (in a separate goroutine)
//   - This function starts a websocket server that listens for incoming messages from the frontend
//
// 2. Initialize gRPC connection
// 3. Initialize sidecar messaging
// 4. Start consuming messages from the sidecar
// 5. Start HTTP server (in a separate goroutine)
//   - This function starts an HTTP server that listens for incoming requests from the 'public' API
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

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	mux := http.NewServeMux()

	apiMux := http.NewServeMux()
	apiMux.Handle("/requestApproval", &ochttp.Handler{Handler: requestHandler()})
	apiMux.Handle("/getAvailableProviders", &ochttp.Handler{Handler: availableProvidersHandler()})
	// go socketServer(apiMux)
	// server := socketio.NewServer(&engineio.Options{
	// 	Transports: []transport.Transport{
	// 		&polling.Transport{
	// 			CheckOrigin: allowOriginFunc,
	// 		},
	// 		&websocket.Transport{
	// 			CheckOrigin: allowOriginFunc,
	// 		},
	// 	},
	// })
	server := socketio.NewServer(nil)

	server.OnError("/", func(s socketio.Conn, e error) {
		logger.Sugar().Infow("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		logger.Sugar().Infow("closed", reason)
	})

	// Serve Socket.IO requests at "/socket.io/" prefix
	apiMux.Handle("/socket.io/", server)
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

// Easier to get running with CORS. Thanks for help @Vindexus and @erkie
var allowOriginFunc = func(r *http.Request) bool {
	return true
}
