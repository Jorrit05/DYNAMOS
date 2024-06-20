// Certainly! Let's enhance the mergeData function to handle different data structures more gracefully. Specifically, I'll demonstrate how to handle the following cases when identical fields are encountered:

// Both fields are lists: Merge by appending the second list to the first.
// One field is a list, and the other is a singular value: Append the singular value to the list.
// Both fields are singular values but of different types (e.g., one is a string and the other is a number): Convert them to strings and concatenate with a delimiter.
// Both fields are singular values of the same type: For simplicity, replace the first with the second, but a specific merge logic could be applied based on the type.
// Here's the enhanced version:
package main

// func mergeData2(data1, data2 *structpb.Struct) *structpb.Struct {
// 	mergedData := &structpb.Struct{
// 		Fields: make(map[string]*structpb.Value),
// 	}

// 	// First, merge data1 into mergedData.
// 	for key, value := range data1.GetFields() {
// 		mergedData.Fields[key] = value
// 	}

// 	// Then, merge data2 into mergedData, handling identical fields carefully.
// 	for key, value2 := range data2.GetFields() {
// 		if value1, exists := mergedData.Fields[key]; exists {
// 			// Check if both are list values.
// 			if listValue1, ok1 := value1.Kind.(*structpb.Value_ListValue); ok1 {
// 				if listValue2, ok2 := value2.Kind.(*structpb.Value_ListValue); ok2 {
// 					// Both fields are lists: Merge by appending.
// 					mergedList := append(listValue1.ListValue.Values, listValue2.ListValue.Values...)
// 					mergedData.Fields[key] = structpb.NewListValue(&structpb.ListValue{Values: mergedList})
// 				} else {
// 					// Field 1 is a list, field 2 is a singular value: Append field 2 to the list.
// 					mergedData.Fields[key].GetListValue().Values = append(mergedData.Fields[key].GetListValue().Values, value2)
// 				}
// 			} else if _, ok2 := value2.Kind.(*structpb.Value_ListValue); ok2 {
// 				// Field 1 is a singular value, field 2 is a list: Prepend field 1 to the list.
// 				mergedData.Fields[key] = structpb.NewListValue(&structpb.ListValue{Values: append([]*structpb.Value{value1}, value2.GetListValue().Values...)})
// 			} else {
// 				// Both fields are singular values: Merge based on specific logic.
// 				// For demonstration, convert to strings and concatenate if different types.
// 				if reflect.TypeOf(value1.Kind) != reflect.TypeOf(value2.Kind) {
// 					mergedValue := fmt.Sprintf("%v|%v", value1.AsString(), value2.AsString())
// 					mergedData.Fields[key] = structpb.NewStringValue(mergedValue)
// 				} else {
// 					// If of the same type, replace value1 with value2 (could apply other logic).
// 					mergedData.Fields[key] = value2
// 				}
// 			}
// 		} else {
// 			// If the field exists only in data2, add it to mergedData.
// 			mergedData.Fields[key] = value2
// 		}
// 	}

// 	return mergedData
// }

// func main() {
// 	// Example usage with various data structures
// 	data1 := &structpb.Struct{
// 		Fields: map[string]*structpb.Value{
// 			"ListAndValue": structpb.NewListValue(&structpb.ListValue{Values: []*structpb.Value{
// 				structpb.NewStringValue("1"),
// 				structpb.NewStringValue("2"),
// 			}}),
// 			"DifferentTypes": structpb.NewNumberValue(1),
// 		},
// 	}

// 	data2 := &structpb.Struct{
// 		Fields: map[string]*structpb.Value{
// 			"ListAndValue":   structpb.NewStringValue("3"),
// 			"DifferentTypes": structpb.NewStringValue("Two"),
// 		},
// 	}

// 	mergedData := mergeData(data1, data2)
// 	fmt.Println(mergedData)
// }
