package gray

import (
	"github.com/liuhailove/seamiter-golang/core/base"
	"github.com/liuhailove/seamiter-golang/logging"
	"strings"
)

// TagTrafficSelector 标签流量选择器
type TagTrafficSelector struct {
	// owner 所归属的流量选择controller
	owner *TrafficSelectorController

	// force
	// 路由结果为空时，是否强制返回
	// force=false: 当路由结果为空，降级请求tag为空的提供者。
	// force=true: 当路由结果为空，直接返回异常。
	force bool
	// 关联资源
	resource string
	// 标签
	tags []GTag
}

func (t *TagTrafficSelector) BoundOwner() *TrafficSelectorController {
	return t.owner
}

// CalculateAllowedResource 计算被允许的执行资源
func (t *TagTrafficSelector) CalculateAllowedResource(ctx *base.EntryContext) (string, string) {
	// 没有标签集合的话，退化到原始接口
	if len(t.tags) == 0 {
		return t.resource, ""
	}
	var classification = ctx.Resource.Classification()
	for _, tag := range t.tags {
		var meet = false
		if base.ResTypeWeb == classification {
			var headers = ctx.Input.Headers
			if headers[tag.TagKey][0] == strings.TrimSpace(tag.TagValue) {
				meet = true
			}
		} else if base.ResTypeMicro == classification {
			var metadata = ctx.Input.MetaData
			if metadata[tag.TagKey] == strings.TrimSpace(tag.TagValue) {
				meet = true
			}
		}
		if meet {
			var resource = tag.TargetResource
			if strings.TrimSpace(tag.TargetVersion) != "" {
				resource += "." + strings.TrimSpace(tag.TargetVersion)
			}
			return resource, tag.EffectiveAddresses
		}
	}
	if t.force {
		return "", ""
	}
	return t.resource, ""
}

// NewTagTrafficSelector 新建标签流量选择器
func NewTagTrafficSelector(owner *TrafficSelectorController, rule *Rule) TrafficSelector {
	if rule == nil {
		logging.Warn("[NewTagTrafficSelector] rule is nil")
		return nil
	}
	if rule.RouterStrategy != TagRouter {
		return nil
	}
	if len(rule.GrayTagList) == 0 {
		// 当标签数组为空是，退化为原始请求资源
		logging.Warn("[NewTagTrafficSelector] gray tag list len is 0")
		if rule.Force {
			// force=true: 当路由结果为空，直接返回nil
			return nil
		}
	}
	var tagTrafficSelector = &TagTrafficSelector{owner: owner, tags: rule.GrayTagList, force: rule.Force, resource: rule.Resource}
	return tagTrafficSelector
}
