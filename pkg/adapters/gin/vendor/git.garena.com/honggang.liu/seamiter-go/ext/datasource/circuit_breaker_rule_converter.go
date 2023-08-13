package datasource

import cb "git.garena.com/honggang.liu/seamiter-go/core/circuitbreaker"

// CircuitBreakerRule encompasses the fields of circuit breaking rule.
type CircuitBreakerRule struct {
	// unique id
	Id string `json:"id,omitempty"`
	// resource name
	Resource string      `json:"resource"`
	Strategy cb.Strategy `json:"strategy"`
	// RetryTimeoutMs represents recovery timeout (in milliseconds) before the circuit breaker opens.
	// During the open period, no requests are permitted until the timeout has elapsed.
	// After that, the circuit breaker will transform to half-open state for trying a few "trial" requests.
	RetryTimeoutMs uint32 `json:"retryTimeoutMs"`
	// MinRequestAmount represents the minimum number of requests (in an active statistic time span)
	// that can trigger circuit breaking.
	MinRequestAmount uint64 `json:"minRequestAmount"`
	// StatIntervalMs represents statistic time interval of the internal circuit breaker (in ms).
	// Currently, the statistic interval is collected by sliding window.
	StatIntervalMs uint32 `json:"statIntervalMs"`
	// StatSlidingWindowBucketCount represents the bucket count of statistic sliding window.
	// The statistic will be more precise as the bucket count increases, but the memory cost increases too.
	// The following must be true — “StatIntervalMs % StatSlidingWindowBucketCount == 0”,
	// otherwise StatSlidingWindowBucketCount will be replaced by 1.
	// If it is not set, default value 1 will be used.
	StatSlidingWindowBucketCount uint32 `json:"statSlidingWindowBucketCount"`
	// MaxAllowedRtMs indicates that any invocation whose response time exceeds this value (in ms)
	// will be recorded as a slow request.
	// MaxAllowedRtMs only takes effect for SlowRequestRatio strategy
	MaxAllowedRtMs uint64 `json:"maxAllowedRtMs"`
	// Threshold represents the threshold of circuit breaker.
	// for SlowRequestRatio, it represents the max slow request ratio
	// for ErrorRatio, it represents the max error request ratio
	// for ErrorCount, it represents the max error request count
	SlowRatioThreshold float64 `json:"slowRatioThreshold"`
	// Threshold represents the threshold of circuit breaker.
	// for SlowRequestRatio, it represents the max slow request ratio
	// for ErrorRatio, it represents the max error request ratio
	// for ErrorCount, it represents the max error request count
	Threshold float64 `json:"threshold"`
	//ProbeNum is number of probes required when the circuit breaker is half-open.
	//when the probe num are set  and circuit breaker in the half-open state.
	//if err occurs during the probe, the circuit breaker is opened immediately.
	//otherwise,the circuit breaker is closed only after the number of probes is reached
	ProbeNum uint64 `json:"probeNum"`
}

func transToCircuitBreakerRule(source []cb.Rule) []CircuitBreakerRule {
	degradeRules := make([]CircuitBreakerRule, 0, 8)
	if source == nil || len(source) == 0 {
		return degradeRules
	}
	for _, s := range source {
		dg := CircuitBreakerRule{
			Id:                           s.Id,
			Resource:                     s.Resource,
			Strategy:                     s.Strategy,
			RetryTimeoutMs:               s.RetryTimeoutMs,
			MinRequestAmount:             s.MinRequestAmount,
			StatIntervalMs:               s.StatIntervalMs,
			StatSlidingWindowBucketCount: s.StatSlidingWindowBucketCount,
			MaxAllowedRtMs:               s.MaxAllowedRtMs,
			Threshold:                    s.Threshold,
		}
		if s.Strategy == cb.SlowRequestRatio {
			dg.Threshold = float64(s.MaxAllowedRtMs)
			dg.SlowRatioThreshold = s.Threshold
		}
		degradeRules = append(degradeRules, dg)
	}
	return degradeRules
}
