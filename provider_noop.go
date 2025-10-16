package metricsx

import (
	"context"
	"time"
)

// noopProvider implements a no-op metrics provider for testing
type noopProvider struct{}

// newNoopProvider creates a new no-op provider
func newNoopProvider() Provider {
	return &noopProvider{}
}

func (p *noopProvider) Counter(name string, options *Options) Counter {
	return &noopCounter{}
}

func (p *noopProvider) Gauge(name string, options *Options) Gauge {
	return &noopGauge{}
}

func (p *noopProvider) Histogram(name string, options *Options) Histogram {
	return &noopHistogram{}
}

func (p *noopProvider) Summary(name string, options *Options) Summary {
	return &noopSummary{}
}

func (p *noopProvider) Start(ctx context.Context) error {
	return nil
}

func (p *noopProvider) Stop(ctx context.Context) error {
	return nil
}

type noopCounter struct{}

func (c *noopCounter) Inc(labels ...string)                {}
func (c *noopCounter) Add(value float64, labels ...string) {}

type noopGauge struct{}

func (g *noopGauge) Set(value float64, labels ...string) {}
func (g *noopGauge) Inc(labels ...string)                {}
func (g *noopGauge) Dec(labels ...string)                {}
func (g *noopGauge) Add(value float64, labels ...string) {}
func (g *noopGauge) Sub(value float64, labels ...string) {}

type noopHistogram struct{}

func (h *noopHistogram) Observe(value float64, labels ...string) {}
func (h *noopHistogram) Timer(labels ...string) Timer {
	return &noopTimer{}
}

type noopSummary struct{}

func (s *noopSummary) Observe(value float64, labels ...string) {}

type noopTimer struct {
	start time.Time
}

func (t *noopTimer) ObserveDuration()    {}
func (t *noopTimer) Stop() time.Duration { return 0 }
