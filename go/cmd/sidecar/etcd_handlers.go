package main

import (
	"context"
	"encoding/json"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
)

var (
	etcdClient *clientv3.Client
)

func (s *server) InitEtcd(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	etcdClient = etcd.GetEtcdClient(etcdEndpoints)
	return &emptypb.Empty{}, nil
}

func (s *server) GetDatasetMetadata(ctx context.Context, in *pb.EtcdKey) (*pb.Dataset, error) {
	logger.Debug("Starting GetDatasetMetadata")
	value, err := etcd.GetByteValueFromEtcd(etcdClient, in.Path)
	if err != nil {
		logger.Sugar().Errorf("Error getting json from etcd: %v", err)
		return &pb.Dataset{}, err
	}
	if value == nil {
		logger.Sugar().Warnf("Key not in Etcd: %v", value)
		return &pb.Dataset{}, err
	}

	dataset := &pb.Dataset{}
	err = json.Unmarshal(value, dataset)
	if err != nil {
		logger.Sugar().Warnf("Error json unmarshal dataset: %v", value)
		return &pb.Dataset{}, err
	}

	logger.Info(dataset.Name)
	return dataset, nil
}
