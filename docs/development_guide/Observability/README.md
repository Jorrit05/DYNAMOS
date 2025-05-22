# Observability

Observability is the ability to understand a system’s internal state from the data it produces—like logs, metrics, and traces. It’s essential in distributed systems (e.g., microservices, Kubernetes) for diagnosing issues, optimizing performance, and improving reliability. Unlike basic monitoring, which shows what’s wrong, observability reveals why it’s happening, enabling faster debugging and better system insights.

Key pillars of observability include:
- Logs: Time-stamped records of discrete events, useful for debugging and auditing.
- Metrics: Numeric values representing the state or behavior of a component over time, ideal for alerting and performance tracking. Prometheus is responsible for metric collection in DYNAMOS.
- Tracing: Records of request flows through systems, helping to pinpoint latency bottlenecks or failure points in distributed environments.

More information about specific topics can be found here:
- [Logs](./Logs.md)
- [Tracing](./Tracing.md)

## Centralized Observability
In DYNAMOS, observability is centralized using Grafana. The Grafana UI can be accessed by opening http://localhost:30001/

### Logs
In Grafana, you can access logs by following these steps:
1. In the Grafana UI, go to Explore > Select Loki next to Outline
2. Create a query, such as using the Builder to select "app" and a component of DYNAMOS, such as "api-gateway"
3. Optionally you can add a filter, such as "Line contains" for ERROR or INFO
4. Execute the query and view the results

### Tracing
In Grafana, you can access the traces by following these steps:
1. In the Grafana UI, go to Explore > Select Jaeger next to Outline
2. Create a query by selecting Search as the Query type, and select a service like "api-gateway"
3. Execute the query and select a trace to display

Alternatively, you can port-forward the Jaeger UI to view the logs, as explained in the [Tracing.md file](./Tracing.md)


## Other processes 
Some other processes that can be accessed are:


TODO: list of processes that can be accessed.




TODO: add Grafana for centralized visualization, can be accessed at...
