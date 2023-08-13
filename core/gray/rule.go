package gray

import (
	"fmt"
	"strconv"
)

// RouterStrategy 路由策略
type RouterStrategy int

const (
	ConditionRouter RouterStrategy = 1 // ConditionRouter 条件路由
	TagRouter       RouterStrategy = 2 // TagRouter 标签路由
	WeightRouter    RouterStrategy = 3 // WeightRouter 权重路由
)

func (t RouterStrategy) String() string {
	switch t {
	case ConditionRouter:
		return "ConditionRouter"
	case TagRouter:
		return "TagRouter"
	case WeightRouter:
		return "WeightRouter"
	default:
		return strconv.Itoa(int(t))
	}
}

// RouterParameterType 路由参数类型
type RouterParameterType int

const (
	ParameterTypeCookie    RouterParameterType = 1
	ParameterTypeHeader    RouterParameterType = 2
	ParameterTypeParameter RouterParameterType = 3
	ParameterTypeBody      RouterParameterType = 4
	ParameterTypeMetadata  RouterParameterType = 5
)

func (t RouterParameterType) String() string {
	switch t {
	case ParameterTypeCookie:
		return "ParameterTypeCookie"
	case ParameterTypeHeader:
		return "ParameterTypeHeader"
	case ParameterTypeParameter:
		return "ParameterTypeParameter"
	case ParameterTypeBody:
		return "ParameterTypeBody"
	case ParameterTypeMetadata:
		return "ParameterTypeMetadata"
	default:
		return strconv.Itoa(int(t))
	}
}

// Op 运算符
type Op int

const (
	OpEqual       Op = 1
	OpNotEqual    Op = 2
	OpGreater     Op = 3
	OpGreaterThan Op = 4
	OpLess        Op = 5
	OpLessThan    Op = 6
	OpIn          Op = 7
	OpNotIn       Op = 8
	OpMod100      Op = 9
)

func (t Op) String() string {
	switch t {
	case OpEqual:
		return "OpEqual"
	case OpNotEqual:
		return "OpNotEqual"
	case OpGreater:
		return "OpGreater"
	case OpGreaterThan:
		return "OpGreaterThan"
	case OpLess:
		return "OpLess"
	case OpLessThan:
		return "OpLessThan"
	case OpIn:
		return "OpIn"
	case OpNotIn:
		return "OpNotIn"
	case OpMod100:
		return "OpMod100"
	default:
		return strconv.Itoa(int(t))
	}
}

// GParam 灰度Param
type GParam struct {
	// 路由参数类型
	RouterParameterType RouterParameterType `json:"routerParameterType"`
	// 参数key
	ParamKey string `json:"paramKey"`
	// 参数value
	ParamValue string `json:"paramValue"`
	// 运算符
	Op Op `json:"op"`
}

func (r *GParam) isEqualTo(newGParam *GParam) bool {
	return r.RouterParameterType == newGParam.RouterParameterType && r.ParamKey == newGParam.ParamKey && r.ParamValue == newGParam.ParamValue && r.Op == newGParam.Op

}

func (r *GParam) String() string {
	// fallback string
	return fmt.Sprintf("{routerParameterType=%s, paramKey=%s，paramValue=%s, op=%s}", r.RouterParameterType, r.ParamKey, r.ParamValue, r.Op)
}

func (r *GParam) isStatReusable(newGParam *GParam) bool {
	if newGParam == nil {
		return false
	}
	return r.isEqualTo(newGParam)
}

// Conditions 条件
type Conditions int

const (
	ALL Conditions = 1
	ANY            = 2
)

func (t Conditions) String() string {
	switch t {
	case ALL:
		return "ALL"
	case ANY:
		return "ANY"
	default:
		return strconv.Itoa(int(t))
	}
}

// GCondition 灰度条件集合
type GCondition struct {
	// 生效地址
	// 配置是否只对某几个特定实例生效。
	// 所有实例：addresses: ["0.0.0.0"] 或addresses: ["0.0.0.0:*"] 具体由side值决定。
	// 指定实例：addresses[实例地址列表]。
	EffectiveAddresses string `json:"effectiveAddresses"`
	// 目标资源
	TargetResource string `json:"targetResource"`
	// 目标版本
	TargetVersion string `json:"targetVersion"`
	//  灰度条件
	//  分为所有和任意两类：
	//  所有：当设置的灰度条件全部满足时，才灰度。
	//  任意：当设置的触发条件满足一条或以上时，就灰度
	Conditions Conditions `json:"conditions"`

	// GrayConditionParams 条件集合
	GrayConditionParams []GParam `json:"grayConditionParams"`
}

func (r *GCondition) isEqualTo(newGCondition *GCondition) bool {
	var eq = r.EffectiveAddresses == newGCondition.EffectiveAddresses && r.TargetResource == newGCondition.TargetResource && r.TargetVersion == newGCondition.TargetVersion && r.Conditions == newGCondition.Conditions
	if !eq {
		return false
	}
	if len(r.GrayConditionParams) != len(newGCondition.GrayConditionParams) {
		return false
	}
	for idx, item := range r.GrayConditionParams {
		if !item.isEqualTo(&newGCondition.GrayConditionParams[idx]) {
			return false
		}
	}
	return true
}

func (r *GCondition) String() string {
	return fmt.Sprintf("{effectiveAddresses=%s, targetResource=%s，targetVersion=%s, conditions=%s,grayConditionParams=%v}", r.EffectiveAddresses, r.TargetResource, r.TargetVersion, r.Conditions, r.GrayConditionParams)
}

func (r *GCondition) isStatReusable(newGCondition *GCondition) bool {
	if newGCondition == nil {
		return false
	}
	return r.isEqualTo(newGCondition)
}

// GTag 灰度标签
type GTag struct {
	TagKey   string `json:"tagKey"`
	TagValue string `json:"tagValue"`
	// 生效地址
	// 配置是否只对某几个特定实例生效。
	// 所有实例：addresses: ["0.0.0.0"] 或addresses: ["0.0.0.0:*"] 具体由side值决定。
	// 指定实例：addresses[实例地址列表]。
	EffectiveAddresses string `json:"effectiveAddresses"`
	// 目标资源
	TargetResource string `json:"targetResource"`
	// 目标版本
	TargetVersion string `json:"targetVersion"`
}

func (r *GTag) isEqualTo(newGTag *GTag) bool {
	return r.TagKey == newGTag.TagKey && r.TagValue == newGTag.TagValue && r.EffectiveAddresses == newGTag.EffectiveAddresses && r.TargetResource == newGTag.TargetResource && r.TargetVersion == newGTag.TargetVersion
}

func (r *GTag) String() string {
	return fmt.Sprintf("{tagKey=%s, tagValue=%s，effectiveAddresses=%s, targetResource=%s,targetVersion=%v}", r.TagKey, r.TagKey, r.EffectiveAddresses, r.TargetResource, r.TargetVersion)
}

func (r *GTag) isStatReusable(newGTag *GTag) bool {
	if newGTag == nil {
		return false
	}
	return r.isEqualTo(newGTag)
}

// GWeight 灰度权重
type GWeight struct {
	// 生效地址
	// 配置是否只对某几个特定实例生效。
	// 所有实例：addresses: ["0.0.0.0"] 或addresses: ["0.0.0.0:*"] 具体由side值决定。
	// 指定实例：addresses[实例地址列表]。
	EffectiveAddresses string `json:"effectiveAddresses"`
	// 目标资源
	TargetResource string `json:"targetResource"`
	// 目标版本
	TargetVersion string `json:"targetVersion"`
	// 权重
	Weight float64 `json:"weight"`
}

func (r *GWeight) isEqualTo(newGWeight *GWeight) bool {
	return r.EffectiveAddresses == newGWeight.EffectiveAddresses && r.TargetResource == newGWeight.TargetResource && r.TargetVersion == newGWeight.TargetVersion && r.Weight == newGWeight.Weight
}

func (r *GWeight) String() string {
	return fmt.Sprintf("{ffectiveAddresses=%s, targetResource=%s,targetVersion=%s,weight=%f}", r.EffectiveAddresses, r.TargetResource, r.TargetVersion, r.Weight)
}

func (r *GWeight) isStatReusable(newGWeight *GWeight) bool {
	if newGWeight == nil {
		return false
	}
	return r.isEqualTo(newGWeight)
}

type Rule struct {
	// ID 唯一ID，可选
	ID string `json:"id,omitempty"`
	// LimitApp 限制应用程序
	// 将受来源限制的应用程序名称。
	// 默认的limitApp是{@code default}，表示允许所有源端应用。
	// 对于权限规则，多个源名称可以用逗号（','）分隔。
	LimitApp string `json:"limitApp"`
	// Resource 资源名称
	Resource string `json:"resource"`
	// GrayTag 灰度标签
	GrayTag string `json:"grayTag"`
	// LinkPass 是否链路传递
	LinkPass bool `json:"linkPass"`
	// RouterStrategy 路由策略
	RouterStrategy RouterStrategy `json:"routerStrategy"`
	// Force
	// 路由结果为空时，是否强制返回
	// force=false: 当路由结果为空，降级请求tag为空的提供者。
	// force=true: 当路由结果为空，直接返回异常。
	Force bool `json:"force"`
	// BlackIpAddresses 黑名单
	BlackIpAddresses string `json:"blackIpAddresses"`
	// WhiteIpAddresses 白名单
	WhiteIpAddresses string `json:"whiteIpAddresses"`

	// GrayConditionList 灰度条件
	GrayConditionList []GCondition `json:"grayConditionList"`
	// GrayTagList 灰度Tag
	GrayTagList []GTag `json:"grayTagList"`
	// GrayWeightList 灰度权重
	GrayWeightList []GWeight `json:"grayWeightList"`
}

func (r *Rule) isEqualTo(newRule *Rule) bool {
	var baseEqual = r.LimitApp == newRule.LimitApp && r.Resource == newRule.Resource && r.GrayTag == newRule.GrayTag && r.LinkPass == newRule.LinkPass &&
		r.RouterStrategy == newRule.RouterStrategy && r.Force == newRule.Force && r.BlackIpAddresses == newRule.BlackIpAddresses && r.WhiteIpAddresses == newRule.WhiteIpAddresses
	if !baseEqual {
		return false
	}
	if len(r.GrayConditionList) != len(newRule.GrayWeightList) {
		return false
	}
	for idx, item := range r.GrayConditionList {
		if !item.isEqualTo(&newRule.GrayConditionList[idx]) {
			return false
		}
	}
	if len(r.GrayTagList) != len(r.GrayTagList) {
		return false
	}
	for idx, item := range r.GrayTagList {
		if !item.isEqualTo(&newRule.GrayTagList[idx]) {
			return false
		}
	}

	if len(r.GrayWeightList) != len(r.GrayWeightList) {
		return false
	}
	for idx, item := range r.GrayWeightList {
		if !item.isEqualTo(&newRule.GrayWeightList[idx]) {
			return false
		}
	}
	return true
}

func (r *Rule) String() string {
	return fmt.Sprintf("{resource=%s, grayTag=%s,linkPass=%t,routerStrategy=%d,force=%t,blackIpAddresses=%s,whiteIpAddresses=%s}", r.Resource, r.GrayTag, r.LinkPass, r.RouterStrategy, r.Force, r.BlackIpAddresses, r.WhiteIpAddresses)
}

func (r *Rule) isStatReusable(newRule *Rule) bool {
	if newRule == nil {
		return false
	}
	return r.isEqualTo(newRule)
}
func (r *Rule) ResourceName() string {
	return r.Resource
}
