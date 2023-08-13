package mock

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"github.com/fatih/structs"
)

type TrafficShapingController interface {
	PerformCheckingFunc(ctx *base.EntryContext) *base.TokenResult
	PerformCheckingArgs(ctx *base.EntryContext) *base.TokenResult

	BoundRule() *Rule
	ArgsCheck(ctx *base.EntryContext) *RuleItem
}

type baseTrafficShapingController struct {
	r             *Rule
	strategy      Strategy
	res           string
	specificItems []RuleItem
}

func newBaseTrafficShapingController(r *Rule) *baseTrafficShapingController {
	if r.SpecificItems == nil {
		r.SpecificItems = []RuleItem{}
	}
	return &baseTrafficShapingController{
		r:             r,
		res:           r.Resource,
		specificItems: r.SpecificItems,
		strategy:      r.Strategy,
	}
}

func (c *baseTrafficShapingController) BoundRule() *Rule {
	return c.r
}

func (c *baseTrafficShapingController) extractArgs(ctx *base.EntryContext) []interface{} {
	args := ctx.Input.Args
	return args
}

func (c *baseTrafficShapingController) extractAttachmentArgs(ctx *base.EntryContext) map[interface{}]interface{} {
	attachments := ctx.Input.Attachments
	if attachments == nil {
		if logging.DebugEnabled() {
			logging.Debug("[paramKey] The attachments of ctx is nil", "args", attachments)
		}
		return nil
	}
	return attachments
}

// ArgsCheck 参数检查
func (c *baseTrafficShapingController) ArgsCheck(ctx *base.EntryContext) *RuleItem {
	if c == nil {
		return nil
	}
	if len(c.specificItems) == 0 {
		return nil
	}
	attachmentArgs := c.extractAttachmentArgs(ctx)
	if attachmentArgs != nil && len(attachmentArgs) > 0 {
		for _, item := range c.specificItems {
			if item.WhenParamKey == "" {
				if logging.DebugEnabled() {
					logging.Debug("[paramKey] The param key is nil",
						"args", attachmentArgs, "paramKey", item.WhenParamKey)
				}
				continue
			}
			arg, ok := attachmentArgs[item.WhenParamKey]
			if !ok {
				if logging.DebugEnabled() {
					logging.Debug("[paramKey] extracted data does not exist",
						"args", attachmentArgs, "paramKey", item.WhenParamKey)
				}
				continue
			}
			if item.WhenParamValue == arg {
				return &item
			}
		}
	}
	args := c.extractArgs(ctx)
	if args != nil {
		if ctx.Resource.Classification() == base.ResTypeMicro {
			// micro此处的args只可能有一个参数
			var dataMap = structs.Map(args[0])
			for _, item := range c.specificItems {
				for k, v := range dataMap {
					if item.WhenParamKey == k {
						if item.WhenParamValue == v {
							return &item
						}
						// 否则直接break，避免
						break
					}
				}
			}
		} else {
			for _, item := range c.specificItems {
				var idx int
				if item.WhenParamIdx < 0 {
					idx = len(args) + idx
				}
				if idx < 0 {
					if logging.DebugEnabled() {
						logging.Debug("[extractArgs] The param index of mock traffic shaping controller is invalid",
							"args", args, "paramIndex", item.WhenParamIdx)
					}
					continue
				}
				if idx >= len(args) {
					if logging.DebugEnabled() {
						logging.Debug("[extractArgs] The argument in index doesn't exist",
							"args", args, "paramIndex", item.WhenParamIdx)
					}
					continue
				}
				if item.WhenParamValue == args[item.WhenParamIdx] {
					return &item
				}
			}
		}
	}
	return nil
}
