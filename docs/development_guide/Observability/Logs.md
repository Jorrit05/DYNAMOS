# Logs

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