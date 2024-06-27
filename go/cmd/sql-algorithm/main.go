package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	"github.com/Jorrit05/DYNAMOS/pkg/msinit"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"github.com/gogo/protobuf/jsonpb"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	logger       = lib.InitLogger(logLevel)
	config       *msinit.Configuration
	COORDINATOR  = make(chan struct{})
	receiveMutex = &sync.Mutex{}
)

// Main function
func main() {
	logger.Debug("Starting algorithm service")

	oce, err := lib.InitTracer(serviceName)
	if err != nil {
		logger.Sugar().Fatalf("Failed to create ocagent-exporter: %v", err)
	}

	config, err = msinit.NewConfiguration(serviceName, grpcAddr, COORDINATOR, sideCarMessageHandler, sendDataHandler, receiveMutex)
	if err != nil {
		logger.Sugar().Fatalf("%v", err)
	}

	<-config.StopMicroservice
	config.SafeExit(oce, serviceName)
	os.Exit(0)
}

// // func ReveiceData() {
// // 	// Assuming `row` is your data fetched from the database.
// // 	fields := make(map[string]*structpb.Value)
// // 	fields["name"] = structpb.NewStringValue("Jorrit")
// // 	fields["date_of_birth"] = structpb.NewStringValue("september")
// // 	fields["job_title"] = structpb.NewStringValue("IT")
// // 	fields["other"] = structpb.NewBoolValue(true)

// // 	s := &structpb.Struct{Fields: fields}

// // 	iets := s.Fields["other"].GetListValue().ProtoReflect().Type()
// // 	fmt.Println(iets)
// // 	fmt.Println("xxx")
// // 	fmt.Println(s.GetFields())
// // }

// This is the function being called by the last microservice
func handleSqlDataRequest(ctx context.Context, msComm *pb.MicroserviceCommunication) error {
	ctx, span := trace.StartSpan(ctx, "handleSqlDataRequest")
	defer span.End()

	logger.Info("Start handleSqlDataRequest")
	// Unpack the metadata
	metadata := msComm.Metadata
	// fields := make(map[string]*structpb.Value)
	// dataField := msComm.GetData()
	// Get the "Functcat" field from the struct
	// functcatValue := dataField.Fields["HOOPgeb"]

	// // Check if it's a ListValue
	// if functcatValue != nil {
	// 	if listValue, ok := functcatValue.Kind.(*structpb.Value_ListValue); ok {
	// 		// Iterate over the Values in the ListValue
	// 		for _, item := range listValue.ListValue.GetValues() {
	// 			// item is a *structpb.Value, so we need to get the actual value using one of its getter methods
	// 			switch v := item.Kind.(type) {
	// 			case *structpb.Value_StringValue:
	// 				fmt.Printf("String value: %s\n", v.StringValue)
	// 			case *structpb.Value_NumberValue:
	// 				fmt.Printf("Number value: %f\n", v.NumberValue)
	// 			case *structpb.Value_BoolValue:
	// 				fmt.Printf("Bool value: %v\n", v.BoolValue)
	// 			// etc. for other possible types
	// 			default:
	// 				fmt.Printf("Other value: %v\n", v)
	// 			}
	// 		}
	// 	}
	// }

	// Print each metadata field
	logger.Sugar().Debugf("Length metadata: %s", strconv.Itoa(len(metadata)))
	// for key, value := range metadata {
	// 	fmt.Printf("Key: %s, Value: %+v\n", key, value)
	// }

	sqlDataRequest := &pb.SqlDataRequest{}
	if err := msComm.OriginalRequest.UnmarshalTo(sqlDataRequest); err != nil {
		logger.Sugar().Errorf("Failed to unmarshal sqlDataRequest message: %v", err)
	}

	msComm.Traces["binaryTrace"] = propagation.Binary(span.SpanContext())

	if sqlDataRequest.Options["graph"] {
		// jsonString, _ := json.Marshal(msComm.Data)
		// msComm.Result = jsonString

		m := &jsonpb.Marshaler{}
		jsonString, _ := m.MarshalToString(msComm.Data)
		msComm.Result = []byte(jsonString)

		return nil
	}

	if sqlDataRequest.Algorithm == "average" {
		// jsonString, _ := json.Marshal(msComm.Data)
		// msComm.Result = jsonString

		msComm.Result = getAverage(msComm.Data)
		return nil
	}

	// // Just pass on the data for now...
	// if config.LastService {
	// 	msComm.Result = getAverage(msComm.Data)
	// }

	// Process all data to make this service more realistic.
	ctx, allResults := convertAllData(ctx, msComm.Data)
	msComm.Result = allResults

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

	result := make(map[string]string)
	if maleCount != 0 {
		result["avg_salary_scale_men"] = fmt.Sprintf("%.3f", totalMaleSalary/float64(maleCount))
	}
	if femaleCount != 0 {
		result["avg_salary_scale_women"] = fmt.Sprintf("%.3f", totalFemaleSalary/float64(femaleCount))

	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		logger.Sugar().Error(err)
		return nil
	}

	return jsonResult
}
