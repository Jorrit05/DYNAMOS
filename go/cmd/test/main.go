package main

import (
	"context"
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

func main() {
	fmt.Println(name)
	// deleteJobInfo("jorrit.stutterheim@cloudnation.nl")
	// conn = lib.GetGrpcConnection(grpcAddr)

	// c = lib.InitializeSidecarMessaging(conn, &pb.InitRequest{ServiceName: fmt.Sprintf("%s-in", serviceName), RoutingKey: fmt.Sprintf("%s-in", serviceName), QueueAutoDelete: false})

	// // Define a WaitGroup
	// var wg sync.WaitGroup
	// wg.Add(1)

	// go func() {
	// 	lib.StartConsumingWithRetry(serviceName, c, fmt.Sprintf("%s-in", serviceName), handleIncomingMessages, 5, 5*time.Second, receiveMutex)

	// 	wg.Done() // Decrement the WaitGroup counter when the goroutine finishes
	// }()

	// wg.Wait()

}

func handleIncomingMessages(ctx context.Context, grpcMsg *pb.SideCarMessage) error {
	logger.Debug("Start handleIncomingMessages")
	switch grpcMsg.Type {
	case "requestApproval":
		sendMicroserviceComm(c)

	}
	return nil
}
