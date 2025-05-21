package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"google.golang.org/protobuf/types/known/structpb"
)

func loadCSV(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Function to calculate statistics and return as a map
func differentialPrivacy(values []string) map[string]string {

	logger.Info("Differential Privacy functionality")

	result := map[string]string{
		"Data":       "Anonymized",
		"emptyRatio": fmt.Sprintf("%v", 0.0),
	}

	// // Return early if there are no valid numbers
	// if len(nums) == 0 {
	// 	return result
	// }

	// // Sort numbers to calculate min, max, and median
	// sort.Float64s(nums)

	// //emptyRatio := emptyCount / len(nums)
	// emptyRatio := float64(emptyCount)/float64(len(values))

	// // Calculate min and max
	// min := nums[0]
	// max := nums[len(nums)-1]

	// // Calculate mean
	// sum := 0.0
	// for _, num := range nums {
	// 	sum += num
	// }
	// mean := sum / float64(len(nums))

	// // Calculate median
	// var median float64
	// n := len(nums)
	// if n%2 == 0 {
	// 	median = (nums[n/2-1] + nums[n/2]) / 2
	// } else {
	// 	median = nums[n/2]
	// }

	// // Calculate standard deviation
	// sumOfSquares := 0.0
	// for _, num := range nums {
	// 	sumOfSquares += math.Pow(num-mean, 2)
	// }
	// stdDev := math.Sqrt(sumOfSquares / float64(n))

	// // Update the result map
	// result["emptyRatio"] = emptyRatio
	// result["min"] = min
	// result["max"] = max
	// result["mean"] = mean
	// result["median"] = median
	// result["stdDev"] = stdDev

	return result
}

// This is the function being called by the last microservice
func handleDataRequest(ctx context.Context, msComm *pb.MicroserviceCommunication) error {
	ctx, span := trace.StartSpan(ctx, "handleDataRequest")
	defer span.End()

	logger.Info("Start handleDataRequest")
	// Unpack the metadata
	metadata := msComm.Metadata
	filename := "/res/synthetic_data_sample.csv"

	// Print each metadata field
	logger.Sugar().Debugf("Metadata: %v", metadata)
	logger.Sugar().Debugf("Length metadata: %s", strconv.Itoa(len(metadata)))

	sqlDataRequest := &pb.SqlDataRequest{}
	if err := msComm.OriginalRequest.UnmarshalTo(sqlDataRequest); err != nil {
		logger.Sugar().Errorf("Failed to unmarshal sqlDataRequest message: %v", err)
	}

	msComm.Traces["binaryTrace"] = propagation.Binary(span.SpanContext())

	//load data
	logger.Sugar().Debugf("Loading data from file: %v", filename)
	records, err := loadCSV(filename)
	if err != nil {
		log.Fatal(err)
	}
	logger.Sugar().Debugf("Type of records: %v", reflect.TypeOf(records))
	logger.Sugar().Debugf("Length of records: %v", len(records))

	cols := records[0]
	buildYearsCol := []string{}
	bedroomCol := []string{}
	logger.Sugar().Debugf("Columns: %v", cols)
	for _, record := range records[1:] {
		//fmt.Println(record)
		buildYearsCol = append(buildYearsCol, record[0])
		bedroomCol = append(bedroomCol, record[1])
	}

	result := make(map[string]string)
	// statsBuildYear := calculateStats(buildYearsCol)
	// statsBedroom := calculateStats(bedroomCol)
	result = differentialPrivacy(buildYearsCol)

	logger.Sugar().Debugf("Request Options: %v", sqlDataRequest.Options)
	// if sqlDataRequest.Options["buildYear"] {
	// 	//result["buildYear"] = fmt.Sprintf("%.3f", statsBuildYear["mean"])
	// 	jsonMetrics, err := json.Marshal(statsBuildYear)
	// 	if err != nil {
	// 		fmt.Println("Error:", err)
	// 		return nil
	// 	}
	// 	result["buildYear"] = string(jsonMetrics)
	// }
	// if sqlDataRequest.Options["bedroomWindows"] {
	// 	jsonMetrics, err := json.Marshal(statsBedroom)
	// 	if err != nil {
	// 		fmt.Println("Error:", err)
	// 		return nil
	// 	}
	// 	//result["shower"] = fmt.Sprintf("%.3f", statsShower["mean"])
	// 	result["bedroomWindows"] = string(jsonMetrics)
	// }

	jsonResult, err := json.Marshal(result)
	if err != nil {
		logger.Sugar().Error(err)
		return nil
	}

	// Process all data to make this service more realistic.
	msComm.Result = jsonResult
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
