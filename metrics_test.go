package metricsx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	t.Run("WithHelp sets help text", func(t *testing.T) {
		opts := &Options{}
		WithHelp("test help")(opts)
		assert.Equal(t, "test help", opts.Help)
	})

	t.Run("WithLabels sets labels", func(t *testing.T) {
		opts := &Options{}
		WithLabels("label1", "label2")(opts)
		assert.Equal(t, []string{"label1", "label2"}, opts.Labels)
	})

	t.Run("WithBuckets sets buckets", func(t *testing.T) {
		opts := &Options{}
		buckets := []float64{0.1, 0.5, 1.0}
		WithBuckets(buckets...)(opts)
		assert.Equal(t, buckets, opts.Buckets)
	})

	t.Run("WithObjectives sets objectives", func(t *testing.T) {
		opts := &Options{}
		objectives := map[float64]float64{0.5: 0.05, 0.9: 0.01}
		WithObjectives(objectives)(opts)
		assert.Equal(t, objectives, opts.Objectives)
	})

	t.Run("WithNamespace sets namespace", func(t *testing.T) {
		opts := &Options{}
		WithNamespace("myapp")(opts)
		assert.Equal(t, "myapp", opts.Namespace)
	})

	t.Run("WithSubsystem sets subsystem", func(t *testing.T) {
		opts := &Options{}
		WithSubsystem("http")(opts)
		assert.Equal(t, "http", opts.Subsystem)
	})

	t.Run("multiple options can be combined", func(t *testing.T) {
		opts := &Options{}

		applyOptions := []Option{
			WithHelp("combined test"),
			WithLabels("method", "status"),
			WithNamespace("app"),
			WithSubsystem("api"),
		}

		for _, opt := range applyOptions {
			opt(opts)
		}

		assert.Equal(t, "combined test", opts.Help)
		assert.Equal(t, []string{"method", "status"}, opts.Labels)
		assert.Equal(t, "app", opts.Namespace)
		assert.Equal(t, "api", opts.Subsystem)
	})
}

func TestNewMetrics(t *testing.T) {
	t.Run("creates metrics wrapper with prometheus", func(t *testing.T) {
		provider := newPrometheusProvider(PrometheusConfig{Port: 0, Path: "/metrics"}, getTestLogger())
		metrics := &metricsImpl{provider: provider, logger: getTestLogger()}

		assert.NotNil(t, metrics)
	})

	t.Run("counter delegates to provider", func(t *testing.T) {
		provider := newNoopProvider()
		metrics := &metricsImpl{provider: provider, logger: getTestLogger()}

		counter := metrics.Counter("test_counter", WithHelp("test"))
		assert.NotNil(t, counter)
	})

	t.Run("gauge delegates to provider", func(t *testing.T) {
		provider := newNoopProvider()
		metrics := &metricsImpl{provider: provider, logger: getTestLogger()}

		gauge := metrics.Gauge("test_gauge", WithHelp("test"))
		assert.NotNil(t, gauge)
	})

	t.Run("histogram delegates to provider", func(t *testing.T) {
		provider := newNoopProvider()
		metrics := &metricsImpl{provider: provider, logger: getTestLogger()}

		histogram := metrics.Histogram("test_histogram", WithHelp("test"))
		assert.NotNil(t, histogram)
	})

	t.Run("summary delegates to provider", func(t *testing.T) {
		provider := newNoopProvider()
		metrics := &metricsImpl{provider: provider, logger: getTestLogger()}

		summary := metrics.Summary("test_summary", WithHelp("test"))
		assert.NotNil(t, summary)
	})
}

func TestApplyOptions(t *testing.T) {
	t.Run("applies multiple options", func(t *testing.T) {
		opts := applyOptions(
			WithHelp("test metric"),
			WithLabels("label1", "label2"),
			WithNamespace("test"),
		)

		assert.Equal(t, "test metric", opts.Help)
		assert.Equal(t, []string{"label1", "label2"}, opts.Labels)
		assert.Equal(t, "test", opts.Namespace)
	})

	t.Run("applies no options", func(t *testing.T) {
		opts := applyOptions()

		assert.Empty(t, opts.Help)
		assert.Nil(t, opts.Labels)
		assert.Empty(t, opts.Namespace)
	})
}
