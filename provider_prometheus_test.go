package metricsx

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gostratum/core/logx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestLogger() logx.Logger {
	return logx.NewNoopLogger()
}

func TestPrometheusProvider(t *testing.T) {
	logger := getTestLogger()

	t.Run("creates prometheus provider", func(t *testing.T) {
		config := PrometheusConfig{
			Port: 0, // Don't start server
			Path: "/metrics",
		}

		provider := newPrometheusProvider(config, logger)
		assert.NotNil(t, provider)
	})

	t.Run("counter operations", func(t *testing.T) {
		config := PrometheusConfig{Port: 0, Path: "/metrics"}
		provider := newPrometheusProvider(config, logger)

		counter := provider.Counter("test_counter", &Options{
			Help:   "Test counter",
			Labels: []string{"method"},
		})

		require.NotNil(t, counter)

		// Test Inc
		counter.Inc("GET")
		counter.Inc("POST")

		// Test Add
		counter.Add(5.0, "GET")
	})

	t.Run("gauge operations", func(t *testing.T) {
		config := PrometheusConfig{Port: 0, Path: "/metrics"}
		provider := newPrometheusProvider(config, logger)

		gauge := provider.Gauge("test_gauge", &Options{
			Help:   "Test gauge",
			Labels: []string{"type"},
		})

		require.NotNil(t, gauge)

		// Test Set
		gauge.Set(42.0, "cpu")

		// Test Inc
		gauge.Inc("memory")

		// Test Dec
		gauge.Dec("memory")

		// Test Add
		gauge.Add(10.0, "cpu")

		// Test Sub
		gauge.Sub(5.0, "cpu")
	})

	t.Run("histogram operations", func(t *testing.T) {
		config := PrometheusConfig{Port: 0, Path: "/metrics"}
		provider := newPrometheusProvider(config, logger)

		histogram := provider.Histogram("test_histogram", &Options{
			Help:    "Test histogram",
			Labels:  []string{"endpoint"},
			Buckets: []float64{0.1, 0.5, 1.0, 5.0},
		})

		require.NotNil(t, histogram)

		// Test Observe
		histogram.Observe(0.25, "/api/users")
		histogram.Observe(1.5, "/api/users")

		// Test Timer
		timer := histogram.Timer("/api/products")
		require.NotNil(t, timer)
		time.Sleep(10 * time.Millisecond)
		duration := timer.Stop()
		assert.Greater(t, duration, time.Duration(0))
	})

	t.Run("summary operations", func(t *testing.T) {
		config := PrometheusConfig{Port: 0, Path: "/metrics"}
		provider := newPrometheusProvider(config, logger)

		summary := provider.Summary("test_summary", &Options{
			Help:   "Test summary",
			Labels: []string{"service"},
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		})

		require.NotNil(t, summary)

		// Test Observe
		summary.Observe(0.5, "auth")
		summary.Observe(1.2, "auth")
		summary.Observe(0.8, "billing")
	})

	t.Run("reuses existing metrics", func(t *testing.T) {
		config := PrometheusConfig{Port: 0, Path: "/metrics"}
		provider := newPrometheusProvider(config, logger)

		counter1 := provider.Counter("reuse_test", &Options{
			Help:   "Test",
			Labels: []string{"label"},
		})

		counter2 := provider.Counter("reuse_test", &Options{
			Help:   "Test",
			Labels: []string{"label"},
		})

		// Should return the same instance
		assert.Equal(t, counter1, counter2)
	})

	t.Run("applies namespace and subsystem", func(t *testing.T) {
		config := PrometheusConfig{Port: 0, Path: "/metrics"}
		provider := newPrometheusProvider(config, logger)

		counter := provider.Counter("requests_total", &Options{
			Help:      "Total requests",
			Labels:    []string{"method"},
			Namespace: "myapp",
			Subsystem: "http",
		})

		require.NotNil(t, counter)
		counter.Inc("GET")
	})
}

func TestPrometheusLifecycle(t *testing.T) {
	logger := getTestLogger()

	t.Run("starts and stops server", func(t *testing.T) {
		config := PrometheusConfig{
			Port: 19090, // Use non-standard port for testing
			Path: "/metrics",
		}

		provider := newPrometheusProvider(config, logger)

		ctx := context.Background()

		// Start
		err := provider.Start(ctx)
		require.NoError(t, err)

		// Give server time to start
		time.Sleep(100 * time.Millisecond)

		// Verify server is running
		resp, err := http.Get("http://localhost:19090/metrics")
		if err == nil {
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}

		// Stop
		err = provider.Stop(ctx)
		assert.NoError(t, err)
	})

	t.Run("skip server start when port is 0", func(t *testing.T) {
		config := PrometheusConfig{
			Port: 0,
			Path: "/metrics",
		}

		provider := newPrometheusProvider(config, logger)

		ctx := context.Background()
		err := provider.Start(ctx)
		assert.NoError(t, err)

		// Stop should be no-op
		err = provider.Stop(ctx)
		assert.NoError(t, err)
	})
}

func TestPrometheusMetricsExposure(t *testing.T) {
	logger := getTestLogger()

	t.Run("exposes metrics via HTTP", func(t *testing.T) {
		config := PrometheusConfig{
			Port: 19091,
			Path: "/metrics",
		}

		provider := newPrometheusProvider(config, logger)

		// Create some metrics
		counter := provider.Counter("http_requests_total", &Options{
			Help:   "Total HTTP requests",
			Labels: []string{"method", "status"},
		})
		counter.Inc("GET", "200")
		counter.Add(5, "POST", "201")

		gauge := provider.Gauge("active_connections", &Options{
			Help: "Active connections",
		})
		gauge.Set(42)

		// Start server
		ctx := context.Background()
		err := provider.Start(ctx)
		require.NoError(t, err)
		defer provider.Stop(ctx)

		// Give server time to start
		time.Sleep(100 * time.Millisecond)

		// Fetch metrics
		resp, err := http.Get("http://localhost:19091/metrics")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		metrics := string(body)

		// Verify metrics are present
		assert.Contains(t, metrics, "http_requests_total")
		assert.Contains(t, metrics, "active_connections")
	})
}

func TestPrometheusConfig(t *testing.T) {
	logger := getTestLogger()

	t.Run("enables process metrics", func(t *testing.T) {
		config := PrometheusConfig{
			Port:                 0,
			Path:                 "/metrics",
			EnableProcessMetrics: true,
		}

		provider := newPrometheusProvider(config, logger)
		assert.NotNil(t, provider)
	})

	t.Run("enables go metrics", func(t *testing.T) {
		config := PrometheusConfig{
			Port:            0,
			Path:            "/metrics",
			EnableGoMetrics: true,
		}

		provider := newPrometheusProvider(config, logger)
		assert.NotNil(t, provider)
	})

	t.Run("enables both process and go metrics", func(t *testing.T) {
		config := PrometheusConfig{
			Port:                 0,
			Path:                 "/metrics",
			EnableProcessMetrics: true,
			EnableGoMetrics:      true,
		}

		provider := newPrometheusProvider(config, logger)
		assert.NotNil(t, provider)
	})
}

func TestPrometheusTimer(t *testing.T) {
	logger := getTestLogger()

	t.Run("timer measures duration", func(t *testing.T) {
		config := PrometheusConfig{Port: 0, Path: "/metrics"}
		provider := newPrometheusProvider(config, logger)

		histogram := provider.Histogram("request_duration", &Options{
			Help:    "Request duration",
			Labels:  []string{"endpoint"},
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0},
		})

		timer := histogram.Timer("/api/test")

		// Simulate work
		time.Sleep(20 * time.Millisecond)

		duration := timer.Stop()
		assert.GreaterOrEqual(t, duration, 20*time.Millisecond)
	})

	t.Run("timer observe duration", func(t *testing.T) {
		config := PrometheusConfig{Port: 0, Path: "/metrics"}
		provider := newPrometheusProvider(config, logger)

		histogram := provider.Histogram("processing_time", &Options{
			Help:   "Processing time",
			Labels: []string{"task"},
		})

		timer := histogram.Timer("compute")
		time.Sleep(10 * time.Millisecond)

		// ObserveDuration should record the time
		timer.ObserveDuration()
	})
}
