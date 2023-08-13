package hotspot

import "git.garena.com/honggang.liu/seamiter-go/core/hotspot/cache"

const (
	ConcurrencyMaxCount = 4000
	ParamsCapacityBase  = 4000
	ParamsMaxCapacity   = 20000
)

// ParamsMetric carries real-time counters for frequent ("hot spot") parameters.
//
// For each cache map, the key is the parameter value, while the value is the counter.
type ParamsMetric struct {
	// RuleTimeCounter records the last added token timestamp.
	RuleTimeCounter cache.ConcurrentCounterCache
	// RuleTokenCounter records the number of tokens.
	RuleTokenCounter cache.ConcurrentCounterCache
	// ConcurrencyCounter records the real-time concurrency.
	ConcurrentCounter cache.ConcurrentCounterCache
}
