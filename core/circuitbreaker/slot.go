package circuitbreaker

import (
	"github.com/liuhailove/seamiter-golang/core/base"
	"github.com/liuhailove/seamiter-golang/util"
	"strings"
)

const (
	RuleCheckSlotOrder = 5000
)

var (
	DefaultSlot = &Slot{}
)

type Slot struct {
}

func (s *Slot) Order() uint32 {
	return RuleCheckSlotOrder
}

// Initial
//
// 初始化，如果有初始化工作放入其中
func (s *Slot) Initial() {
}

func (s *Slot) Check(ctx *base.EntryContext) *base.TokenResult {
	resource := ctx.Resource.Name()
	result := ctx.RuleCheckResult
	if len(resource) == 0 {
		return result
	}
	if passed, rule := checkPass(ctx); !passed {
		msg := "circuit breaker check blocked"
		if result == nil {
			result = base.NewTokenResultBlockedWithCause(base.BlockTypeCircuitBreaking, msg, rule, nil)
		} else {
			result.ResetToBlockedWithCause(base.BlockTypeCircuitBreaking, msg, rule, nil)
		}
	}
	return result
}

func checkPass(ctx *base.EntryContext) (bool, *Rule) {
	breakers := getBreakersOfResource(ctx.Resource.Name())
	for _, breaker := range breakers {
		// 来源检查
		var needContinueCheck = true
		if breaker.BoundRule().LimitApp != "" && !strings.EqualFold(breaker.BoundRule().LimitApp, "default") {
			if !util.Contains(ctx.FromService, strings.Split(breaker.BoundRule().LimitApp, ",")) {
				needContinueCheck = false
			}
		}
		if !needContinueCheck {
			continue
		}
		passed := breaker.TryPass(ctx)
		if !passed {
			return false, breaker.BoundRule()
		}
	}
	return true, nil
}
