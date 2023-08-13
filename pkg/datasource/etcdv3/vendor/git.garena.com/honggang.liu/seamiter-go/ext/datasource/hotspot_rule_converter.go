package datasource

import (
	"fmt"
	"git.garena.com/honggang.liu/seamiter-go/core/hotspot"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"github.com/pkg/errors"
	"strconv"
)

type HotspotRule struct {
	// ID is the unique id
	ID string `json:"id,omitempty"`
	// Resource is the resource name
	Resource string `json:"resource"`
	// MetricType indicates the metric type for checking logic.
	// For Concurrency metric, hotspot module will check the each hot parameter's concurrency,
	//		if concurrency exceeds the Threshold, reject the traffic directly.
	// For QPS metric, hotspot module will check the each hot parameter's QPS,
	//		the ControlBehavior decides the behavior of traffic shaping controller
	MetricType hotspot.MetricType `json:"metricType"`
	// ControlBehavior indicates the traffic shaping behaviour.
	// ControlBehavior only takes effect when MetricType is QPS
	ControlBehavior hotspot.ControlBehavior `json:"controlBehavior"`
	// ParamIdx is the index in context arguments slice.
	// if ParamIdx is greater than or equals to zero, ParamIdx means the <ParamIdx>-th parameter
	// if ParamIdx is the negative, ParamIdx means the reversed <ParamIdx>-th parameter
	ParamIdx int `json:"paramIdx"`
	// Threshold is the threshold to trigger rejection
	Threshold float64 `json:"threshold"`
	// MaxQueueingTimeMs only takes effect when ControlBehavior is Throttling and MetricType is QPS
	MaxQueueingTimeMs int64 `json:"maxQueueingTimeMs"`
	// BurstCount is the silent count
	// BurstCount only takes effect when ControlBehavior is Reject and MetricType is QPS
	BurstCount int64 `json:"burstCount"`
	// DurationInSec is the time interval in statistic
	// DurationInSec only takes effect when MetricType is QPS
	DurationInSec int64 `json:"durationInSec"`
	// ParamsMaxCapacity is the max capacity of cache statistic
	ParamsMaxCapacity int64           `json:"paramsMaxCapacity"`
	ParamFlowItems    []ParamFlowItem `json:"paramFlowItemList"`
}

// ParamKind represents the Param kind.
type ParamKind string

const (
	KindInt     ParamKind = "int"
	KindInt32             = "int32"
	KindInt64             = "int64"
	KindString            = "string"
	KindBool              = "bool"
	KindFloat32           = "float32"
	KindFloat64           = "float64"
	KindByte              = "byte"
	KindSum               = ""
)

// ParamFlowItem indicates the specific param, contain the supported param kind and concrete value.
type ParamFlowItem struct {
	ParamKind  ParamKind `json:"paramKind"`  // 参数类型
	ParamValue string    `json:"paramValue"` // 例外参数值，如当参数值为100时
	Threshold  int64     `json:"threshold"`  // 当参数值为Value时的阈值
}

func (s *ParamFlowItem) String() string {
	return fmt.Sprintf("ParamFlowItem: [ParamKind: %+v, Value: %s]", s.ParamKind, s.ParamValue)
}

// parseSpecificItems parses the ParamFlowItem as real value.
func parseSpecificItems(source []ParamFlowItem) map[interface{}]int64 {
	ret := make(map[interface{}]int64, len(source))
	if len(source) == 0 {
		return ret
	}
	for _, item := range source {
		switch item.ParamKind {
		case KindInt:
			realVal, err := strconv.Atoi(item.ParamValue)
			if err != nil {
				logging.Error(errors.Wrap(err, "parseSpecificItems error"), "Failed to parse value for int specific item", "itemValKind", item.ParamKind, "itemValStr", item.ParamValue, "itemThreshold", item.Threshold)
				continue
			}
			ret[realVal] = item.Threshold
		case KindInt32:
			realVal, err := strconv.ParseInt(item.ParamValue, 10, 32)
			if err != nil {
				logging.Error(errors.Wrap(err, "parseSpecificItems error"), "Failed to parse value for int specific item", "itemValKind", item.ParamKind, "itemValStr", item.ParamValue, "itemThreshold", item.Threshold)
				continue
			}
			ret[realVal] = item.Threshold
		case KindInt64:
			realVal, err := strconv.ParseInt(item.ParamValue, 10, 64)
			if err != nil {
				logging.Error(errors.Wrap(err, "parseSpecificItems error"), "Failed to parse value for int specific item", "itemValKind", item.ParamKind, "itemValStr", item.ParamValue, "itemThreshold", item.Threshold)
				continue
			}
			ret[realVal] = item.Threshold

		case KindString:
			ret[item.ParamValue] = item.Threshold
		case KindBool:
			realVal, err := strconv.ParseBool(item.ParamValue)
			if err != nil {
				logging.Error(errors.Wrap(err, "parseSpecificItems error"), "Failed to parse value for bool specific item", "itemValKind", item.ParamKind, "itemValStr", item.ParamValue, "itemThreshold", item.Threshold)
				continue
			}
			ret[realVal] = item.Threshold
		case KindFloat32:
			realVal, err := strconv.ParseFloat(item.ParamValue, 32)
			if err != nil {
				logging.Error(errors.Wrap(err, "parseSpecificItems error"), "Failed to parse value for float specific item", "itemValKind", item.ParamKind, "itemValStr", item.ParamValue, "itemThreshold", item.Threshold)
				continue
			}
			ret[realVal] = item.Threshold
		case KindFloat64:
			realVal, err := strconv.ParseFloat(item.ParamValue, 64)
			if err != nil {
				logging.Error(errors.Wrap(err, "parseSpecificItems error"), "Failed to parse value for float specific item", item.ParamKind, "itemValKind", item.ParamKind, "itemValStr", item.ParamValue, "itemThreshold", item.Threshold)
				continue
			}
			ret[realVal] = item.Threshold
		default:
			logging.Error(errors.New("Unsupported kind for specific item"), "", item.ParamKind, "itemValKind", item.ParamKind, "itemValStr", item.ParamValue, "itemThreshold", item.Threshold)
		}
	}
	return ret
}

// transToSpecificItems trans to the ParamFlowItem as real value.
func transToSpecificItems(source map[interface{}]int64) []ParamFlowItem {
	var ret = make([]ParamFlowItem, 0)
	if len(source) == 0 {
		return ret
	}
	for key, value := range source {
		switch key.(type) {
		case int:
			param := ParamFlowItem{ParamKind: KindInt, ParamValue: key.(string), Threshold: value}
			ret = append(ret, param)
		case int32:
			param := ParamFlowItem{ParamKind: KindInt32, ParamValue: key.(string), Threshold: value}
			ret = append(ret, param)
		case int64:
			param := ParamFlowItem{ParamKind: KindInt64, ParamValue: key.(string), Threshold: value}
			ret = append(ret, param)
		case string:
			param := ParamFlowItem{ParamKind: KindString, ParamValue: key.(string), Threshold: value}
			ret = append(ret, param)
		case float32:
			param := ParamFlowItem{ParamKind: KindFloat32, ParamValue: key.(string), Threshold: value}
			ret = append(ret, param)
		case float64:
			param := ParamFlowItem{ParamKind: KindFloat64, ParamValue: key.(string), Threshold: value}
			ret = append(ret, param)
		case byte:
			param := ParamFlowItem{ParamKind: KindByte, ParamValue: key.(string), Threshold: value}
			ret = append(ret, param)
		case bool:
			param := ParamFlowItem{ParamKind: KindBool, ParamValue: key.(string), Threshold: value}
			ret = append(ret, param)
		default:
			logging.Error(errors.New("Unsupported kind for specific item"), "", key)
		}
	}
	return ret
}
