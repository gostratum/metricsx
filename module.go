package metricsx

import (
	"context"

	"github.com/gostratum/core/logx"
	"go.uber.org/fx"
)

// Params contains dependencies for the metrics module
type Params struct {
	fx.In
	Config Config
	Logger logx.Logger
}

// Result contains outputs from the metrics module
type Result struct {
	fx.Out
	Metrics  Metrics
	Provider Provider
}

// Module provides the metrics module for fx
func Module() fx.Option {
	return fx.Module("metrics",
		fx.Provide(
			NewConfig,
			NewMetrics,
		),
		fx.Invoke(registerLifecycle),
	)
}

// NewMetrics creates a new Metrics instance based on configuration
func NewMetrics(p Params) (Result, error) {
	var provider Provider

	switch p.Config.Provider {
	case "prometheus":
		provider = newPrometheusProvider(p.Config.Prometheus, p.Logger)
	case "noop":
		provider = newNoopProvider()
	default:
		p.Logger.Warn("unknown metrics provider, using noop", logx.String("provider", p.Config.Provider))
		provider = newNoopProvider()
	}

	metrics := &metricsImpl{
		provider: provider,
		logger:   p.Logger,
	}

	return Result{
		Metrics:  metrics,
		Provider: provider,
	}, nil
}

// registerLifecycle registers the metrics lifecycle hooks
func registerLifecycle(lc fx.Lifecycle, provider Provider, logger logx.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting metrics provider")
			return provider.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping metrics provider")
			return provider.Stop(ctx)
		},
	})
}

// metricsImpl implements the Metrics interface
type metricsImpl struct {
	provider Provider
	logger   logx.Logger
}

func (m *metricsImpl) Counter(name string, opts ...Option) Counter {
	options := applyOptions(opts...)
	return m.provider.Counter(name, options)
}

func (m *metricsImpl) Gauge(name string, opts ...Option) Gauge {
	options := applyOptions(opts...)
	return m.provider.Gauge(name, options)
}

func (m *metricsImpl) Histogram(name string, opts ...Option) Histogram {
	options := applyOptions(opts...)
	return m.provider.Histogram(name, options)
}

func (m *metricsImpl) Summary(name string, opts ...Option) Summary {
	options := applyOptions(opts...)
	return m.provider.Summary(name, options)
}
