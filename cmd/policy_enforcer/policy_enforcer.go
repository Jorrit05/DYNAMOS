package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

var (
	logger                      = lib.InitLogger()
	etcdClient *clientv3.Client = lib.GetEtcdClient(etcdEndpoints)
	c          pb.SideCarClient
	conn       *grpc.ClientConn
)

func main() {

	c, conn = lib.InitializeRabbit(grpcAddr, &pb.ServiceRequest{ServiceName: fmt.Sprintf("%s-in", serviceName), RoutingKey: fmt.Sprintf("%s-in", serviceName), QueueAutoDelete: false})
	defer conn.Close()

	// Define a WaitGroup
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		startConsumingWithRetry(c, fmt.Sprintf("%s-in", serviceName), 5, 5*time.Second)

		wg.Done() // Decrement the WaitGroup counter when the goroutine finishes
	}()

	wg.Wait()

}

func agreementsHandler(etcdRoot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := lib.GetRequestBody(w, r, serviceName)
		if err != nil {
			return
		}

		var requestApproval lib.RequestApproval
		err = json.Unmarshal(body, &requestApproval)
		if err != nil {
			logger.Sugar().Infof("%s: Error unmarshalling body into RequestApproval: %v", serviceName, err)
			http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
			return
		}

		err = checkRequestApproval(&requestApproval)
		if err != nil {
			logger.Sugar().Infof("%s: checkRequestApproval: %v", serviceName, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// In this function I want to simulate checking the policy Enforcer to see whether:
//   - I can have an agreement with each data steward
//   - Get a result channel or endpoint
//   - Return an access token
//   - Start a composition request
func checkRequestApproval(requestApproval *lib.RequestApproval) error {

	checkDataStewards(requestApproval)

	return nil
}

func checkDataStewards(requestApproval *lib.RequestApproval) {
	for _, steward := range requestApproval.DataProviders {
		output, err := lib.GetValueFromEtcd(etcdClient, "/policyEnforcer/agreements/"+steward)
		if err != nil {
			fmt.Println("do something")
		}

		if output == "" {
			fmt.Println("key not found")
		}

		var agreement lib.Agreement
		err = json.Unmarshal([]byte(output), &agreement)
		if err != nil {
			logger.Sugar().Errorw("%s: error unmarshalling agreement. %v", serviceName, err)
		}

		fmt.Println(agreement)

	}
}

func isUserInAgreement(agreement lib.Agreement, requestApproval lib.RequestApproval) bool {
	// This should be replaced by the appropriate value.
	// userIDKey := requestApproval.User.ID
	userName := requestApproval.User.UserName

	if relation, ok := agreement.Relations[userName]; ok {
		// Check if the ID matches
		if relation.ID == requestApproval.User.ID {
			return true
		}
	}

	// The user was not found in the relations map
	return false
}
