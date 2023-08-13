package gray

import (
	"fmt"
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"github.com/buger/jsonparser"
	"github.com/fatih/structs"
	jsoniter "github.com/json-iterator/go"
	"strconv"
	"strings"
)

var (
	jsonHold = jsoniter.ConfigCompatibleWithStandardLibrary
)

// ConditionTrafficSelector 条件流量选择器
type ConditionTrafficSelector struct {
	// owner 所归属的流量选择controller
	owner *TrafficSelectorController

	// force
	// 路由结果为空时，是否强制返回
	// force=false: 当路由结果为空，降级请求tag为空的提供者。
	// force=true: 当路由结果为空，直接返回异常。
	force bool
	// 关联资源
	resource string
	// 条件
	conditions []GCondition
}

func (c *ConditionTrafficSelector) BoundOwner() *TrafficSelectorController {
	return c.owner
}

// CalculateAllowedResource 计算被允许的执行资源
func (c *ConditionTrafficSelector) CalculateAllowedResource(ctx *base.EntryContext) string {
	// 没有条件集合的话，退化到原始接口
	if len(c.conditions) == 0 {
		return c.resource
	}
	var classification = ctx.Resource.Classification()
	for _, cond := range c.conditions {
		var meet = false
		meet = c.DoCheck(ctx, cond, classification)
		if meet {
			var resource = cond.TargetResource
			if strings.TrimSpace(cond.TargetVersion) != "" {
				resource += "." + strings.TrimSpace(cond.TargetVersion)
			}
			return resource
		}
	}
	if c.force {
		return ""
	}
	return c.resource
}

func (c *ConditionTrafficSelector) DoCheck(ctx *base.EntryContext, cond GCondition, classification base.ResourceType) bool {
	var meet = false
	for _, gParam := range cond.GrayConditionParams {
		if gParam.RouterParameterType == ParameterTypeCookie {
			if len(ctx.Input.Cookies[gParam.ParamKey]) == 0 {
				meet = false
			} else {
				meet = c.CheckLock(ctx.Input.Cookies[gParam.ParamKey][0], gParam)
			}
		} else if gParam.RouterParameterType == ParameterTypeBody {
			if len(ctx.Input.Body[gParam.ParamKey]) == 0 {
				meet = false
			} else {
				meet = c.CheckLock(ctx.Input.Body[gParam.ParamKey][0], gParam)
			}
		} else if gParam.RouterParameterType == ParameterTypeHeader {
			if len(ctx.Input.Headers[gParam.ParamKey]) == 0 {
				meet = false
			} else {
				meet = c.CheckLock(ctx.Input.Headers[gParam.ParamKey][0], gParam)
			}
		} else if gParam.RouterParameterType == ParameterTypeParameter {
			if classification == base.ResTypeMicro && structs.IsStruct(ctx.Input.Args[0]) {
				if requestJsonData, err := jsonHold.Marshal(ctx.Input.Args[0]); err == nil {
					valByte, dataType, _, err := jsonparser.Get(requestJsonData, strings.Split(gParam.ParamKey, ".")...)
					if err != nil {
						logging.Warn("[CalculateAllowedResource] get property failed", "property", gParam.ParamKey, "request data", requestJsonData, "err", err)
						meet = false
					} else {
						var val string
						if dataType == jsonparser.Array || dataType == jsonparser.Boolean {
							val = fmt.Sprint(``, string(valByte), ``)
						} else {
							val = string(valByte)
						}
						meet = c.CheckLock(val, gParam)
					}
				}
			} else {
				meet = false
			}
		} else if gParam.RouterParameterType == ParameterTypeMetadata {
			meet = c.CheckLock(ctx.Input.MetaData[gParam.ParamKey], gParam)
		}
		if !meet && cond.Conditions == ALL {
			break
		}
		if meet && cond.Conditions == ANY {
			break
		}
	}
	return meet
}

// CheckLock 条件check
func (c *ConditionTrafficSelector) CheckLock(matchVal string, gParam GParam) bool {
	if OpEqual == gParam.Op {
		return matchVal == gParam.ParamValue
	}
	if OpNotEqual == gParam.Op {
		return matchVal != gParam.ParamValue
	}
	if OpGreater == gParam.Op {
		return matchVal > gParam.ParamValue
	}
	if OpGreaterThan == gParam.Op {
		return matchVal >= gParam.ParamValue
	}
	if OpLess == gParam.Op {
		return matchVal < gParam.ParamValue
	}
	if OpLessThan == gParam.Op {
		return matchVal <= gParam.ParamValue
	}
	if OpIn == gParam.Op {
		var params = strings.Split(gParam.ParamValue, ",")
		for _, item := range params {
			if item == matchVal {
				return true
			}
		}
		return false
	}
	if OpNotIn == gParam.Op {
		var params = strings.Split(gParam.ParamValue, ",")
		for _, item := range params {
			if item == matchVal {
				return false
			}
		}
		return true
	}
	if OpMod100 == gParam.Op {
		matchValInt, err := strconv.ParseInt(matchVal, 10, 64)
		if err != nil {
			logging.Warn("[ConditionTrafficSelector] Do check parser matchVal error", "err", err)
			return false
		}
		paramValInt, err := strconv.ParseInt(gParam.ParamValue, 10, 64)
		{
			if err != nil {
				logging.Warn("[ConditionTrafficSelector] Do check parser param error", "err", err)
				return false
			}
		}
		return matchValInt%100 == paramValInt
	}
	return false
}

// NewConditionTrafficSelector 新建条件流量选择器
func NewConditionTrafficSelector(owner *TrafficSelectorController, rule *Rule) TrafficSelector {
	if rule == nil {
		logging.Warn("[NewWeightTrafficSelector] rule is nil")
		return nil
	}
	if rule.RouterStrategy != ConditionRouter {
		return nil
	}
	if len(rule.GrayConditionList) == 0 {
		// 当条件数组为空是，退化为原始请求资源
		logging.Warn("[NewConditionTrafficSelector] gray weight list len is 0")
		if rule.Force {
			// force=true: 当路由结果为空，直接返回nil
			return nil
		}
		rule.GrayConditionList = append(rule.GrayConditionList, GCondition{EffectiveAddresses: "[0.0.0.0:*]", TargetResource: rule.Resource, TargetVersion: ""})
	}
	var conditionTrafficSelector = &ConditionTrafficSelector{owner: owner, conditions: rule.GrayConditionList, force: rule.Force, resource: rule.Resource}
	return conditionTrafficSelector
}
