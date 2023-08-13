package isolation

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
)

// MetricType represents the target metric type.
type MetricType int32

const (
	UNKNOWN MetricType = iota
	// Concurrency represents concurrency (in-flight requests).
	Concurrency
)

func (s MetricType) String() string {
	switch s {
	case Concurrency:
		return "Concurrency"
	default:
		return "Undefined"
	}
}

// Rule 描述隔离策略（例如信号量隔离）
type Rule struct {
	// ID 规则唯一ID（可选）
	ID string `json:"id,omitempty"`
	// Resource 目标资源
	Resource string `json:"resource"`
	// MetricType 检查逻辑的metric类型。
	// 目前支持Concurrency进行并发限制
	MetricType MetricType `json:"metricType"`
	// Threshold 阈值
	Threshold uint32 `json:"threshold"`
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
