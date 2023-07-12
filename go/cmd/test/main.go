package main

import (
	"os"
	"strconv"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	logger      = lib.InitLogger(zap.DebugLevel)
	conn        *grpc.ClientConn
	serviceName string = "test"
)

type MsCommunication struct {
	SecureChannel   // This is embedding. Go's version of inheritance.
	NextServicePort string
	Client          MicroserviceStub
}

func NewMsCommunication(grpcAddr string) *MsCommunication {
	nextServicePort := ""
	if last, _ := strconv.Atoi(os.Getenv("LAST")); last > 0 {
		nextServicePort = os.Getenv("SIDECAR_PORT")
	} else {
		designatedGRPCPort, _ := strconv.Atoi(os.Getenv("DESIGNATED_GRPC_PORT"))
		nextServicePort = strconv.Itoa(designatedGRPCPort + 1)
	}

	ms := &MsCommunication{
		SecureChannel:   SecureChannel{grpcAddr, nextServicePort},
		NextServicePort: nextServicePort,
	}

	// Assuming you have a function to create a new MicroserviceStub
	ms.Client = NewMicroserviceStub(ms.SecureChannel)

	return ms
}

func main() {
	defer logger.Sync() // flushes buffer, if any

}
