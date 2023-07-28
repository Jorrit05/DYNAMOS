package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type SQLDataRequest struct {
	Type             string            `json:"type"`
	Query            string            `json:"query"`
	Graph            bool              `json:"graph"`
	Algorithm        string            `json:"algorithm"`
	AlgorithmColumns map[string]string `json:"algorithmColumns"`
	User             User              `json:"user"`
	RequestMetadata  Metadata          `json:"requestMetadata"`
}

type User struct {
	ID       string `json:"id"`
	UserName string `json:"userName"`
}

type Metadata struct {
	JobID string `json:"jobId"`
}

func getRequestApproval() []byte {
	request := `{
		"type": "sqlDataRequest",
		"user": {
			"ID": "1234",
			"userName": "jorrit.stutterheim@cloudnation.nl"
		},
		"dataProviders": ["VU","UVA","RUG"],
		"syncServices": true
	}`

	return []byte(request)
}

func getAcceptedDataRequest(res *vegeta.Result) (*pb.AcceptedDataRequest, error) {

	var response *pb.AcceptedDataRequest
	if err := json.Unmarshal(res.Body, &response); err != nil {
		// prettyPrint(res)
		return nil, err
	}
	return response, nil
}

func generateSqlLimit() int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	num := r.Intn(100) + 1 // Generate a random number between 1 and 20000.
	fmt.Println(num)
	return num
}

func getDataRequest(acceptedDataRequest *pb.AcceptedDataRequest) *SQLDataRequest {

	return &SQLDataRequest{
		Type:      "sqlDataRequest",
		Query:     "SELECT * FROM Personen p JOIN Aanstellingen s LIMIT " + strconv.Itoa(generateSqlLimit()),
		Graph:     true,
		Algorithm: "average",
		AlgorithmColumns: map[string]string{
			"Geslacht": "Aanst_22, Gebdat",
		},
		User: User{
			ID:       "1234",
			UserName: acceptedDataRequest.User.UserName,
		},
		RequestMetadata: Metadata{
			JobID: acceptedDataRequest.JobId,
		},
	}
}

func prettyPrint(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		logger.Sugar().Fatalf("Error when trying to pretty-print: %v", err)
	}
	fmt.Println(string(b))
}
