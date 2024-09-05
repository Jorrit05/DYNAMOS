# Introduction 

This guide provides information on creating a new Microservice. Currently services in GO and Python are implemented, however, you may decide to implement in other languages, but there is no support provided for this currently. 

TODO: Probably move this to its own document of background information
# Common components
All services have a few components in common, these should be implemented in all services to ensure that consistency is maintained throughout the entire code base. These components are described below:

## Tracing
Tracing is a way to track the journey of a request as it moves through different services in a system. Think of it like a map that shows you how a user's action, like clicking a button, travels through multiple backend services to get a result. In DYNAMOS, we use [Jaeger](https://www.jaegertracing.io/) as a distributed tracing agent. 

There are a few terminologies that you should know:

- **Trace**: A trace is a record of a request as it flows through various services.

- **Span**: Each step or operation that a service performs during that request is called a span. Spans together form a trace.

- **Context**: Metadata or information that gets passed along with a request as it moves through different services. This information helps track the request across multiple services and combine all the spans (steps) into a single trace.

DYNAMOS has a library (`go/pkg/lib/tracing.go`) that handles the code related to tracing, you should use this library for your implementation.

The generated traces can be viewed in the deployed Jaeger instance. To view the Jaeger UI, first run:
```sh
kubectl port-forward -n linkerd-jaeger service/jaeger 16686:16686
```
The UI can then be viewed [here](http://localhost:16686/jaeger/search) 

### Code Snippets

- Initialize a trace
```go
oce, err := lib.InitTracer(serviceName)
```

- Start remote parent span
```go
ctx, span, err := lib.StartRemoteParentSpan(ctx, serviceName+"/func: <name of function>", msComm.Traces) // The traces are of type make(map[string][]byte)
```
TODO: Add more examples 

## Logs

Logging is used within the codebase to make debugging easier, make sure to add logs frequently with the appropriate log level.

For example:
- A debug level log:
```go
logger.Sugar().Debugf("Starting %s service", serviceName)
```
- Fatal error log
```go
logger.Sugar().Fatalf("Failed with error: %v", err)
```
# GO Service

## Service Folder

A new Go service should be created in the `go/cmd/` directory.  

### Configs
Each Go service requires a local and a prod config file. The local config can be used outside of the docker and k8s environment, whereas the production configs can be used within the k8s environment and servers.

A standard template for the configs are provided below, they can be copied and pasted into your new service. If you need globally accessible variables, you may add them here as well.

An example of each is as follows:

#### Local Config (config_local.go)
```go
//go:build local
// +build local

package main

import "go.uber.org/zap"

// Change to the desired output log level
var logLevel = zap.DebugLevel

// Change to name of your service
var serviceName = "serviceNameHere"

var etcdEndpoints = "http://localhost:30005"

var grpcAddr = "localhost:"
```
#### Prod Config (config_prod.go)
```go
//go:build !local
// +build !local

package main

import "go.uber.org/zap"

var logLevel = zap.DebugLevel

// Change to name of your service
var serviceName = "serviceNameHere"

var etcdEndpoints = "http://etcd-0.etcd-headless.core.svc.cluster.local:2379,http://etcd-1.etcd-headless.core.svc.cluster.local:2379,http://etcd-2.etcd-headless.core.svc.cluster.local:2379"

var grpcAddr = "localhost:"
var tracingHost = "collector.linkerd-jaeger:55678"
```

### Main (main.go)
The `main.go` file is the entry file to your microservice. There is a common pattern that is used with all DYNAMOS services. 

Imports for the service are handled here, communication configurations are set here and the message handler is initialized here.  

All of the mentioned above are described below:

#### Imports
The following are the standard packages used in the DYNAMOS services, usually IDEs will automatically update this for you when adding new dependencies.
```go
package main

import (
	"context"
	"os"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	"github.com/Jorrit05/DYNAMOS/pkg/msinit"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

var (
	logger      = lib.InitLogger(logLevel)
	COORDINATOR = make(chan struct{})
)

```

#### Message Handler
Each service requires a function which will be responsible of handling communication between other services
```go
func messageHandler(config *msinit.Configuration) func(ctx context.Context, msComm *pb.MicroserviceCommunication) error {
	return func(ctx context.Context, msComm *pb.MicroserviceCommunication) error {
		ctx, span, err := lib.StartRemoteParentSpan(ctx, serviceName+"/func: messageHandler", msComm.Traces)
		if err != nil {
			logger.Sugar().Warnf("Error starting span: %v", err)
		}
		defer span.End()

		// Wait till all services and connections have started
		<-COORDINATOR

		switch msComm.RequestType {
		case "CHANGE_TO_YOUR_REQUEST_TYPE":
      //  Change the function below with your implementation
      //    The code for this should be in the application logic file
			err := handleDataRequest(ctx, msComm)
			if err != nil {
				logger.Sugar().Errorf("Failed to process %s message: %v", msComm.RequestType, err)
			}

		default:
			logger.Sugar().Errorf("Unknown RequestType type: %v", msComm.RequestType)
		}

    // Send data to the next MS, handled by the msInit configuration 
		config.NextClient.SendData(ctx, msComm)

		close(config.StopMicroservice)
		return nil
	}
}
```

#### Main function
Each Go MS requires a main function, a typical MS main function can look like so:
```go
func main() {
	logger.Sugar().Debugf("Starting %s service", serviceName)

	oce, err := lib.InitTracer(serviceName)
	if err != nil {
		logger.Sugar().Fatalf("Failed to create ocagent-exporter: %v", err)
	}

  //  
	config, err := msinit.NewConfiguration(context.Background(), serviceName, grpcAddr, COORDINATOR, messageHandler)
	if err != nil {
		logger.Sugar().Fatalf("%v", err)
	}

	// Wait here until the message arrives in the messageHandler
	<-config.StopMicroservice

	config.SafeExit(oce, serviceName)
	os.Exit(0)
}
```

### Application Logic
All other application logic based code should be in separate Go files, this will allow the consistency and quality of the code base to be maintained.
Some services have a specific file named `application_logic.go`, this is our convention.

### Makefile

When you want to build and test your new service, make sure to add it the Makefile corresponding with the language, by adding the folder name to a task such as `sql_microservices` it will automatically be build the next time the corresponding make task is run.

For example, lets add a service called 'my-new-service' to the 'sql_microservices' variable in the Go Makefile:
```Makefile
sql_microservices := sql-algorithm sql-anonymize sql-aggregate sql-test my-new-service
```
Now when running `make sql_microservices` in your terminal, that service will be build along with the others.

```
