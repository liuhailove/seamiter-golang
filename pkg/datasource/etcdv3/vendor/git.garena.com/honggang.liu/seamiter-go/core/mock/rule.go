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

// RuleItem 规则子项，针对具体参数
type RuleItem struct {
	// 参数索引
	WhenParamIdx int32 `json:"whenParamIdx"`
	// 参数名称，优先级大于参数索引
	WhenParamKey string `json:"whenParamKey"`
	// 参数值
	WhenParamValue  interface{}     `json:"whenParamValue"`
	ControlBehavior ControlBehavior `json:"controlBehavior"`
	// mock数据
	ThenReturnMockData string `json:"thenReturnMockData"`
	// ThenReturnWaitingTimeMs 等待时间
	ThenReturnWaitingTimeMs int64 `json:"thenReturnWaitingTimeMs"`
	// mock数据
	ThenThrowMsg string `json:"thenThrowMsg"`
}

func (r *RuleItem) isEqualTo(newRuleItem *RuleItem) bool {
	return r.WhenParamIdx == newRuleItem.WhenParamIdx && r.WhenParamKey == newRuleItem.WhenParamKey && r.WhenParamValue == newRuleItem.WhenParamValue &&
		r.ControlBehavior == newRuleItem.ControlBehavior && r.ThenReturnMockData == newRuleItem.ThenReturnMockData && r.ThenReturnWaitingTimeMs == newRuleItem.ThenReturnWaitingTimeMs &&
		r.ThenThrowMsg == newRuleItem.ThenThrowMsg
}

// Rule mock 规则.
type Rule struct {
	// unique id
	Id string `json:"id,omitempty"`
	// resource name
	Resource        string          `json:"resource"`
	ControlBehavior ControlBehavior `json:"controlBehavior"`
	Strategy        Strategy        `json:"strategy"`
	// mock数据
	ThenReturnMockData string `json:"thenReturnMockData"`
	// ThenReturnWaitingTimeMs 等待时间
	ThenReturnWaitingTimeMs int64 `json:"thenReturnWaitingTimeMs"`
	// mock数据
	ThenThrowMsg string `json:"thenThrowMsg"`
	// SpecificItems indicates the special mock data for specific value
	SpecificItems []RuleItem `json:"specificItems"`
}

func (r *Rule) String() string {
	// fallback string
	return fmt.Sprintf("{id=%s, resource=%s, controlBehavior=%d,strategy=%d,thenReturnWaitingTimeMs=%d,thenThrowMsg=%s}", r.Id, r.Resource, r.ControlBehavior, r.Strategy, r.ThenReturnWaitingTimeMs, r.ThenThrowMsg)
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
	var baseEqual = r.Resource == newRule.Resource && r.ControlBehavior == newRule.ControlBehavior && r.Strategy == newRule.Strategy && r.ThenReturnMockData == newRule.ThenReturnMockData && r.ThenThrowMsg == newRule.ThenThrowMsg
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
	return true
}

func (r *Rule) isEqualTo(newRule *Rule) bool {
	return r.isEqualsToBase(newRule)
}
