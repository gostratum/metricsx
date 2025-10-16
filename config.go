package metricsx

import (
	"github.com/gostratum/core/configx"
)

// Config contains configuration for the metrics module
type Config struct {
	// Enabled determines if metrics collection is enabled
	Enabled bool `mapstructure:"enabled" default:"true"`

	// Provider specifies which metrics provider to use (prometheus, noop)
	Provider string `mapstructure:"provider" default:"prometheus"`

	// Prometheus configuration
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
}

// Prefix enables configx.Bind
func (Config) Prefix() string { return "metrics" }

// PrometheusConfig contains Prometheus-specific configuration
type PrometheusConfig struct {
	// Namespace for all metrics
	Namespace string `mapstructure:"namespace" default:""`

	// Subsystem for all metrics
	Subsystem string `mapstructure:"subsystem" default:""`

	// Path where metrics are exposed (default: /metrics)
	Path string `mapstructure:"path" default:"/metrics"`

	// Port for the metrics HTTP server (if separate from main app)
	// If 0, metrics will be exposed on the main HTTP server
	Port int `mapstructure:"port" default:"0"`

	// EnableProcessMetrics enables Go process metrics
	EnableProcessMetrics bool `mapstructure:"enable_process_metrics" default:"true"`

	// EnableGoMetrics enables Go runtime metrics
	EnableGoMetrics bool `mapstructure:"enable_go_metrics" default:"true"`
}

// NewConfig creates a new Config from the configuration loader
func NewConfig(loader configx.Loader) (Config, error) {
	var cfg Config
	if err := loader.Bind(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
