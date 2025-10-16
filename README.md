# metricsx

**Observability module for metrics collection in gostratum framework**

`metricsx` provides a unified interface for collecting and exposing application metrics with support for multiple providers (Prometheus, StatsD, etc.). It follows the gostratum philosophy of fx-first design and integrates seamlessly with other modules.

## Features

- 🎯 **Provider-agnostic interface** - Switch between Prometheus, StatsD, or custom providers
- 📊 **Standard metric types** - Counter, Gauge, Histogram, Summary
- 🔧 **Fx-first design** - Seamless dependency injection
- 🎨 **Zero-allocation fast paths** - Performance-optimized
- 🔌 **Easy integration** - Works with `httpx`, `dbx`, and other modules
- 📝 **Type-safe** - Compile-time safety with Go interfaces
- 🧪 **Testable** - Includes no-op provider for testing

## Installation

```bash
go get github.com/gostratum/metricsx
```

## Quick Start

### Basic Usage

```go
package main

import (
    "go.uber.org/fx"
    
    "github.com/gostratum/core"
    "github.com/gostratum/metricsx"
)

func main() {
    fx.New(
        core.Module(),
        metricsx.Module(),
        
        fx.Invoke(func(metrics metricsx.Metrics) {
            // Create a counter
            requestCounter := metrics.Counter("http_requests_total",
                metricsx.WithLabels("method", "path", "status"),
                metricsx.WithHelp("Total HTTP requests"),
            )
            
            // Increment counter
            requestCounter.Inc("GET", "/api/users", "200")
            
            // Create a histogram
            duration := metrics.Histogram("request_duration_seconds",
                metricsx.WithLabels("method", "path"),
                metricsx.WithHelp("Request duration in seconds"),
                metricsx.WithBuckets(0.001, 0.01, 0.1, 1, 10),
            )
            
            // Observe duration
            timer := duration.Timer("GET", "/api/users")
            defer timer.ObserveDuration()
            
            // ... your code here ...
        }),
    ).Run()
}
```

### Configuration

Add to your `config.yaml`:

```yaml
metrics:
  enabled: true
  provider: prometheus
  prometheus:
    namespace: myapp
    subsystem: api
    path: /metrics
    port: 0  # 0 = use main HTTP server, or specify separate port
    enable_process_metrics: true
    enable_go_metrics: true
```

## Metric Types

### Counter

Monotonically increasing value (e.g., request count, error count):

```go
counter := metrics.Counter("operations_total",
    metricsx.WithLabels("operation", "status"),
    metricsx.WithHelp("Total operations"),
)

counter.Inc("create", "success")           // Increment by 1
counter.Add(5, "delete", "success")       // Increment by 5
```

### Gauge

Value that can go up and down (e.g., active connections, queue size):

```go
gauge := metrics.Gauge("active_connections",
    metricsx.WithLabels("service"),
    metricsx.WithHelp("Current active connections"),
)

gauge.Set(100, "api")      // Set to specific value
gauge.Inc("api")            // Increment by 1
gauge.Dec("api")            // Decrement by 1
gauge.Add(10, "api")        // Add 10
gauge.Sub(5, "api")         // Subtract 5
```

### Histogram

Samples observations and counts them in buckets (e.g., request duration, response size):

```go
histogram := metrics.Histogram("request_duration_seconds",
    metricsx.WithLabels("method", "endpoint"),
    metricsx.WithHelp("Request duration in seconds"),
    metricsx.WithBuckets(0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5),
)

// Manual observation
histogram.Observe(0.025, "GET", "/api/users")

// Or use timer for convenience
timer := histogram.Timer("POST", "/api/orders")
// ... do work ...
timer.ObserveDuration()
```

### Summary

Similar to histogram but with quantiles (e.g., response time percentiles):

```go
summary := metrics.Summary("response_time_seconds",
    metricsx.WithLabels("service"),
    metricsx.WithHelp("Response time in seconds"),
    metricsx.WithObjectives(map[float64]float64{
        0.5:  0.05,   // 50th percentile with 5% error
        0.9:  0.01,   // 90th percentile with 1% error
        0.99: 0.001,  // 99th percentile with 0.1% error
    }),
)

summary.Observe(0.123, "api")
```

## Integration with httpx

Automatic HTTP metrics middleware:

```go
package main

import (
    "go.uber.org/fx"
    
    "github.com/gostratum/core"
    "github.com/gostratum/httpx"
    "github.com/gostratum/metricsx"
)

func main() {
    fx.New(
        core.Module(),
        metricsx.Module(),
        httpx.Module(),
        // httpx automatically detects metricsx and adds middleware
    ).Run()
}
```

This automatically exposes:
- `http_requests_total{method, path, status}` - Total requests
- `http_request_duration_seconds{method, path, status}` - Request duration
- `http_requests_in_flight{method}` - Current in-flight requests

## Integration with dbx

Automatic database query metrics:

```go
package main

import (
    "go.uber.org/fx"
    
    "github.com/gostratum/core"
    "github.com/gostratum/dbx"
    "github.com/gostratum/metricsx"
)

func main() {
    fx.New(
        core.Module(),
        metricsx.Module(),
        dbx.Module(),
        // dbx automatically detects metricsx and adds hooks
    ).Run()
}
```

This automatically exposes:
- `db_query_duration_seconds{operation, table}` - Query duration
- `db_queries_total{operation, table, status}` - Total queries
- `db_connections_open` - Current open connections

## Custom Metrics

### Application Metrics

```go
type OrderService struct {
    ordersCreated metricsx.Counter
    orderValue    metricsx.Histogram
}

func NewOrderService(metrics metricsx.Metrics) *OrderService {
    return &OrderService{
        ordersCreated: metrics.Counter("orders_created_total",
            metricsx.WithLabels("status"),
            metricsx.WithHelp("Total orders created"),
        ),
        orderValue: metrics.Histogram("order_value_dollars",
            metricsx.WithLabels("status"),
            metricsx.WithHelp("Order value in dollars"),
            metricsx.WithBuckets(10, 50, 100, 500, 1000, 5000),
        ),
    }
}

func (s *OrderService) CreateOrder(value float64) error {
    s.ordersCreated.Inc("pending")
    s.orderValue.Observe(value, "pending")
    // ... create order ...
    return nil
}
```

## Providers

### Prometheus (Default)

Exposes metrics in Prometheus format at `/metrics`:

```yaml
metrics:
  provider: prometheus
  prometheus:
    path: /metrics
    namespace: myapp
    subsystem: orders
```

Access metrics: `http://localhost:8080/metrics`

Example output:
```
# HELP myapp_orders_http_requests_total Total HTTP requests
# TYPE myapp_orders_http_requests_total counter
myapp_orders_http_requests_total{method="GET",path="/api/orders",status="200"} 42
```

### No-op Provider

For testing and development:

```yaml
metrics:
  provider: noop
```

All metric operations become no-ops (zero overhead).

## Best Practices

### 1. **Use Appropriate Metric Types**

- **Counter**: Things that only increase (requests, errors, bytes sent)
- **Gauge**: Things that go up and down (connections, queue size, memory)
- **Histogram**: Distributions and timing (duration, size)
- **Summary**: Quantiles over time windows (rarely needed, use histogram instead)

### 2. **Label Cardinality**

⚠️ **Avoid high-cardinality labels** (e.g., user IDs, timestamps):

```go
// ❌ BAD - Too many unique label combinations
counter.Inc(userID, timestamp, requestID)

// ✅ GOOD - Limited label cardinality
counter.Inc("POST", "/api/users", "200")
```

### 3. **Metric Naming**

Follow Prometheus naming conventions:

```go
// ✅ GOOD
metrics.Counter("http_requests_total")           // plural + _total
metrics.Histogram("http_request_duration_seconds") // singular + unit
metrics.Gauge("active_connections")               // describe current state

// ❌ BAD
metrics.Counter("request")                        // too vague
metrics.Histogram("latency")                      // no unit
```

### 4. **Reuse Metrics**

Create metrics once and reuse them:

```go
// ✅ GOOD - Create once
type Handler struct {
    requestCounter metricsx.Counter
}

func NewHandler(metrics metricsx.Metrics) *Handler {
    return &Handler{
        requestCounter: metrics.Counter("requests_total"),
    }
}

// ❌ BAD - Creating metrics in hot path
func (h *Handler) Handle() {
    counter := h.metrics.Counter("requests_total") // Don't do this!
    counter.Inc()
}
```

### 5. **Use Timers**

For measuring durations, use the built-in timer:

```go
// ✅ GOOD
timer := duration.Timer("GET", "/api/users")
defer timer.ObserveDuration()

// ✅ ALSO GOOD - If you need the duration value
timer := duration.Timer("GET", "/api/users")
// ... do work ...
elapsed := timer.Stop()
logger.Info("request completed", "duration", elapsed)
```

## Testing

Use the no-op provider in tests:

```go
func TestOrderService(t *testing.T) {
    metrics := noop.NewProvider()
    
    service := NewOrderService(metrics)
    // ... test your service ...
}
```

Or create a test helper:

```go
import (
    "testing"
    "github.com/gostratum/metricsx/noop"
)

func NewTestMetrics(t *testing.T) metricsx.Metrics {
    provider := noop.NewProvider()
    return &metricsImpl{provider: provider}
}
```

## Advanced Usage

### Custom Buckets

For histograms with specific requirements:

```go
// Latency-focused buckets (milliseconds to seconds)
latencyBuckets := []float64{
    0.001, 0.002, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10,
}

histogram := metrics.Histogram("api_latency_seconds",
    metricsx.WithBuckets(latencyBuckets...),
)

// Size-focused buckets (bytes to megabytes)
sizeBuckets := []float64{
    1024, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216,
}

histogram := metrics.Histogram("response_size_bytes",
    metricsx.WithBuckets(sizeBuckets...),
)
```

### Namespacing

Override namespace/subsystem per metric:

```go
counter := metrics.Counter("cache_hits_total",
    metricsx.WithNamespace("myapp"),
    metricsx.WithSubsystem("redis"),
    metricsx.WithHelp("Total cache hits"),
)
// Metric name: myapp_redis_cache_hits_total
```

## Dependencies

- **Core**: `github.com/gostratum/core` (for config and logging)
- **Prometheus**: `github.com/prometheus/client_golang` (Prometheus provider)
- **Fx**: `go.uber.org/fx` (dependency injection)

## Architecture

```
metricsx/
├── metrics.go              # Core interfaces
├── config.go               # Configuration
├── module.go               # Fx module
├── prometheus/             # Prometheus implementation
│   └── provider.go
└── noop/                   # No-op implementation
    └── provider.go
```

## Roadmap

- [ ] StatsD provider
- [ ] DataDog provider
- [ ] Custom exporter support
- [ ] Metric aggregation
- [ ] Exemplars support (OpenTelemetry)

## License

MIT License - see LICENSE file for details

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

## Related Modules

- [`core`](../core/README.md) - Foundation module
- [`httpx`](../httpx/README.md) - HTTP server with automatic metrics
- [`dbx`](../dbx/README.md) - Database with automatic query metrics
- [`tracingx`](../tracingx/README.md) - Distributed tracing
