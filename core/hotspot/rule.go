package hotspot

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"reflect"
	"strconv"
)

// ControlBehavior indicates the traffic shaping behaviour.
type ControlBehavior int32

const (
	Reject ControlBehavior = iota
	Throttling
)

func (t ControlBehavior) String() string {
	switch t {
	case Reject:
		return "Reject"
	case Throttling:
		return "Throttling"
	default:
		return strconv.Itoa(int(t))
	}
}

// MetricType 表示目标度量类型。
type MetricType int32

const (
	// Concurrency 标识并发统计
	Concurrency MetricType = iota
	// QPS 标识每秒的统计数
	QPS
)

func (t MetricType) String() string {
	switch t {
	case Concurrency:
		return "Concurrency"
	case QPS:
		return "QPS"
	default:
		return "Undefined"
	}
}

// Rule 代表热点（频繁）参数流控规则
type Rule struct {
	// ID 唯一ID
	ID string `json:"id,omitempty"`
	// LimitApp 限制应用程序
	// 将受来源限制的应用程序名称。
	// 默认的limitApp是{@code default}，表示允许所有源端应用。
	// 对于权限规则，多个源名称可以用逗号（','）分隔。
	LimitApp string `json:"limitApp"`
	// Resource 资源名称
	Resource string `json:"resource"`
	// MetricType 表示检查逻辑的metric类型。
	// 对于 Concurrency 指标，热点模块将检查每个热点参数的并发度，
	//     如果并发超过Threshold，则直接拒绝流量。
	// 对于 QPS 指标，热点模块会检查每个热点参数的QPS，
	//     ControlBehavior 决定流量整形控制器的行为
	MetricType MetricType `json:"metricType"`
	// ControlBehavior 标识流量整形行为。
	// 仅仅当MetricType是QPS时才会生效
	ControlBehavior ControlBehavior `json:"controlBehavior"`
	// ParamIdx 是上下文参数切片中的索引。
	// 如果 ParamIdx 大于或等于 0，ParamIdx 表示第 <ParamIdx> 参数
	// 如果ParamIdx为负数，则ParamIdx表示反转的第<ParamIdx>参数
	ParamIdx int `json:"paramIdx"`
	// ParamKey 是 EntryContext.Input.Attachments 映射中的键。
	// ParamKey可以作为ParamIdx的补充，方便规则从大量参数中快速获取参数
	// ParamKey与ParamIdx互斥，ParamKey比ParamIdx优先级高
	ParamKey string `json:"paramKey"`
	// Threshold是触发拒绝的阈值
	Threshold float64 `json:"threshold"`
	// MaxQueueingTimeMs 仅在ControlBehavior为Throttling且MetricType为QPS时生效
	MaxQueueingTimeMs int64 `json:"maxQueueingTimeMs"`
	// BurstCount 是静默计数
	// BurstCount 仅在ControlBehavior为Reject且MetricType为QPS时生效
	BurstCount int64 `json:"burstCount"`
	// DurationInSec 为统计的时间间隔
	// DurationInSec 仅在MetricType为QPS时生效
	DurationInSec int64 `json:"durationInSec"`
	// ParamsMaxCapacity 是缓存统计的最大容量
	ParamsMaxCapacity int64 `json:"ParamsMaxCapacity"`
	// SpecificItems 表示特定值的特殊阈值
	SpecificItems map[interface{}]int64 `json:"specificItems"`
}

func (r *Rule) String() string {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b, err := json.Marshal(r)
	if err != nil {
		// Return the fallback string
		return fmt.Sprintf("{Id:%s, Resource:%s, MetricType:%+v, ControlBehavior:%+v, ParamIdx:%d, ParamKey:%s, Count:%d, MaxQueueingTimeMs:%d, BurstCount:%d, DurationInSec:%d, ParamsMaxCapacity:%d, ParamFlowItems:%+v}",
			r.ID, r.Resource, r.MetricType, r.ControlBehavior, r.ParamIdx, r.ParamKey, r.Threshold, r.MaxQueueingTimeMs, r.BurstCount, r.DurationInSec, r.ParamsMaxCapacity, r.SpecificItems)
	}
	return string(b)
}

func (r *Rule) ResourceName() string {
	return r.Resource
}

// IsStatReusable checks whether current rule is "statistically" equal to the given rule.
func (r *Rule) IsStatReusable(newRule *Rule) bool {
	return r.Resource == newRule.Resource && r.ControlBehavior == newRule.ControlBehavior &&
		r.ParamsMaxCapacity == newRule.ParamsMaxCapacity && r.DurationInSec == newRule.DurationInSec &&
		r.MetricType == newRule.MetricType
}

// Equals checks whether current rule is consistent with the given rule.
func (r *Rule) Equals(newRule *Rule) bool {
	baseCheck := r.Resource == newRule.Resource && r.MetricType == newRule.MetricType && r.ControlBehavior == newRule.ControlBehavior && r.ParamsMaxCapacity == newRule.ParamsMaxCapacity && r.ParamIdx == newRule.ParamIdx && r.ParamKey == newRule.ParamKey && r.Threshold == newRule.Threshold && r.DurationInSec == newRule.DurationInSec && r.LimitApp == newRule.LimitApp && reflect.DeepEqual(r.SpecificItems, newRule.SpecificItems)
	if !baseCheck {
		return false
	}
	if r.ControlBehavior == Reject {
		return r.BurstCount == newRule.BurstCount
	}
	if r.ControlBehavior == Throttling {
		return r.MaxQueueingTimeMs == newRule.MaxQueueingTimeMs
	}
	return false
}
