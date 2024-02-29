package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	"github.com/Jorrit05/DYNAMOS/pkg/msinit"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	logger      = lib.InitLogger(logLevel)
	config      *msinit.Configuration
	COORDINATOR = make(chan struct{})
)

// Main function
func main() {
	logger.Sugar().Debugf("Starting %s", serviceName)

	oce, err := lib.InitTracer(serviceName)
	if err != nil {
		logger.Sugar().Fatalf("Failed to create ocagent-exporter: %v", err)
	}

	config, err = msinit.NewConfiguration(serviceName, grpcAddr, COORDINATOR, sideCarMessageHandler, sendDataHandler)
	if err != nil {
		logger.Sugar().Fatalf("%v", err)
	}

	<-config.Stopped
	logger.Sugar().Infof("Wait 2 seconds before ending %s", serviceName)

	oce.Flush()
	time.Sleep(2 * time.Second)
	oce.Stop()
	config.CloseConnection()
	logger.Sugar().Infof("Exiting %s", serviceName)

	os.Exit(0)
}

// This is the function being called by the last microservice
func handleSqlDataRequest(ctx context.Context, msComm *pb.MicroserviceCommunication) error {
	ctx, span := trace.StartSpan(ctx, "handleSqlDataRequest")
	defer span.End()

	logger.Sugar().Infof("Start %s handleSqlDataRequest", serviceName)

	sqlDataRequest := &pb.SqlDataRequest{}
	if err := msComm.OriginalRequest.UnmarshalTo(sqlDataRequest); err != nil {
		logger.Sugar().Errorf("Failed to unmarshal sqlDataRequest message: %v", err)
	}

	// Coordinator ensures all services are started before further processing messages
	<-COORDINATOR
	msComm.Traces["binaryTrace"] = propagation.Binary(span.SpanContext())

	c := pb.NewMicroserviceClient(config.GrpcConnection)

	// Process all data to make this service more realistic.
	ctx, allResults := convertAllData(ctx, msComm.Data)
	msComm.Result = allResults
	// err := os.WriteFile("text.txt", allResults, 0644)
	// if err != nil {
	// 	return err
	// }

	c.SendData(ctx, msComm)

	close(config.StopServer)
	return nil
}

func convertAllData(ctx context.Context, data *structpb.Struct) (context.Context, []byte) {
	ctx, span := trace.StartSpan(ctx, "convertAllData")
	defer span.End()
	keys := make([]string, 0)
	allValues := make([][]string, 0)
	maxLength := 0

	for key, value := range data.GetFields() {
		stringValues := value.GetListValue().GetValues()
		if len(stringValues) > 0 {
			keys = append(keys, key)
			rowValues := make([]string, len(stringValues))
			for i, v := range stringValues {
				rowValues[i] = v.GetStringValue()
			}
			allValues = append(allValues, rowValues)
			if len(rowValues) > maxLength {
				maxLength = len(rowValues)
			}
		}
	}

	result := make([][]string, maxLength+1)
	result[0] = keys
	for i := 1; i < maxLength+1; i++ {
		row := make([]string, len(keys))
		for j := 0; j < len(keys); j++ {
			if i <= len(allValues[j]) {
				row[j] = allValues[j][i-1]
			} else {
				row[j] = ""
			}
		}
		result[i] = row
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		fmt.Printf("Error while marshalling to JSON: %v\n", err)
		return ctx, nil
	}

	return ctx, jsonData
}

func getFirstRow(data *structpb.Struct) []byte {
	keys := make([]string, 0)
	values := make([]string, 0)
	for key, value := range data.GetFields() {
		stringValues := value.GetListValue().GetValues()
		if len(stringValues) > 0 {
			keys = append(keys, key)
			values = append(values, stringValues[0].GetStringValue())
		}
	}

	// Convert to JSON format
	result := []interface{}{keys, values}
	jsonData, err := json.Marshal(result)
	if err != nil {
		fmt.Printf("Error while marshalling to JSON: %v\n", err)
		return nil
	}

	return jsonData
}

func getAverage(data *structpb.Struct) []byte {

	gendersField, ok1 := data.GetFields()["Geslacht"]
	salariesField, ok2 := data.GetFields()["Salschal"]

	if !ok1 || !ok2 {
		logger.Error("Genders or Salaries field not found")
		return nil
	}

	genders := gendersField.GetListValue().GetValues()
	salaries := salariesField.GetListValue().GetValues()

	var totalMaleSalary, totalFemaleSalary float64
	maleCount, femaleCount := 0, 0

	for index, gender := range genders {
		genderStr := gender.GetStringValue()
		if salaryStr := salaries[index].GetStringValue(); salaryStr != "" {
			salary, err := strconv.ParseFloat(salaryStr, 64)
			if err != nil {
				fmt.Printf("Error parsing salary value: %v\n", err)
				continue
			}

			if genderStr == "M" {
				totalMaleSalary += salary
				maleCount++
			} else if genderStr == "V" {
				totalFemaleSalary += salary
				femaleCount++
			}
		}
	}

	result := make(map[string]float64)
	if maleCount != 0 {
		result["average_male_salary"] = totalMaleSalary / float64(maleCount)
	}
	if femaleCount != 0 {
		result["average_female_salary"] = totalFemaleSalary / float64(femaleCount)
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		logger.Sugar().Error(err)
		return nil
	}

	return jsonResult
}