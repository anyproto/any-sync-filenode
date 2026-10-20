package index

import (
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
)

type bloomStat struct {
	// BF.EXISTS returned true for a key that is not in redis
	positive atomic.Uint64
	// the persistent store had no value for a positive key: a wasted IndexGet
	falsePositive atomic.Uint64
}

func registerMetric(s *bloomStat, registry *prometheus.Registry) {
	registry.MustRegister(prometheus.NewCounterFunc(prometheus.CounterOpts{
		Namespace: "filenode",
		Subsystem: "index",
		Name:      "bloom_positive",
		Help:      "keys missing in redis that the bloom filter reported as present",
	}, func() float64 {
		return float64(s.positive.Load())
	}))
	registry.MustRegister(prometheus.NewCounterFunc(prometheus.CounterOpts{
		Namespace: "filenode",
		Subsystem: "index",
		Name:      "bloom_false_positive",
		Help:      "bloom filter positives not found in the persistent store",
	}, func() float64 {
		return float64(s.falsePositive.Load())
	}))
}
