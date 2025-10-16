package metricsx

import (
	"context"
	"time"
)

// Metrics is the main interface for recording metrics
// This abstraction allows for different metric providers (Prometheus, StatsD, etc.)
type Metrics interface {
	// Counter creates or retrieves a counter metric
	Counter(name string, opts ...Option) Counter

	// Gauge creates or retrieves a gauge metric
	Gauge(name string, opts ...Option) Gauge

	// Histogram creates or retrieves a histogram metric
	Histogram(name string, opts ...Option) Histogram

	// Summary creates or retrieves a summary metric
	Summary(name string, opts ...Option) Summary
}

// Counter is a monotonically increasing metric
type Counter interface {
	// Inc increments the counter by 1
	Inc(labels ...string)

	// Add increments the counter by the given value
	Add(value float64, labels ...string)
}

// Gauge is a metric that can go up and down
type Gauge interface {
	// Set sets the gauge to the given value
	Set(value float64, labels ...string)

	// Inc increments the gauge by 1
	Inc(labels ...string)

	// Dec decrements the gauge by 1
	Dec(labels ...string)

	// Add adds the given value to the gauge
	Add(value float64, labels ...string)

	// Sub subtracts the given value from the gauge
	Sub(value float64, labels ...string)
}

// Histogram samples observations and counts them in configurable buckets
type Histogram interface {
	// Observe adds a single observation to the histogram
	Observe(value float64, labels ...string)

	// Timer creates a timer that will observe the duration when stopped
	Timer(labels ...string) Timer
}

// Summary samples observations over a sliding time window
type Summary interface {
	// Observe adds a single observation to the summary
	Observe(value float64, labels ...string)
}

// Timer provides a convenient way to measure durations
type Timer interface {
	// ObserveDuration observes the duration since the timer was created
	ObserveDuration()

	// Stop stops the timer and returns the duration
	Stop() time.Duration
}

// Option configures metric options
type Option func(*Options)

// Options contains configuration for metrics
type Options struct {
	// Help text describing the metric
	Help string

	// Labels are the label names for this metric
	Labels []string

	// Buckets for histograms (optional, uses defaults if not set)
	Buckets []float64

	// Objectives for summaries (optional, uses defaults if not set)
	Objectives map[float64]float64

	// Namespace for the metric (optional)
	Namespace string

	// Subsystem for the metric (optional)
	Subsystem string
}

// WithHelp sets the help text for the metric
func WithHelp(help string) Option {
	return func(o *Options) {
		o.Help = help
	}
}

// WithLabels sets the label names for the metric
func WithLabels(labels ...string) Option {
	return func(o *Options) {
		o.Labels = labels
	}
}

// WithBuckets sets the buckets for histogram metrics
func WithBuckets(buckets ...float64) Option {
	return func(o *Options) {
		o.Buckets = buckets
	}
}

// WithObjectives sets the objectives for summary metrics
func WithObjectives(objectives map[float64]float64) Option {
	return func(o *Options) {
		o.Objectives = objectives
	}
}

// WithNamespace sets the namespace for the metric
func WithNamespace(namespace string) Option {
	return func(o *Options) {
		o.Namespace = namespace
	}
}

// WithSubsystem sets the subsystem for the metric
func WithSubsystem(subsystem string) Option {
	return func(o *Options) {
		o.Subsystem = subsystem
	}
}

// applyOptions applies the given options and returns the final Options
func applyOptions(opts ...Option) *Options {
	options := &Options{
		Labels:     []string{},
		Buckets:    DefaultBuckets,
		Objectives: DefaultObjectives,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// DefaultBuckets are the default histogram buckets
var DefaultBuckets = []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10}

// DefaultObjectives are the default summary objectives
var DefaultObjectives = map[float64]float64{
	0.5:  0.05,
	0.9:  0.01,
	0.99: 0.001,
}

// Provider is the interface that metric providers must implement
type Provider interface {
	// Counter creates or retrieves a counter
	Counter(name string, options *Options) Counter

	// Gauge creates or retrieves a gauge
	Gauge(name string, options *Options) Gauge

	// Histogram creates or retrieves a histogram
	Histogram(name string, options *Options) Histogram

	// Summary creates or retrieves a summary
	Summary(name string, options *Options) Summary

	// Start starts the metrics provider (e.g., HTTP server for Prometheus)
	Start(ctx context.Context) error

	// Stop stops the metrics provider
	Stop(ctx context.Context) error
}
