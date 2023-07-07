package main

import (
	"fmt"
	"os"
	"time"
)

func getTarget() string {
	time.Sleep(5 * time.Second)
	return "RESULT"
}

func main() {

	// Create a channel that can hold a string value
	targetChan := make(chan string)

	if os.Getenv("DESIGNATED_GRPC_PsORT") != "" {
		fmt.Println("NOETOTET")
	}
	// Start a goroutine that gets the target and sends it to the channel
	go func() {
		target := getTarget() // assuming getTarget returns a string
		targetChan <- target
	}()

	fmt.Println("start wait")
	target := <-targetChan

	fmt.Println("Result: " + target)
}
