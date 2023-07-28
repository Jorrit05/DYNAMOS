package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

var (
	logger = lib.InitLogger(logLevel)

	portMap = map[string]string{
		"surf": "30040",
		"uva":  "30030",
	}
)

func main() {

	rate := vegeta.Rate{Freq: 1, Per: time.Second}
	duration := 5 * time.Second

	metrics := &vegeta.Metrics{}
	attacker := vegeta.NewAttacker(vegeta.Timeout(60 * time.Second))

	res := &vegeta.Result{}
	var wg sync.WaitGroup

	// First attack
	for r := range attacker.Attack(customTargeter(res), rate, duration, "Get Request approval") {
		metrics.Add(r)
		if r.Code == http.StatusOK {
			res = r
			wg.Add(1)

			// Send data request
			go func(res *vegeta.Result) {
				// Decrement the counter when the goroutine completes.
				defer wg.Done()

				// Second attack
				for r := range attacker.Attack(dataRequestTargeter(res), rate, duration, "Data Request") {
					metrics.Add(r)
				}
			}(res)
		} else {
			// Optionally log the error here
			logger.Sugar().Infof("Error: %s, Code: %d, Body: %s\n", r.Error, r.Code, r.Body)
			continue
		}
	}

	// Wait for all attacks to finish
	wg.Wait()
	metrics.Close()
	fmt.Printf("99th percentile: %s\n", metrics.Latencies.P99)
	report := vegeta.NewTextReporter(metrics)
	report.Report(os.Stdout)

}

func createDataRequest(res *vegeta.Result) ([]byte, string) {
	acceptedRequest, err := getAcceptedDataRequest(res)
	if err != nil {
		logger.Sugar().Warnf("error: %v", err)
	}

	fmt.Println(acceptedRequest.JobId)
	dataRequest := getDataRequest(acceptedRequest)

	jsonData, err := json.Marshal(dataRequest)
	if err != nil {
		logger.Sugar().Fatalf("error marshalling sqldatarequest: %v", err)
	}

	target := ""
	for k := range acceptedRequest.AuthorizedProviders {
		target = strings.ToLower(k)
		break
	}
	endpoint := fmt.Sprintf("http://localhost:%s/agent/v1/sqlDataRequest/%s", portMap[target], target)
	return jsonData, endpoint
}

func customTargeter(res *vegeta.Result) vegeta.Targeter {

	requestApproval := "http://localhost:30010/api/v1/requestapproval"

	headers := http.Header{}
	headers.Add("Content-Type", "application/json")

	return vegeta.NewStaticTargeter(vegeta.Target{
		Method: "POST",
		URL:    requestApproval,
		Body:   getRequestApproval(),
		Header: headers,
	})
}

func dataRequestTargeter(res *vegeta.Result) vegeta.Targeter {

	headers := http.Header{}
	headers.Add("Content-Type", "application/json")
	headers.Add("Authorization", "bearer 1234")

	jsonData, endpoint := createDataRequest(res)

	return vegeta.NewStaticTargeter(vegeta.Target{
		Method: "POST",
		URL:    endpoint,
		Body:   jsonData,
		Header: headers,
	})
}
