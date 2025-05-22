package main

import (
	"bytes"
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
	logger        = lib.InitLogger(logLevel)
	counter       = 0
	archetypeMap  = make(map[int]string)
	customMetrics = &CustomMetrics{Requests: 0, Successes: 0, Failures: 0}
)

type CustomMetrics struct {
	Requests  int // Total number of requests
	Successes int // Total number of successful responses
	Failures  int // Total number of failed responses
	// Add more fields as needed...
}

func sendDataRequest(res *vegeta.Result) error {

	jsonData, endpoint := createDataRequest(res)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
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
		customMetrics.Failures++
		logger.Sugar().Infof("sendDataRequest, Code: %d, Body: %s\n", resp.StatusCode, resp.Body)

	} else {
		customMetrics.Successes++
	}

	return nil
}

func doDataRequest(attacker *vegeta.Attacker, rate vegeta.ConstantPacer, duration time.Duration, metrics *vegeta.Metrics, wg *sync.WaitGroup) {
	res := &vegeta.Result{}

	// First attack
	for r := range attacker.Attack(requesApprovalTargeter(res), rate, duration, "Get Request approval") {
		metrics.Add(r)
		if r.Code == http.StatusOK {
			res = r

			wg.Add(1)
			go func() {
				// Second call
				// nrOfCalls := 1 //getRandomInt(2)
				// customMetrics.Requests += nrOfCalls
				// for i := 1; i <= nrOfCalls; i++ {
				time.Sleep(1 * time.Second)
				err := sendDataRequest(res)
				if err != nil {
					logger.Sugar().Warn("error: %v", err)
				}
				// time.Sleep(time.Duration(getRandomInt(2)) * time.Second)
				// }
				wg.Done()
			}()
			time.Sleep(2 * time.Second)

			// for r := range attacker.Attack(dataRequestTargeter(res), rate, duration, "Data Request") {
			// 	if r.Code != http.StatusOK {
			// 		logger.Sugar().Infof("Second attack! Error: %s, Code: %d, Body: %s\n", r.Error, r.Code, r.Body)

			// 	}
			// 	metrics.Add(r)
			// }
		} else {
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
	// go progressBar()
	archetypeMap[1] = "computeToData"
	archetypeMap[2] = "dataThroughTtp"

	metrics := &vegeta.Metrics{}
	var wg sync.WaitGroup
	start := time.Now()

	err := updateArchetype(archetypeMap[2])
	if err != nil {
		logger.Sugar().Warnf("update archetype err: %v", err)
	}
	for i := 1; i <= 3; i++ {

		logger.Sugar().Debugf("Series: %d", i)
		// Series 1
		attacker, rate, duration := getAttackerAttributes(1, 4)

		doDataRequest(attacker, rate, duration, metrics, &wg)

		time.Sleep(3 * time.Second)
		// // logger.Debug("start 2 second sleep")

		// // Series 2
		// // attacker, rate, duration = getAttackerAttributes(2, 3)

		// // doDataRequest(attacker, rate, duration, metrics, &wg)

		// // // Series 3
		// err := updateArchetype(archetypeMap[getRandomInt(2)])
		// if err != nil {
		// 	logger.Sugar().Warnf("update archetype err: %v", err)
		// }
		// time.Sleep(5 * time.Second)

		// err := updateArchetype(archetypeMap[2])
		// if err != nil {
		// 	logger.Sugar().Warnf("update archetype err: %v", err)
		// }
		// time.Sleep(5 * time.Second)
		// // logger.Debug("start 4 second sleep")
		// // time.Sleep(time.Duration(getRandomInt(3)) * time.Second)
		// attacker, rate, duration = getAttackerAttributes(1, 5)

		// doDataRequest(attacker, rate, duration, metrics, &wg)
		// time.Sleep(time.Duration(getRandomInt(3)) * time.Second)
	}
	// Wait for all attacks to finish

	wg.Wait()
	metrics.Close()
	fmt.Printf("99th percentile: %s\n", metrics.Latencies.P99)
	report := vegeta.NewTextReporter(metrics)
	report.Report(os.Stdout)

	logger.Info("Custom metrics:")
	logger.Sugar().Infof("Request: %d", customMetrics.Requests)
	logger.Sugar().Infof("Successes: %d", customMetrics.Successes)
	logger.Sugar().Infof("Failures: %d", customMetrics.Failures)
	elapsed := time.Since(start)

	fmt.Printf("The test took %s seconds to execute.\n", elapsed)

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
	// logger.Info(string(jsonData))
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
	return func(t *vegeta.Target) error {
		t.Method = "POST"
		t.URL = endpoint
		t.Body = jsonData
		t.Header = headers
		return nil
	}
}
