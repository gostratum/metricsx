package metricsx

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gostratum/core/logx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// prometheusProvider implements the Provider interface for Prometheus
type prometheusProvider struct {
	config   PrometheusConfig
	logger   logx.Logger
	registry *prometheus.Registry
	server   *http.Server

	mu         sync.RWMutex
	counters   map[string]*prometheusCounterVec
	gauges     map[string]*prometheusGaugeVec
	histograms map[string]*prometheusHistogramVec
	summaries  map[string]*prometheusSummaryVec
}

// newPrometheusProvider creates a new Prometheus provider
func newPrometheusProvider(config PrometheusConfig, logger logx.Logger) Provider {
	registry := prometheus.NewRegistry()

	// Register default collectors if enabled
	if config.EnableProcessMetrics {
		registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	}
	if config.EnableGoMetrics {
		registry.MustRegister(prometheus.NewGoCollector())
	}

	return &prometheusProvider{
		config:     config,
		logger:     logger,
		registry:   registry,
		counters:   make(map[string]*prometheusCounterVec),
		gauges:     make(map[string]*prometheusGaugeVec),
		histograms: make(map[string]*prometheusHistogramVec),
		summaries:  make(map[string]*prometheusSummaryVec),
	}
}

// Counter creates or retrieves a counter metric
func (p *prometheusProvider) Counter(name string, options *Options) Counter {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := p.metricKey(name, options)
	if c, exists := p.counters[key]; exists {
		return c
	}

	counterVec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: p.namespace(options),
			Subsystem: p.subsystem(options),
			Name:      name,
			Help:      options.Help,
		},
		options.Labels,
	)

	p.registry.MustRegister(counterVec)

	counter := &prometheusCounterVec{
		vec:    counterVec,
		labels: options.Labels,
	}

	p.counters[key] = counter
	return counter
}

// Gauge creates or retrieves a gauge metric
func (p *prometheusProvider) Gauge(name string, options *Options) Gauge {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := p.metricKey(name, options)
	if g, exists := p.gauges[key]; exists {
		return g
	}

	gaugeVec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: p.namespace(options),
			Subsystem: p.subsystem(options),
			Name:      name,
			Help:      options.Help,
		},
		options.Labels,
	)

	p.registry.MustRegister(gaugeVec)

	gauge := &prometheusGaugeVec{
		vec:    gaugeVec,
		labels: options.Labels,
	}

	p.gauges[key] = gauge
	return gauge
}

// Histogram creates or retrieves a histogram metric
func (p *prometheusProvider) Histogram(name string, options *Options) Histogram {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := p.metricKey(name, options)
	if h, exists := p.histograms[key]; exists {
		return h
	}

	histogramVec := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: p.namespace(options),
			Subsystem: p.subsystem(options),
			Name:      name,
			Help:      options.Help,
			Buckets:   options.Buckets,
		},
		options.Labels,
	)

	p.registry.MustRegister(histogramVec)

	histogram := &prometheusHistogramVec{
		vec:    histogramVec,
		labels: options.Labels,
	}

	p.histograms[key] = histogram
	return histogram
}

// Summary creates or retrieves a summary metric
func (p *prometheusProvider) Summary(name string, options *Options) Summary {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := p.metricKey(name, options)
	if s, exists := p.summaries[key]; exists {
		return s
	}

	summaryVec := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  p.namespace(options),
			Subsystem:  p.subsystem(options),
			Name:       name,
			Help:       options.Help,
			Objectives: options.Objectives,
		},
		options.Labels,
	)

	p.registry.MustRegister(summaryVec)

	summary := &prometheusSummaryVec{
		vec:    summaryVec,
		labels: options.Labels,
	}

	p.summaries[key] = summary
	return summary
}

// Start starts the Prometheus HTTP server if a port is configured
func (p *prometheusProvider) Start(ctx context.Context) error {
	if p.config.Port == 0 {
		p.logger.Info("metrics will be exposed on main HTTP server", logx.String("path", p.config.Path))
		return nil
	}

	addr := fmt.Sprintf(":%d", p.config.Port)
	p.logger.Info("starting metrics HTTP server", logx.String("addr", addr), logx.String("path", p.config.Path))

	mux := http.NewServeMux()
	mux.Handle(p.config.Path, promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{}))

	p.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			p.logger.Error("metrics HTTP server error", logx.Err(err))
		}
	}()

	return nil
}

// Stop stops the Prometheus HTTP server
func (p *prometheusProvider) Stop(ctx context.Context) error {
	if p.server == nil {
		return nil
	}

	p.logger.Info("stopping metrics HTTP server")
	return p.server.Shutdown(ctx)
}

// Handler returns the HTTP handler for metrics
func (p *prometheusProvider) Handler() http.Handler {
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})
}

// metricKey generates a unique key for a metric
func (p *prometheusProvider) metricKey(name string, options *Options) string {
	return fmt.Sprintf("%s_%s_%s", p.namespace(options), p.subsystem(options), name)
}

// namespace returns the namespace to use for metrics
func (p *prometheusProvider) namespace(options *Options) string {
	if options.Namespace != "" {
		return options.Namespace
	}
	return p.config.Namespace
}

// subsystem returns the subsystem to use for metrics
func (p *prometheusProvider) subsystem(options *Options) string {
	if options.Subsystem != "" {
		return options.Subsystem
	}
	return p.config.Subsystem
}

// prometheusCounterVec implements Counter
type prometheusCounterVec struct {
	vec    *prometheus.CounterVec
	labels []string
}

func (c *prometheusCounterVec) Inc(labels ...string) {
	c.vec.WithLabelValues(labels...).Inc()
}

func (c *prometheusCounterVec) Add(value float64, labels ...string) {
	c.vec.WithLabelValues(labels...).Add(value)
}

// prometheusGaugeVec implements Gauge
type prometheusGaugeVec struct {
	vec    *prometheus.GaugeVec
	labels []string
}

func (g *prometheusGaugeVec) Set(value float64, labels ...string) {
	g.vec.WithLabelValues(labels...).Set(value)
}

func (g *prometheusGaugeVec) Inc(labels ...string) {
	g.vec.WithLabelValues(labels...).Inc()
}

func (g *prometheusGaugeVec) Dec(labels ...string) {
	g.vec.WithLabelValues(labels...).Dec()
}

func (g *prometheusGaugeVec) Add(value float64, labels ...string) {
	g.vec.WithLabelValues(labels...).Add(value)
}

func (g *prometheusGaugeVec) Sub(value float64, labels ...string) {
	g.vec.WithLabelValues(labels...).Sub(value)
}

// prometheusHistogramVec implements Histogram
type prometheusHistogramVec struct {
	vec    *prometheus.HistogramVec
	labels []string
}

func (h *prometheusHistogramVec) Observe(value float64, labels ...string) {
	h.vec.WithLabelValues(labels...).Observe(value)
}

func (h *prometheusHistogramVec) Timer(labels ...string) Timer {
	return &prometheusTimer{
		histogram: h,
		labels:    labels,
		start:     time.Now(),
	}
}

// prometheusSummaryVec implements Summary
type prometheusSummaryVec struct {
	vec    *prometheus.SummaryVec
	labels []string
}

func (s *prometheusSummaryVec) Observe(value float64, labels ...string) {
	s.vec.WithLabelValues(labels...).Observe(value)
}

// prometheusTimer implements Timer
type prometheusTimer struct {
	histogram *prometheusHistogramVec
	labels    []string
	start     time.Time
}

func (t *prometheusTimer) ObserveDuration() {
	duration := time.Since(t.start).Seconds()
	t.histogram.Observe(duration, t.labels...)
}

func (t *prometheusTimer) Stop() time.Duration {
	duration := time.Since(t.start)
	t.histogram.Observe(duration.Seconds(), t.labels...)
	return duration
}
