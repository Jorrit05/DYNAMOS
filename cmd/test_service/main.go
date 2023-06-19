package main

import (
	"fmt"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
)

var logger = lib.InitLogger()

func printsth() {
	for i := 1; i <= 1000000000; i++ { // maximum of 7 retries

		logger.Info("SuperCool message this is.")
		time.Sleep(10 * time.Second) // wait for 10 seconds before retrying
		if i == 3 {
			err := fmt.Errorf("this is my eerror")
			logger.Sugar().Fatalw("did not connect to grpc server: %v", err)
		}
	}

}

func main() {

	defer logger.Sync() // flushes buffer, if any

	logger.Info("This is an INFO message")
	logger.Warn("This is a WARN message")
	logger.Error("This is an ERROR message")
	go printsth()

	select {}
	//logger.Fatal("This is a FATAL message") // Note: .Fatal() will cause the program to exit.
}

// func main() {

// 	fmt.Println("This is an info message using fmt.Println")
// 	logger.Info("This is an info message using log.Println")

// 	logger.Sugar().Infow("This is a formatted info message using log.Printf: %d", 42)

// 	// fmt.
// 	go printsth()
// 	//	logger.Sugar().Fatalw("This is a fatal error message using log.Fatalf: %s", "something went wrong")
// 	select {}
// }
