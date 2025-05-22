# Tracing
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

### Blank space in traces
In the Jaeger UI, the tracing output sometimes shows blank spaces (periods where no spans are visible). This often occurs when something is waiting, such as during a data request. These gaps typically indicate that the system is waiting for something, such as service startup or coordination, unless traces are missing or broken.

These blank spaces do not necessarily need to be filled. In many cases, it would be unnecessarily complex or not worth the effort. For example, during a data request, multiple gRPC functions are used to perform health checks before the query can begin. Adding tracing for each of these would result in a large number of additional spans, increasing complexity without much benefit. Moreover, gRPC is primarily used for communication between the services, such as carrying the trace context between services. Since context propagation begins only after the gRPC channel is established, it’s not straightforward to trace the operations that occur before gRPC is invoked. Tracing these pre-gRPC steps would require deep instrumentation of low-level communication or setup code, which is often impractical or infeasible.

### Code Snippets

- Initialize a trace
```go
oce, err := lib.InitTracer(serviceName)
```

- Start remote parent span
```go
ctx, span, err := lib.StartRemoteParentSpan(ctx, serviceName+"/func: <name of function>", msComm.Traces) // The traces are of type make(map[string][]byte)
```

- Debug traces by printing the values
```go
// In pkg/lib/tracing.go there is a function called PrettyPrintSpanContext(), which you can use to print the span information
// In your code after creating a span, add a line calling this function:
<add above function to start a remote span with lib.StartRemoteParentSpan>
if err != nil {
	logger.Sugar().Warnf("Error starting span: %v", err)
}
// Print traces for debugging
lib.PrettyPrintSpanContext(span.SpanContext())
// This will print it in a format like this for example:
sql-algorithm Trace ID: 0ffa2f5cd4a3c4e15b27a50a90821bbd
sql-algorithm Span ID: 3fd1f17c653dfa02
sql-algorithm Trace options: 1
sql-algorithm Trace IsSampled: true
// You can use the Trace ID in the Jaeger UI (after forwarding the UI, see docs/helpers/cheat_sheet.md) to view the trace information if it was created successfully.
```
TODO: Add more examples 