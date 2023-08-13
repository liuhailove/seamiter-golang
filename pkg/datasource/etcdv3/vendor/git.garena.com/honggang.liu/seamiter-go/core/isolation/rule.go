package isolation

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
)

// MetricType represents the target metric type.
type MetricType int32

const (
	// Concurrency represents concurrency (in-flight requests).
	Concurrency MetricType = iota
)

func (s MetricType) String() string {
	switch s {
	case Concurrency:
		return "Concurrency"
	default:
		return "Undefined"
	}
}

// Rule describes the isolation policy (e.g. semaphore isolation).
type Rule struct {
	// ID represents the unique ID of the rule (optional).
	ID string `json:"id,omitempty"`
	// Resource represents the target resource definition.
	Resource string `json:"resource"`
	// MetricType indicates the metric type for checking logic.
	// Currently, Concurrency is supported for concurrency limiting.
	MetricType MetricType `json:"metricType"`
	Threshold  uint32     `json:"threshold"`
}

func (r *Rule) String() string {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b, err := json.Marshal(r)
	if err != nil {
		// Return the fallback string
		return fmt.Sprintf("{Id=%s, Resource=%s, MetricType=%s, Count=%d}", r.ID, r.Resource, r.MetricType.String(), r.Threshold)
	}
	return string(b)
}

func (r *Rule) ResourceName() string {
	return r.Resource
}
