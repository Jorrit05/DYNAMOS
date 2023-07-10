package main

import (
	"fmt"
	"time"
)

func getTarget(target chan string) string {
	time.Sleep(3 * time.Second)
	target <- "TEST"
	return "RESULT"
}
func func2(target chan string) string {
	x := <-target
	fmt.Println(x)
	return "3"
}

func main() {

	// Create a channel that can hold a string value
	targetChan := make(chan string)

	// Start a goroutine that gets the target and sends it to the channel
	go getTarget(targetChan) // assuming getTarget returns a string

	func2(targetChan)
	// target := <-targetChan

	// fmt.Println("Result: " + target)
}
