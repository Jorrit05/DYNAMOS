package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	logger        = lib.InitLogger(zap.DebugLevel)
	conn          *grpc.ClientConn
	etcdEndpoints = "http://localhost:30005"

	etcdClient *clientv3.Client = etcd.GetEtcdClient(etcdEndpoints)

	serviceName  string = "test"
	grpcAddr            = "localhost:50051"
	c            pb.SideCarClient
	receiveMutex = &sync.Mutex{}
)

func deleteJobInfo(userName string) {
	key := fmt.Sprintf("/agents/jobs/UVA/%s", userName)

	jobNames, err := etcd.GetKeysFromPrefix(etcdClient, key, etcd.WithMaxElapsedTime(2*time.Second))
	if err != nil {
		logger.Sugar().Warnf("error get agents: %v", err)
	}

	for _, job := range jobNames {
		fmt.Printf("job: %s", job)
	}
}

func getAvailableAgents() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get the value from etcd.
	resp, err := etcdClient.Get(ctx, "/agents/online", clientv3.WithPrefix())
	if err != nil {
		// logger.Sugar().Errorf("failed to get value from etcd: %v", err)
		fmt.Printf("failed to get value from etcd: %v", err)
	}

	// Initialize an empty map to store the unmarshaled structs.
	result := make(map[string]lib.AgentDetails)
	// Iterate through the key-value pairs and unmarshal the values into structs.
	for _, kv := range resp.Kvs {
		var target lib.AgentDetails
		err = json.Unmarshal(kv.Value, &target)
		if err != nil {
			// return nil, fmt.Errorf("failed to unmarshal JSON for key %s: %v", key, err)
		}
		result[string(target.Name)] = target
	}

	fmt.Printf("result: %v", result)
}
func main() {
	// deleteJobInfo("jorrit.stutterheim@cloudnation.nl")
	getAvailableAgents()
	// conn = lib.GetGrpcConnection(grpcAddr)
	conn = lib.GetGrpcConnection(grpcAddr)

	c = lib.InitializeSidecarMessaging(conn, &pb.InitRequest{ServiceName: fmt.Sprintf("%s-in", serviceName), RoutingKey: fmt.Sprintf("%s-in", serviceName), QueueAutoDelete: false})

	// Define a WaitGroup
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		lib.StartConsumingWithRetry(serviceName, c, fmt.Sprintf("%s-in", serviceName), handleIncomingMessages, 5, 5*time.Second, receiveMutex)

		wg.Done() // Decrement the WaitGroup counter when the goroutine finishes
	}()

	wg.Wait()

}

func handleIncomingMessages(ctx context.Context, grpcMsg *pb.SideCarMessage) error {
	logger.Debug("Start handleIncomingMessages")
	switch grpcMsg.Type {
	case "requestApproval":
		logger.Debug("Start requestApproval")

		sendMicroserviceComm(c)
	default:
		logger.Sugar().Warnf("Uknown Type: %s", grpcMsg.Type)
		sendMicroserviceComm(c)

	}

	return nil
}
