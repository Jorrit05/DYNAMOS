package main

import "errors"

// func chooseArchetype() {
// 	var allArchetypes [][]string
// 	dataproviderMap := make(map[string][]string)
// 	dataproviderMap["1"] = []string{"computeToData", "dataThroughTtp"}
// 	dataproviderMap["2"] = []string{"computeToData", "dataThroughTtp"}

// 	for _, dataProvider := range dataproviderMap {
// 		allArchetypes = append(allArchetypes, dataProvider)
// 	}

// 	var results [][]string
// 	for i, _ := range allArchetypes {
// 		results = append(results, intersect.HashGeneric(allArchetypes[i], allArchetypes[i+1]))

// 	}
// 	test := intersect.HashGeneric(allArchetypes[0], allArchetypes[1])

//		for k := range test {
//			fmt.Println(k)
//		}
//	}
func chooseArchetype() (string, error) {
	intersection := make(map[string]bool)

	first := true
	for _, dataProvider := range validationResponse.ValidDataproviders {
		if first {
			for _, archType := range dataProvider.ArcheTypes {
				intersection[archType] = true
			}
			first = false
		} else {
			newIntersection := make(map[string]bool)
			for _, archType := range dataProvider.ArcheTypes {
				if intersection[archType] {
					newIntersection[archType] = true
				}
			}
			intersection = newIntersection
		}
	}

	if len(intersection) == 0 {
		return "", errors.New("no common archetypes found")
	}

	// return the first common archetype
	for key := range intersection {
		return key, nil
	}

	return "", errors.New("unexpected error: could not retrieve an archetype from the intersection")
}

func main() {

	// Create a channel that can hold a string value
	chooseArchetype()

	// fmt.Println("Result: " + target)
}
