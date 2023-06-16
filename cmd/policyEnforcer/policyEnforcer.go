package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Jorrit05/micro-recomposer/pkg/lib"
	"github.com/gorilla/handlers"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	log, logFile                  = lib.InitLogger(logFileLocation, serviceName)
	etcdClient   *clientv3.Client = lib.GetEtcdClient(etcdEndpoints)
)

func main() {

	defer logFile.Close()
	defer etcdClient.Close()
	defer lib.HandlePanicAndFlushLogs(log, logFile)

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	mux := http.NewServeMux()
	mux.HandleFunc("/validate/", agreementsHandler("/policyEnforcer"))

	log.Info(fmt.Sprintf("Starting http server on %s/30012", port))
	go func() {
		if err := http.ListenAndServe(port, handlers.CORS(originsOk, headersOk, methodsOk)(mux)); err != nil {
			log.Fatalf("Error starting HTTP server: %s", err)
		}
	}()

	select {}

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
			log.Printf("%s: Error unmarshalling body into RequestApproval: %v", serviceName, err)
			http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
			return
		}

		err = checkRequestApproval(&requestApproval)
		if err != nil {
			log.Printf("%s: checkRequestApproval: %v", serviceName, err)
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
			fmt.Println("do somthing")
		}

		if output == "" {
			fmt.Println("key not found")
		}

		var agreement lib.Agreement
		err = json.Unmarshal([]byte(output), &agreement)
		if err != nil {
			log.Errorf("%s: error unmarshalling agreement. %v", serviceName, err)
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
