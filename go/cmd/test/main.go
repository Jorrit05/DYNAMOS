package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/api"
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

var (
	UVA1 = &pb.DataProvider{
		Archetypes: []string{"computeToData", "dataThroughTtp"},
	}
	VU1 = &pb.DataProvider{
		Archetypes: []string{"computeToData", "dataThroughTtp"},
	}

	test1 = &pb.ValidationResponse{
		Type:            "validationResponse",
		RequestType:     "sqlDataRequest",
		RequestApproved: true,
		ValidArchetypes: &pb.UserArchetypes{
			Archetypes: map[string]*pb.UserAllowedArchetypes{
				"UVA": {Archetypes: []string{"computeToData", "dataThroughTtp"}},
				"VU":  {Archetypes: []string{"computeToData", "dataThroughTtp"}},
			}},
		User: &pb.User{
			Id:       "1234",
			UserName: "jorrit.stutterheim@cloudnation.nl",
		},
		ValidDataproviders: map[string]*pb.DataProvider{
			"UVA": UVA1,
			"VU":  VU1,
		},
		InvalidDataproviders: []string{},
	}

	agentDetails1 = map[string]lib.AgentDetails{
		"UVA": {Name: "UVA", RoutingKey: "UVA-in", Dns: "uva.uva.svc.cluster.local"},
		"VU":  {Name: "VU", RoutingKey: "VU-in", Dns: "vu.vu.svc.cluster.local"},
	}
)

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

	fmt.Println(test1.ValidArchetypes.Archetypes["UVA"].Archetypes)
	fmt.Println(test1.ValidArchetypes.Archetypes["VU"].Archetypes)

	os.Exit(0)

	fmt.Printf("Start %s\n", serviceName)

	var archeTypes = &api.Archetype{}

	archeTypess, _ := etcd.GetPrefixListEtcd(etcdClient, "/archetypes", archeTypes)

	fmt.Println(archeTypess)
	for _, archeType := range archeTypess {
		fmt.Println(archeType)
	}
	fmt.Println(len(archeTypess))
	os.Exit(0)
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
