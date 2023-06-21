package main

import (
	"encoding/json"
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	logger                      = lib.InitLogger()
	etcdClient *clientv3.Client = lib.GetEtcdClient(etcdEndpoints)
)

func main() {

	jsonStr := `{
		"type": "sqlDataRequest",
		"user": {
			"ID": "<GUID>",
			"userName": "jorrit.stutterheim@cloudnation.nl"
		},
		"dataProviders": ["VU","UVA","RUG"],
		"syncServices" : true
	}`

	var requestApproval lib.RequestApproval
	json.Unmarshal([]byte(jsonStr), &requestApproval)

	checkDataStewards(&requestApproval)

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
			logger.Sugar().Errorw("%s: error unmarshalling agreement. %v", serviceName, err)
		}

		fmt.Println(agreement)

	}
}
