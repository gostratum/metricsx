package metricsx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNoopProvider(t *testing.T) {
	t.Run("creates noop provider", func(t *testing.T) {
		provider := newNoopProvider()
		assert.NotNil(t, provider)
	})

	t.Run("noop counter", func(t *testing.T) {
		provider := newNoopProvider()
		counter := provider.Counter("test_counter", &Options{})

		// Should not panic
		counter.Inc()
		counter.Inc("label1")
		counter.Add(5.0)
		counter.Add(10.0, "label1", "label2")
	})

	t.Run("noop gauge", func(t *testing.T) {
		provider := newNoopProvider()
		gauge := provider.Gauge("test_gauge", &Options{})

		// Should not panic
		gauge.Set(42.0)
		gauge.Set(100.0, "label1")
		gauge.Inc()
		gauge.Inc("label1")
		gauge.Dec()
		gauge.Dec("label1")
		gauge.Add(5.0)
		gauge.Add(10.0, "label1")
		gauge.Sub(3.0)
		gauge.Sub(7.0, "label1")
	})

	t.Run("noop histogram", func(t *testing.T) {
		provider := newNoopProvider()
		histogram := provider.Histogram("test_histogram", &Options{})

		// Should not panic
		histogram.Observe(0.5)
		histogram.Observe(1.5, "label1")

		timer := histogram.Timer()
		assert.NotNil(t, timer)

		timerWithLabels := histogram.Timer("label1", "label2")
		assert.NotNil(t, timerWithLabels)
	})

	t.Run("noop summary", func(t *testing.T) {
		provider := newNoopProvider()
		summary := provider.Summary("test_summary", &Options{})

		// Should not panic
		summary.Observe(0.5)
		summary.Observe(1.5, "label1")
	})

	t.Run("noop timer", func(t *testing.T) {
		provider := newNoopProvider()
		histogram := provider.Histogram("test", &Options{})
		timer := histogram.Timer()

		time.Sleep(10 * time.Millisecond)

		// Should not panic
		timer.ObserveDuration()

		duration := timer.Stop()
		assert.Greater(t, duration, time.Duration(0))
	})

	t.Run("noop lifecycle", func(t *testing.T) {
		provider := newNoopProvider()
		ctx := context.Background()

		// Should not error
		err := provider.Start(ctx)
		assert.NoError(t, err)

		err = provider.Stop(ctx)
		assert.NoError(t, err)
	})
}
