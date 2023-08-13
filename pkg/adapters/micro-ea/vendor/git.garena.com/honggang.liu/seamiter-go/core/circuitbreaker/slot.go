package circuitbreaker

import "git.garena.com/honggang.liu/seamiter-go/core/base"

const (
	RuleCheckSlotOrder = 5000
)

var (
	DefaultSlot = &Slot{}
)

type Slot struct {
}

func (s Slot) Order() uint32 {
	return RuleCheckSlotOrder
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
		passed := breaker.TryPass(ctx)
		if !passed {
			return false, breaker.BoundRule()
		}
	}
	return true, nil
}
