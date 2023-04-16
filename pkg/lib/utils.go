package lib

import (
	"fmt"

	"os"
	"strings"

	"github.com/google/uuid"
)

func ReadFile(fileName string) (string, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return string(""), err
	}
	str := strings.TrimSuffix(string(data), "\n")

	return str, nil
}

func GetAMQConnectionString() (string, error) {
	user := os.Getenv("AMQ_USER")
	pwFile := os.Getenv("AMQ_PASSWORD_FILE")
	pw, err := ReadFile(pwFile)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("amqp://%s:%s@rabbit:5672/", user, pw), nil
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
	log.Printf("GetDefaultRoutingKey: %s", routingKey)
	return routingKey
}

// func ConvertRequest(req *http.Request) {
// 	body, err := ioutil.ReadAll(req.Body)
// 	if err != nil {
// 		log.Printf("Error reading body: %v", err)
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

func SliceDifferenceString(sliceA, sliceB []string) []string {
	diff := make([]string, 0)
	elementsInB := make(map[string]bool)

	for _, item := range sliceB {
		elementsInB[item] = true
	}

	for _, item := range sliceA {
		if !elementsInB[item] {
			diff = append(diff, item)
		}
	}

	return diff
}
