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
	logger  = lib.InitLogger(logLevel)
	counter = 0
)

func doDataRequest(attacker *vegeta.Attacker, rate vegeta.ConstantPacer, duration time.Duration, metrics *vegeta.Metrics, wg *sync.WaitGroup) {
	res := &vegeta.Result{}

	// First attack
	for r := range attacker.Attack(requesApprovalTargeter(res), rate, duration, "Get Request approval") {
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

					if r.Code != http.StatusOK {
						logger.Sugar().Infof("Second attack! Error: %s, Code: %d, Body: %s\n", r.Error, r.Code, r.Body)

					}
					metrics.Add(r)
				}
			}(res)
		} else {
			// Optionally log the error here
			logger.Sugar().Infof("Error: %s, Code: %d, Body: %s\n", r.Error, r.Code, r.Body)
			continue
		}
	}

}

func getAttackerAttributes(frequency int, length int) (*vegeta.Attacker, vegeta.ConstantPacer, time.Duration) {

	rate := vegeta.Rate{Freq: frequency, Per: time.Second}
	duration := time.Duration(length) * time.Second

	attacker := vegeta.NewAttacker(vegeta.Timeout(90 * time.Second))
	return attacker, rate, duration
}

func main() {

	metrics := &vegeta.Metrics{}
	attacker, rate, duration := getAttackerAttributes(2, 2)
	var wg sync.WaitGroup

	// Series 1
	doDataRequest(attacker, rate, duration, metrics, &wg)

	time.Sleep(2 * time.Second)
	attacker, rate, duration = getAttackerAttributes(1, 5)

	doDataRequest(attacker, rate, duration, metrics, &wg)

	err := updateArchetype("computeToData")
	if err != nil {
		logger.Sugar().Warnf("update archetype err: %v", err)
	}

	time.Sleep(3 * time.Second)
	attacker, rate, duration = getAttackerAttributes(3, 3)

	doDataRequest(attacker, rate, duration, metrics, &wg)

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

	// fmt.Println(acceptedRequest.JobId)
	dataRequest := getDataRequest(acceptedRequest)

	jsonData, err := json.Marshal(dataRequest)
	if err != nil {
		logger.Sugar().Fatalf("error marshalling sqldatarequest: %v", err)
	}

	target := ""
	url := ""
	for k, v := range acceptedRequest.AuthorizedProviders {
		target = strings.ToLower(k)
		url = v
		break
	}
	endpoint := fmt.Sprintf("http://%s:80/agent/v1/sqlDataRequest/%s", url, target)
	return jsonData, endpoint
}

func requesApprovalTargeter(res *vegeta.Result) vegeta.Targeter {

	requestApproval := "http://orchestrator.orchestrator.svc.cluster.local:80/api/v1/requestapproval"

	headers := http.Header{}
	headers.Add("Content-Type", "application/json")

	return func(t *vegeta.Target) error {
		t.Method = "POST"
		t.URL = requestApproval
		t.Body = getRequestApproval()
		t.Header = headers
		return nil
	}
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
