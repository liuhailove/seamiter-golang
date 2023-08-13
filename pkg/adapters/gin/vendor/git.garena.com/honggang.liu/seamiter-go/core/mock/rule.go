package mock

import (
	"fmt"
	"reflect"
	"strconv"
)

// ControlBehavior indicates the traffic shaping behaviour.
// 	// 0:什么也不做，1:抛出异常，2:返回Mock数据，3:等待,4:等待指定时间后抛出异常,5:等待指定时间后返回数据
type ControlBehavior int32

const (
	DoNothing ControlBehavior = iota // 什么也不做，默认行为
	Panic
	Mock
	Waiting
	WaitingThenPanic
	WaitingThenMock
)

const (
	OneResourceLimit = 1000
)

func (t ControlBehavior) String() string {
	switch t {
	case DoNothing:
		return "DoNothing"
	case Panic:
		return "Panic"
	case Mock:
		return "Mock"
	case Waiting:
		return "Waiting"
	case WaitingThenPanic:
		return "WaitingThenPanic"
	case WaitingThenMock:
		return "WaitingThenMock"
	default:
		return strconv.Itoa(int(t))
	}
}

// Strategy 表示作用范围策略
// 0：整个方法，1：具体参数
type Strategy int32

const (
	Func  Strategy = iota // 作用于整个方法
	Param                 // 作用于参数
)

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
	KindArr               = "array"
)

// MatchPattern 匹配模式
// 0:精确匹配，1:前缀匹配，2:后缀匹配，3:包含匹配,4:正则匹配
type MatchPattern int32

// ReplaceMode mock替换模式，0：不替换，1：替换响应，2：替换请求
type ReplaceMode int32

const (
	ExactMatch   MatchPattern = iota // 精确匹配
	PrefixMatch                      // 前缀匹配
	SuffixMatch                      // 后缀匹配
	ContainMatch                     // 包含匹配
	RegularMatch                     // 正则匹配
)
const (
	None ReplaceMode = iota // 不替换
	Resp                    // 替换响应
	Req                     // 替换请求
)

// RuleItem 规则子项，针对具体参数
type RuleItem struct {
	// 参数索引
	WhenParamIdx int32 `json:"whenParamIdx"`
	// 参数名称，优先级大于参数索引
	WhenParamKey string `json:"whenParamKey"`
	// 参数值
	WhenParamValue string `json:"whenParamValue"`
	// 参数类型
	WhenParamKind ParamKind `json:"whenParamKind"`
	// ControlBehavior 控制行为
	ControlBehavior ControlBehavior `json:"controlBehavior"`
	// 匹配模式
	MatchPattern MatchPattern `json:"matchPattern"`
	// mock数据
	ThenReturnMockData string `json:"thenReturnMockData"`
	// ThenReturnWaitingTimeMs 等待时间
	ThenReturnWaitingTimeMs int64 `json:"thenReturnWaitingTimeMs"`
	// mock数据
	ThenThrowMsg string `json:"thenThrowMsg"`
	// mockReplace mock替换，开启则WhenParamValue不起作用，0不替换，1替换响应，2替换请求
	MockReplace ReplaceMode `json:"mockReplace"`
	// ReplaceAttribute 替换mock中的属性值
	ReplaceAttribute string `json:"replaceAttribute"`
	// RequestHold 请求保留，如果为true，则在内存中保存一份请求，并上报到Server，默认总体保留5000个请求
	RequestHold bool `json:"requestHold"`
	// 附加参数Key
	AdditionalItemKey string `json:"additionalItemKey"`
	// 附加参数Value
	AdditionalItemValue string      `json:"additionalItemValue"`
	TmpData             interface{} `json:"-"`
}

func (r *RuleItem) isEqualTo(newRuleItem *RuleItem) bool {
	return r.WhenParamIdx == newRuleItem.WhenParamIdx && r.WhenParamKey == newRuleItem.WhenParamKey && r.WhenParamValue == newRuleItem.WhenParamValue &&
		r.ControlBehavior == newRuleItem.ControlBehavior && r.ThenReturnMockData == newRuleItem.ThenReturnMockData && r.ThenReturnWaitingTimeMs == newRuleItem.ThenReturnWaitingTimeMs &&
		r.ThenThrowMsg == newRuleItem.ThenThrowMsg && r.MatchPattern == newRuleItem.MatchPattern && r.MockReplace == newRuleItem.MockReplace && r.ReplaceAttribute == newRuleItem.ReplaceAttribute &&
		r.AdditionalItemKey == newRuleItem.AdditionalItemValue && r.AdditionalItemValue == newRuleItem.AdditionalItemValue

}

// Op 操作类型，并或者或
type Op int32

const (
	And Op = iota // 并
	Or            // 或
)

// AdditionalType 附加类型,Web Header/ grpc context
type AdditionalType int32

const (
	WebHeader AdditionalType = iota
	GrpcContext
)

// AdditionalItem 附加结构体
type AdditionalItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Rule mock 规则.
type Rule struct {
	// unique id
	Id string `json:"id,omitempty"`
	// resource name
	Resource        string          `json:"resource"`
	ControlBehavior ControlBehavior `json:"controlBehavior"`
	Strategy        Strategy        `json:"strategy"`
	// 操作类型，仅读header或者context不为空时起作用
	Op             Op             `json:"op"`
	AdditionalType AdditionalType `json:"additionalType"`
	// 参数key=value数组，关系可以是"并"或者为"或"
	AdditionalItems []AdditionalItem `json:"additionalItems"`
	// mock数据
	ThenReturnMockData string `json:"thenReturnMockData"`
	// ThenReturnWaitingTimeMs 等待时间
	ThenReturnWaitingTimeMs int64 `json:"thenReturnWaitingTimeMs"`
	// mock数据
	ThenThrowMsg string `json:"thenThrowMsg"`

	// RequestHold 请求保留，如果为true，则在内存中保存一份请求，并上报到Server，默认总体保留5000个请求
	RequestHold bool `json:"requestHold"`

	// SpecificItems indicates the special mock data for specific value
	SpecificItems []RuleItem `json:"specificItems"`
}

func (r *Rule) String() string {
	// fallback string
	return fmt.Sprintf("{id=%s, resource=%s, controlBehavior=%d,strategy=%d,thenReturnWaitingTimeMs=%d,thenThrowMsg=%s,requestHold=%v}", r.Id, r.Resource, r.ControlBehavior, r.Strategy, r.ThenReturnWaitingTimeMs, r.ThenThrowMsg, r.RequestHold)
}

func (r *Rule) isStatReusable(newRule *Rule) bool {
	if newRule == nil {
		return false
	}
	return r.Resource == newRule.Resource && r.ControlBehavior == newRule.ControlBehavior && r.Strategy == newRule.Strategy && r.ThenReturnMockData == newRule.ThenReturnMockData && r.ThenThrowMsg == newRule.ThenThrowMsg && reflect.DeepEqual(r.SpecificItems, newRule.SpecificItems)
}

func (r *Rule) ResourceName() string {
	return r.Resource
}

func (r *Rule) isEqualsToBase(newRule *Rule) bool {
	if newRule == nil {
		return false
	}
	var baseEqual = r.Resource == newRule.Resource && r.ControlBehavior == newRule.ControlBehavior && r.Strategy == newRule.Strategy && r.ThenReturnMockData == newRule.ThenReturnMockData && r.ThenThrowMsg == newRule.ThenThrowMsg && r.RequestHold == newRule.RequestHold
	if !baseEqual {
		return false
	}
	if len(r.SpecificItems) != len(newRule.SpecificItems) {
		return false
	}
	for idx, item := range r.SpecificItems {
		if !item.isEqualTo(&newRule.SpecificItems[idx]) {
			return false
		}
	}
	if len(r.AdditionalItems) != len(newRule.AdditionalItems) {
		return false
	}
	for idx, item := range r.AdditionalItems {
		if item.Key != newRule.AdditionalItems[idx].Key || item.Value != newRule.AdditionalItems[idx].Value {
			return false
		}
	}
	return true
}

func (r *Rule) isEqualTo(newRule *Rule) bool {
	return r.isEqualsToBase(newRule)
}
