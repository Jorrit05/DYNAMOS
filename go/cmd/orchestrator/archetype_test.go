package main

// import (
// 	"context"
// 	"fmt"
// 	"testing"

// 	"github.com/Jorrit05/DYNAMOS/pkg/lib"
// 	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
// 	clientv3 "go.etcd.io/etcd/client/v3"
// 	"go.etcd.io/etcd/mvcc/mvccpb"

// 	"gotest.tools/assert"
// )

// type mockKVGetter struct {
// 	response *clientv3.GetResponse
// 	err      error
// }

// func (m *mockKVGetter) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
// 	return m.response, m.err
// }

// var (
// 	UVA1 = &pb.DataProvider{
// 		Archetypes: []string{"computeToData", "dataThroughTtp"},
// 	}
// 	VU1 = &pb.DataProvider{
// 		Archetypes: []string{"computeToData", "dataThroughTtp"},
// 	}

// 	test1 = &pb.ValidationResponse{
// 		Type:            "validationResponse",
// 		RequestType:     "sqlDataRequest",
// 		RequestApproved: true,
// 		Options: map[string]bool{
// 			"aggregate": false,
// 		},
// 		ValidArchetypes: &pb.UserArchetypes{
// 			Archetypes: map[string]*pb.UserAllowedArchetypes{
// 				"UVA": {Archetypes: []string{"computeToData", "dataThroughTtp"}},
// 				"VU":  {Archetypes: []string{"computeToData", "dataThroughTtp"}},
// 			}},
// 		User: &pb.User{
// 			Id:       "1234",
// 			UserName: "jorrit.stutterheim@cloudnation.nl",
// 		},
// 		ValidDataproviders: map[string]*pb.DataProvider{
// 			"UVA": UVA1,
// 			"VU":  VU1,
// 		},
// 		InvalidDataproviders: []string{},
// 	}

// 	agentDetails1 = map[string]lib.AgentDetails{
// 		"UVA": {Name: "UVA", RoutingKey: "UVA-in", Dns: "uva.uva.svc.cluster.local"},
// 		"VU":  {Name: "VU", RoutingKey: "VU-in", Dns: "vu.vu.svc.cluster.local"},
// 	}
// )

// func TestSelectArchetype(t *testing.T) {
// 	tests := []struct {
// 		validationResponse      *pb.ValidationResponse
// 		authorizedDataProviders map[string]lib.AgentDetails
// 		expected                string
// 	}{
// 		{test1, agentDetails1, "dataThroughTtp"},
// 	}

// 	// Prepare a mock response:
// 	kvs := []*mvccpb.KeyValue{
// 		{
// 			Key:   []byte("/my/prefix/1"),
// 			Value: []byte(`{"id": "1", "name": "one"}`),
// 		},
// 		{
// 			Key:   []byte("/my/prefix/2"),
// 			Value: []byte(`{"id": "2", "name": "two"}`),
// 		},
// 	}

// 	mockResponse := &clientv3.GetResponse{Kvs: kvs}

// 	// Create the mock client:
// 	mockClient := &mockKVGetter{
// 		response: mockResponse,
// 		err:      nil, // no error
// 	}

// 	for _, test := range tests {
// 		t.Run("", func(t *testing.T) {
// 			result, err := chooseArchetype(test.validationResponse, test.authorizedDataProviders)
// 			if err != nil {
// 				fmt.Printf("err: %v", err)
// 			}
// 			assert.Equal(t, test.expected, result)
// 		})
// 	}
// }
