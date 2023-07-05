package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Jorrit05/DYNAMOS/pkg/api"
	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

type TypeField struct {
	Type string `json:"type"`
}

func sqlDataRequestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Entering sqlDataRequestHandler")

		body, err := api.GetRequestBody(w, r, serviceName)
		if err != nil {
			return
		}

		sqlDataRequest := &pb.SqlDataRequest{}

		err = protojson.Unmarshal(body, sqlDataRequest)
		if err != nil {
			logger.Sugar().Errorf("Error unmarshalling sqlDataRequest: %v", err)
			return
		}

		jobName, err := etcd.GetValueFromEtcd(etcdClient, fmt.Sprintf("/activeJobs/%s", sqlDataRequest.User.UserName))
		if err != nil {
			logger.Sugar().Errorf("Error getting target: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		sqlDataRequest.Target = jobName

		logger.Debug("Sending sqlDataReqeuest ")
		go c.SendSqlDataRequest(context.Background(), sqlDataRequest)

		//Handle response information
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("send SqlDataRequest to: %s", jobName)))

		// logger.Error("Unknown message type: " + typeField.Type)
		// http.Error(w, "Page not found", http.StatusNotFound)
	}
}
