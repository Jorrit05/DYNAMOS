package lib

import (
	"encoding/json"
	"fmt"

	"os"
	"strings"

	"regexp"

	"github.com/google/uuid"
)

type Set map[string]struct{}

func NewSet(slice []string) Set {
	s := make(Set)
	for _, each := range slice {
		s[each] = struct{}{}
	}
	return s
}

func (s Set) Has(v string) bool {
	_, ok := s[v]
	return ok
}

// Function expects a valid email address and the length of the desired GUID
// Example:
//
// fmt.Println(GenerateJobName("example.two@cloud.com", 8))
// prints: example-two-12345678 (some GUID of length 8, prefixed with a hyphen)
func GenerateJobName(email string, length int) string {
	// Extract the part before '@' symbol
	atIndex := strings.Index(email, "@")
	if atIndex == -1 {
		return ""
	}
	domain := email[:atIndex]

	// Remove special characters and replace with hyphen
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	domain = re.ReplaceAllString(domain, "-")

	// Generate a random GUID of the given length
	guid := uuid.New().String()
	if len(guid) > length {
		guid = guid[:length]
	}

	// Construct the final email address
	return fmt.Sprintf("%s-%s", domain, guid)
}

func ReadFile(fileName string) (string, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return string(""), err
	}
	str := strings.TrimSuffix(string(data), "\n")

	return str, nil
}

func GetSQLConnectionString() (string, error) {
	user := os.Getenv("DB_USER")
	pwFile := os.Getenv("MYSQL_ROOT_PASSWORD_FILE")
	pw, err := ReadFile(pwFile)
	if err != nil {
		return "", err
	}

	pw = strings.TrimSuffix(pw, "\n")

	return fmt.Sprintf("%s:%s@tcp(mysql:3306)/%s", user, pw, os.Getenv("MYSQL_DATABASE")), nil
}

func GetDefaultRoutingKey(serviceName string) string {
	routingKey := os.Getenv("OUTPUT_QUEUE")
	if routingKey == "" {
		routingKey = "service." + strings.Split(serviceName, "_")[0]
	} else {
		routingKey = "service." + strings.Split(routingKey, "_")[0]
	}
	logger.Sugar().Infof("GetDefaultRoutingKey: %s", routingKey)
	return routingKey
}

// func ConvertRequest(req *http.Request) {
// 	body, err := ioutil.ReadAll(req.Body)
// 	if err != nil {
// 		logger.Sugar().Infof("Error reading body: %v", err)
// 		http.Error(w, "Error reading request body", http.StatusBadRequest)
// 		return
// 	}
// 	defer req.Body.Close()
// 	return body
// }

func LastPartAfterSlash(s string) string {
	splitted := strings.Split(s, "/")
	return splitted[len(splitted)-1]
}

// GenerateGuid returns a UUID string with the specified number of parts separated by dashes.
// If parts is 0 or greater than or equal to the total number of parts in the UUID,
// the full UUID string is returned.
func GenerateGuid(parts int) string {

	id := uuid.New()
	split := strings.Split(id.String(), "-")

	if parts == 0 || parts >= len(split)-1 {
		return id.String()
	}

	// Join the desired number of parts back together and return the resulting string
	return strings.Join(split[:parts], "-")
}

func SplitImageAndTag(fullImageName string) (string, string) {
	splitted := strings.Split(fullImageName, ":")
	if len(splitted) == 1 {
		return splitted[0], "latest"
	}
	return splitted[0], splitted[1]
}

// createMapFromSlice takes a slice of strings and returns a map
// where the keys are the elements of the slice and the values are set to true.
// This effectively creates a set representation of the slice,
// which will be used for efficient lookups.
func createMapFromSlice(slice []string) map[string]bool {
	m := make(map[string]bool)

	for _, element := range slice {
		m[element] = true
	}

	return m
}

// getNotMatchedElements takes a map and a slice of strings and returns a slice containing the elements that did not match (set difference).
// It iterates over the map keys and appends them into the notMatched slice.
func GetNotMatchedElements(mapA map[string]bool) []string {
	notMatched := []string{}

	for key := range mapA {
		notMatched = append(notMatched, key)
	}

	return notMatched
}

// getMatchedElements takes two slices of strings and returns a slice containing the matched elements (set intersection)
// and a map that contains the elements of sliceA that are not matched. It uses a map for efficient lookup of matching elements.
func GetMatchedElements(sliceA, sliceB []string) ([]string, map[string]bool) {
	mapA := createMapFromSlice(sliceA)
	matched := []string{}

	for _, b := range sliceB {
		if mapA[b] {
			matched = append(matched, b)
			delete(mapA, b)
		}
	}

	return matched, mapA
}

// // SliceIntersectAndDifference takes two slices of strings and returns two slices:
// // one containing the matched elements (set intersection), and the other containing
// // the elements that did not match (set difference). This function treats the input slices
// // as sets and removes duplicate elements from the output slices.
// //
// // Example 1:
// // sliceA := []string{"unl1_agent", "unl2_agent", "unl3"}
// // sliceB := []string{"unl2_agent", "unl5"}
// // matched, notMatched := SliceIntersectAndDifference(sliceA, sliceB)
// // Output: matched = [unl2_agent], notMatched = [unl1_agent unl3]
func SliceIntersectAndDifference(sliceA, sliceB []string) (matched []string, notMatched []string) {
	matched, mapA := GetMatchedElements(sliceA, sliceB)
	notMatched = GetNotMatchedElements(mapA)

	return matched, notMatched
}

func UnmarshalJsonFile[T any](fileLocation string, target *T) {
	jsonRep, err := os.ReadFile(fileLocation)
	if err != nil {
		logger.Sugar().Fatalw("Failed to read the config file: %v", err)
	}

	if err := json.Unmarshal(jsonRep, &target); err != nil {
		logger.Sugar().Fatalw("Failed to unmarshall the config: %v", err)
	}
}

// Returns a slice containing keys of the input map
func mapKeysToSlice(inputMap map[string]struct{}) []string {
	var outputSlice []string
	for key := range inputMap {
		outputSlice = append(outputSlice, key)
	}
	return outputSlice
}
