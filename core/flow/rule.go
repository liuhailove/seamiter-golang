package flow

import (
	"fmt"
	"github.com/liuhailove/seamiter-golang/util"
	jsoniter "github.com/json-iterator/go"
)

// RelationStrategy indicates the flow control strategy based on the relation of invocations.
type RelationStrategy int32

const (
	// CurrentResource means flow control by current resource directly.
	CurrentResource RelationStrategy = iota
	// AssociatedResource means flow control by the associated resource rather than current resource.
	AssociatedResource
)

func (s RelationStrategy) String() string {
	switch s {
	case CurrentResource:
		return "CurrentResource"
	case AssociatedResource:
		return "AssociatedResource"
	default:
		return "Undefined"
	}
}

type TokenCalculateStrategy int32

const (
	Direct TokenCalculateStrategy = iota
	WarmUp
	MemoryAdaptive
)

func (s TokenCalculateStrategy) String() string {
	switch s {
	case Direct:
		return "Direct"
	case WarmUp:
		return "WarmUp"
	case MemoryAdaptive:
		return "MemoryAdaptive"
	default:
		return "Undefined"
	}
}

// ControlBehavior defines the behavior when requests have reached the capacity of the resource.
type ControlBehavior int32

const (
	Reject ControlBehavior = iota
	// Throttling indicates that pending requests will be throttled, wait in queue (until free capacity is available)
	Throttling
)

func (s ControlBehavior) String() string {
	switch s {
	case Reject:
		return "Reject"
	case Throttling:
		return "Throttling"
	default:
		return "Undefined"
	}
}

// Rule describes the strategy of flow control, the flow control strategy is based on QPS statistic metric
type Rule struct {
	// ID represents the unique ID of the rule (optional).
	ID string `json:"id,omitempty"`

	// LimitApp 限制应用程序
	// 将受来源限制的应用程序名称。
	// 默认的limitApp是{@code default}，表示允许所有源端应用。
	// 对于权限规则，多个源名称可以用逗号（','）分隔。
	LimitApp string `json:"limitApp"`
	// Resource 资源名称
	Resource               string                 `json:"resource"`
	TokenCalculateStrategy TokenCalculateStrategy `json:"tokenCalculateStrategy"`
	ControlBehavior        ControlBehavior        `json:"controlBehavior"`
	// Threshold means the threshold during StatIntervalInMs
	// If StatIntervalInMs is 1000(1 second), Threshold means QPS
	Threshold        float64          `json:"threshold"`
	RelationStrategy RelationStrategy `json:"relationStrategy"`
	RefResource      string           `json:"refResource"`
	// MaxQueueingTimeMs only takes effect when ControlBehavior is Throttling.
	// When MaxQueueingTimeMs is 0, it means Throttling only controls interval of requests,
	// and requests exceeding the threshold will be rejected directly.
	MaxQueueingTimeMs uint32 `json:"maxQueueingTimeMs"`
	// 预热时间
	WarmUpPeriodSec uint32 `json:"warmUpPeriodSec"`
	// 预热期内的令牌生产减缓因子，固定值3
	WarmUpColdFactor uint32 `json:"warmUpColdFactor"`
	// StatIntervalInMs indicates the statistic interval and it's the optional setting for flow Rule.
	// If user doesn't set StatIntervalInMs, that means using default metric statistic of resource.
	// If the StatIntervalInMs user specifies can not reuse the global statistic of resource,
	// sea will generate independent statistic structure for this rule.
	StatIntervalInMs uint32 `json:"statIntervalInMs"`

	// adaptive flow control algorithm related parameters'
	// limitation: LowMemUsageThreshold > HighMemUsageThreshold && MemHighWaterMarkBytes > MemLowWaterMarkBytes
	// if the current memory usage is less than or equals to MemLowWaterMarkBytes, threshold == LowMemUsageThreshold
	// if the current memory usage is more than or equals to MemHighWaterMarkBytes, threshold == HighMemUsageThreshold
	// if  the current memory usage is in (MemLowWaterMarkBytes, MemHighWaterMarkBytes), threshold is in (HighMemUsageThreshold, LowMemUsageThreshold)
	LowMemUsageThreshold  int64 `json:"lowMemUsageThreshold"`
	HighMemUsageThreshold int64 `json:"highMemUsageThreshold"`
	MemLowWaterMarkBytes  int64 `json:"memLowWaterMarkBytes"`
	MemHighWaterMarkBytes int64 `json:"memHighWaterMarkBytes"`
}

func (r *Rule) isEqualsTo(newRule *Rule) bool {
	if newRule == nil {
		return false
	}
	if !(r.Resource == newRule.Resource &&
		r.RelationStrategy == newRule.RelationStrategy &&
		r.RefResource == newRule.RefResource &&
		r.StatIntervalInMs == newRule.StatIntervalInMs &&
		r.TokenCalculateStrategy == newRule.TokenCalculateStrategy &&
		r.ControlBehavior == newRule.ControlBehavior &&
		util.Float64Equals(r.Threshold, newRule.Threshold) &&
		r.MaxQueueingTimeMs == newRule.MaxQueueingTimeMs &&
		r.WarmUpPeriodSec == newRule.WarmUpPeriodSec &&
		r.WarmUpColdFactor == newRule.WarmUpColdFactor &&
		r.LowMemUsageThreshold == newRule.LowMemUsageThreshold &&
		r.HighMemUsageThreshold == newRule.HighMemUsageThreshold &&
		r.MemLowWaterMarkBytes == newRule.MemLowWaterMarkBytes &&
		r.MemHighWaterMarkBytes == newRule.MemHighWaterMarkBytes &&
		r.LimitApp == newRule.LimitApp) {
		return false
	}
	return true
}

func (r *Rule) isStatReusable(newRule *Rule) bool {
	if newRule == nil {
		return false
	}
	return r.Resource == newRule.Resource &&
		r.RelationStrategy == newRule.RelationStrategy &&
		r.RefResource == newRule.RefResource &&
		r.StatIntervalInMs == newRule.StatIntervalInMs &&
		r.needStatistic() && newRule.needStatistic()
}

func (r *Rule) needStatistic() bool {
	return r.TokenCalculateStrategy == WarmUp || r.ControlBehavior == Reject
}

func (r *Rule) String() string {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b, err := json.Marshal(r)
	if err != nil {
		// Return the fallback string
		return fmt.Sprintf("Rule{Resource=%s, TokenCalculateStrategy=%s, ControlBehavior=%s, "+
			"Count=%.2f, RelationStrategy=%s, RefResource=%s, MaxQueueingTimeMs=%d, WarmUpPeriodSec=%d, WarmUpColdFactor=%d, StatIntervalInMs=%d, "+
			"LowMemUsageThreshold=%v, HighMemUsageThreshold=%v, MemLowWaterMarkBytes=%v, MemHighWaterMarkBytes=%v}",
			r.Resource, r.TokenCalculateStrategy, r.ControlBehavior, r.Threshold, r.RelationStrategy, r.RefResource,
			r.MaxQueueingTimeMs, r.WarmUpPeriodSec, r.WarmUpColdFactor, r.StatIntervalInMs,
			r.LowMemUsageThreshold, r.HighMemUsageThreshold, r.MemLowWaterMarkBytes, r.MemHighWaterMarkBytes)
	}
	return string(b)
}

func (r *Rule) ResourceName() string {
	return r.Resource
}
