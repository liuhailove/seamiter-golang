package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	registry *prometheus.Registry

	httpHandler http.Handler
)

func init() {
	registry = prometheus.NewRegistry()

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	httpHandler = promhttp.InstrumentMetricHandler(registry, h)
}

type Counter struct {
	cv *prometheus.CounterVec
}

type Gauge struct {
	gv *prometheus.GaugeVec
}

type Histogram struct {
	hv *prometheus.HistogramVec
}

func NewCounter(name, namespace, desc string, labelNames []string, constLabels map[string]string) *Counter {
	return &Counter{cv: prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        name,
			Namespace:   namespace,
			Help:        desc,
			ConstLabels: constLabels,
		}, labelNames)}
}
func (c *Counter) Add(value float64, labelValues ...string) {
	c.cv.WithLabelValues(labelValues...).Add(value)
}

func (c *Counter) Register() error {
	return registry.Register(c.cv)
}

func (c *Counter) Unregister() bool {
	return registry.Unregister(c.cv)
}

func (c *Counter) Reset() {
	c.cv.Reset()
}

func NewGauge(name, namespace, desc string, labelNames []string, constLabels map[string]string) *Gauge {
	return &Gauge{
		gv: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:        name,
			Namespace:   namespace,
			Help:        desc,
			ConstLabels: constLabels,
		}, labelNames),
	}
}

func (g *Gauge) Set(value float64, labelValues ...string) {
	g.gv.WithLabelValues(labelValues...).Set(value)
}

func (g *Gauge) Register() error {
	return registry.Register(g.gv)
}

func (g *Gauge) Unregister() bool {
	return registry.Unregister(g.gv)
}

func (g *Gauge) Reset() {
	g.gv.Reset()
}

func NewHistogram(name, namespace, desc string, buckets []float64, labelNames []string, constLabels map[string]string) *Histogram {
	return &Histogram{
		hv: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   namespace,
			Name:        name,
			Help:        desc,
			ConstLabels: constLabels,
		}, labelNames),
	}
}

func (h *Histogram) Observe(value float64, labelValues ...string) {
	h.hv.WithLabelValues(labelValues...).Observe(value)
}

func (h *Histogram) Register() error {
	return registry.Register(h.hv)
}

func (h *Histogram) Unregister() bool {
	return registry.Unregister(h.hv)
}

func (h *Histogram) Reset() {
	h.hv.Reset()
}

func HTTPHandler() http.Handler {
	return httpHandler
}
