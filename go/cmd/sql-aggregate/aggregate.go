// To aggregate two structpb.Struct instances by merging all identical columns, you would first need to iterate over the fields of both structs, compare them, and merge the identical ones. Here’s a simplified approach to how you might do it:

// Iterate through the fields of both data structures.
// For each field, check if the field exists in both structs.
// If a field exists in both, merge the values. This step depends on the nature of the data. For simplicity, let's assume we're just appending the values if they are lists or replacing them if they are singular values.
// If a field exists in only one of the structs, just carry it over to the resulting struct.
// This approach assumes that the data in identical columns can be meaningfully merged. For scalar fields (e.g., strings, numbers), you may need to decide whether to take one value over the other or to merge them based on specific logic.

// Here’s an example function that performs the merge:

package main

import (
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

func mergeData(msCommList []*pb.MicroserviceCommunication) *pb.MicroserviceCommunication {
	mergedData := &structpb.Struct{
		Fields: make(map[string]*structpb.Value),
	}

	// Merge data1 into mergedData
	for key, value := range msCommList[0].Data.GetFields() {
		mergedData.Fields[key] = value
	}
	for _, msComm := range mscommList {

		// Merge rest of the data into mergedData, checking for and handling identical fields
		for key, value2 := range msComm.Data.GetFields() {

			if value1, exists := mergedData.Fields[key]; exists {
				// Handle merging identical fields. This example simply appends the lists.
				// Custom logic needed based on the actual data structure and requirements.
				if value1.GetListValue() != nil && value2.GetListValue() != nil {
					mergedList := append(value1.GetListValue().Values, value2.GetListValue().Values...)
					mergedData.Fields[key] = structpb.NewListValue(&structpb.ListValue{Values: mergedList})
				} else {
					// For non-list values, decide how to merge. This example simply replaces the value.
					mergedData.Fields[key] = value2
				}
			} else {
				// If the field is not in data1, add it from data2.
				mergedData.Fields[key] = value2
			}
		}
	}

	// For now simply return the first mscommList with the merged data.
	msCommList[0].Data = mergedData
	return msCommList[0]
}
