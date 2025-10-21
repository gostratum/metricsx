package metricsx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsSanitizeAndSummary(t *testing.T) {
	cfg := Config{
		Enabled:  true,
		Provider: "prometheus",
		Prometheus: PrometheusConfig{
			Path: "/metrics",
			Port: 9090,
		},
	}

	s := cfg.Sanitize()
	if s == &cfg {
		t.Fatalf("Sanitize must return a copy")
	}

	sum := cfg.ConfigSummary()
	if sum["provider"] != "prometheus" {
		t.Fatalf("unexpected provider in summary")
	}
}

func TestConfigStructure(t *testing.T) {
	t.Run("config has correct prefix", func(t *testing.T) {
		cfg := Config{}
		assert.Equal(t, "metrics", cfg.Prefix())
	})

	t.Run("creates config with values", func(t *testing.T) {
		config := Config{
			Enabled:  true,
			Provider: "prometheus",
			Prometheus: PrometheusConfig{
				Port: 9090,
				Path: "/metrics",
			},
		}

		assert.True(t, config.Enabled)
		assert.Equal(t, "prometheus", config.Provider)
		assert.Equal(t, 9090, config.Prometheus.Port)
		assert.Equal(t, "/metrics", config.Prometheus.Path)
	})

	t.Run("prometheus config structure", func(t *testing.T) {
		config := PrometheusConfig{
			Namespace:            "myapp",
			Subsystem:            "http",
			Port:                 8080,
			Path:                 "/custom-metrics",
			EnableProcessMetrics: true,
			EnableGoMetrics:      false,
		}

		assert.Equal(t, "myapp", config.Namespace)
		assert.Equal(t, "http", config.Subsystem)
		assert.Equal(t, 8080, config.Port)
		assert.Equal(t, "/custom-metrics", config.Path)
		assert.True(t, config.EnableProcessMetrics)
		assert.False(t, config.EnableGoMetrics)
	})
}

func TestPrometheusConfigValidation(t *testing.T) {
	t.Run("valid prometheus config", func(t *testing.T) {
		config := PrometheusConfig{
			Port: 9090,
			Path: "/metrics",
		}

		// Should not panic or error when creating provider
		assert.NotPanics(t, func() {
			newPrometheusProvider(config, getTestLogger())
		})
	})

	t.Run("prometheus config with custom settings", func(t *testing.T) {
		config := PrometheusConfig{
			Port:                 8080,
			Path:                 "/custom",
			Namespace:            "app",
			Subsystem:            "api",
			EnableProcessMetrics: false,
			EnableGoMetrics:      false,
		}

		provider := newPrometheusProvider(config, getTestLogger())
		assert.NotNil(t, provider)
	})

	t.Run("prometheus config with port 0", func(t *testing.T) {
		config := PrometheusConfig{
			Port: 0, // Metrics on main server
			Path: "/metrics",
		}

		provider := newPrometheusProvider(config, getTestLogger())
		assert.NotNil(t, provider)
	})
}
