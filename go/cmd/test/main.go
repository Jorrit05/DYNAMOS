package main

import (
	"flag"
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	"go.uber.org/zap"
)

var (
	logLevel = zap.LevelFlag("level", zap.InfoLevel, "log level")

	logger = lib.InitLogger(*logLevel)

	port          = flag.Int("port", 3005, "The server port")
	etcdEndpoints = "http://localhost:30005"
	serviceName   = "test"
)

func main() {

	flag.Parse()
	logger.Debug("Debug")
	logger.Info(fmt.Sprintf("Port: %v", *port))

}
