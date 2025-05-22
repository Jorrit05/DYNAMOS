package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	progress "github.com/cheggaaa/pb/v3"

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

type RequestApproval struct {
	Type             string   `json:"type"`
	DataProviders    []string `json:"dataProviders"`
	Graph            bool     `json:"graph"`
	SyncServices     bool     `json:"syncServices"`
	User             User     `json:"user"`
	DestinationQueue string   `json:"destinationQueue"`
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
			"ID": "",
			"userName": "jorrit.stutterheim@cloudnation.nl"
		},
		"dataProviders": ["VU","UVA","RUG"],
		"syncServices": true
	}`

	requestApproval := &RequestApproval{}
	err := json.Unmarshal([]byte(request), requestApproval)
	if err != nil {
		logger.Sugar().Fatalf("Error json.Unmarsha: %v", err)

	}

	counter++
	requestApproval.User.ID = fmt.Sprintf("1234%s", strconv.Itoa(counter))

	json, err := json.Marshal(requestApproval)
	if err != nil {
		logger.Sugar().Fatalf("Error json.Marshal: %v", err)

	}
	// logger.Info(string(json))
	return json
}

func getAcceptedDataRequest(res *vegeta.Result) (*pb.RequestApprovalResponse, error) {

	var response *pb.RequestApprovalResponse
	if err := json.Unmarshal(res.Body, &response); err != nil {
		// prettyPrint(res)
		return nil, err
	}
	return response, nil
}

func getRandomInt(maxSize int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	num := r.Intn(maxSize) + 1 // Generate a random number between 1 and 20000.
	// fmt.Println(num)
	return num
}

func getDataRequest(requestApprovalResponse *pb.RequestApprovalResponse) *SQLDataRequest {

	return &SQLDataRequest{
		Type:      "sqlDataRequest",
		Query:     "SELECT * FROM Personen p JOIN Aanstellingen s LIMIT " + strconv.Itoa(getRandomInt(100)),
		Graph:     true,
		Algorithm: "average",
		AlgorithmColumns: map[string]string{
			"Geslacht": "Aanst_22, Gebdat",
		},
		User: User{
			ID:       requestApprovalResponse.User.Id,
			UserName: requestApprovalResponse.User.UserName,
		},
		RequestMetadata: Metadata{
			JobID: requestApprovalResponse.JobId,
		},
	}
}

func progressBar() {
	// Create a new progress bar with a total of 10.
	bar := progress.StartNew(10)

	for i := 0; i < 10; i++ {
		// Sleep for 1 second to simulate work.
		time.Sleep(time.Second)

		// Increment the progress bar.
		bar.Increment()
	}

	// Finish the progress bar.
	bar.Finish()
}
func prettyPrint(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		logger.Sugar().Fatalf("Error when trying to pretty-print: %v", err)
	}
	fmt.Println(string(b))
}

func updateArchetype(allowedArchetypes string) error {
	url := "http://orchestrator.orchestrator.svc.cluster.local:80/api/v1/policyEnforcer/agreements"

	jsonData := []byte(fmt.Sprintf(`{
		"name": "UVA",
		"relations": {
			"jorrit.stutterheim@cloudnation.nl" : {
				"ID" : "GUID",
				"requestTypes" : ["sqlDataRequest"],
				"dataSets" : ["wageGap"],
				"allowedArchetypes" : ["%s"],
				"allowedComputeProviders" : ["SURF"]
			}
		},
		"computeProviders" : ["SURF", "otherCompany"],
		"archetypes" : ["computeToData", "dataThroughTtp",  "reproducableScience"]
	}`, allowedArchetypes))

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "bearer 1234")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if http.StatusOK != resp.StatusCode {
		logger.Warn(strconv.Itoa(resp.StatusCode))
	}
	return nil
}
