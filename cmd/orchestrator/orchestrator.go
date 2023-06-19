package main

import (
	"context"
	"net/http"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"github.com/gorilla/handlers"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	logger                      = lib.InitLogger()
	etcdClient *clientv3.Client = lib.GetEtcdClient(etcdEndpoints)
)

func main() {
	defer logger.Sync() // flushes buffer, if any
	defer etcdClient.Close()
	// Set up a connection to the server.

	conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Sugar().Fatalw("did not connect to grpc server: %v", err)

	}
	defer conn.Close()
	c := pb.NewSideCarClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.StartService(ctx, &pb.ServiceRequest{ServiceName: serviceName, RoutingKey: serviceName, QueueAutoDelete: false, StartConsuming: false})
	if err != nil {
		logger.Sugar().Fatalw("could not greet: %v", err)
	}
	logger.Sugar().Infow("Greeting: %s", r.GetMessage())
	registerPolicyEnforcerConfiguration()
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	// Connect to AMQ queue, declare own routingKey as queue, start listening for messages
	// _, conn, _, err := lib.SetupConnection(serviceName, serviceName, false)
	// if err != nil {
	// 	logger.Sugar().Fatalw("Failed to setup proper connection to RabbitMQ: %v", err)
	// }
	// defer conn.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/archetypes/", archetypesHandler(etcdClient, "/archetypes"))
	mux.HandleFunc("/requesttypes/", requestTypesHandler(etcdClient, "/requestTypes"))
	mux.HandleFunc("/requestTypes/", requestTypesHandler(etcdClient, "/requestTypes"))
	mux.HandleFunc("/microservices/", microserviceMetadataHandler(etcdClient, "/microservices"))
	mux.HandleFunc("/policyEnforcer/", agreementsHandler(etcdClient, "/policyEnforcer"))

	mux.HandleFunc("/requestapproval", requestApprovalHandler())

	logger.Sugar().Infow("Starting http server on %s/30011", port)
	go func() {
		if err := http.ListenAndServe(port, handlers.CORS(originsOk, headersOk, methodsOk)(mux)); err != nil {
			logger.Sugar().Fatalw("Error starting HTTP server: %s", err)
		}
	}()

	select {}

}
