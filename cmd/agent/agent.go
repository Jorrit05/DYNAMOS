package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

var (
	logger                      = lib.InitLogger()
	etcdClient *clientv3.Client = etcd.GetEtcdClient(etcdEndpoints)
	c          pb.SideCarClient
	conn       *grpc.ClientConn
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

	wg.Wait()

}
