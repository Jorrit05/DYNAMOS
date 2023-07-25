package main

import (
	"context"

	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	logger        = lib.InitLogger(zap.DebugLevel)
	conn          *grpc.ClientConn
	etcdEndpoints = "http://localhost:30005"

	etcdClient *clientv3.Client = etcd.GetEtcdClient(etcdEndpoints)

	serviceName string = "test"
	grpcAddr           = "localhost:50052"
)

func main() {
	etcd.PutEtcdWithGrant(context.Background(), etcdClient, "mykey", "some val", 3)

	select {}
}
