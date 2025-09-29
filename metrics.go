package obs

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	SwipeTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "swipe_actions_total",
			Help: "Total des actions swipe par type",
		},
		[]string{"action"},
	)
	SwipeErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "swipe_errors_total",
			Help: "Total des erreurs lors des swipes",
		},
	)
	MatchesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "matches_total",
			Help: "Total des matches obtenus",
		},
	)
	RequestLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "request_latency_seconds",
			Help:    "Latence des requêtes vers l'API",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func init() {
	prometheus.MustRegister(SwipeTotal, SwipeErrors, MatchesTotal, RequestLatency)
}

func Instrument(next func() error) error {
	start := time.Now()
	err := next()
	RequestLatency.Observe(time.Since(start).Seconds())
	return err
}

func Handler() http.Handler {
	return promhttp.Handler()
}
