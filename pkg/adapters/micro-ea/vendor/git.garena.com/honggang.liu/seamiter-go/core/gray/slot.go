package gray

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"math"
)

const (
	// RuleCheckSlotOrder 灰度策略需要在最后一个，因为执行完此结果就返回了
	RuleCheckSlotOrder = math.MaxInt32
)

var (
	DefaultSlot = &Slot{}
)

type Slot struct {
}

func (s Slot) Order() uint32 {
	return RuleCheckSlotOrder
}

func (s Slot) Router(ctx *base.EntryContext) *base.TokenResult {
	res := ctx.Resource.Name()
	tcs := getTrafficControllerListFor(res)
	result := ctx.RuleCheckResult

	// Check rules in order
	for _, tc := range tcs {
		if tc == nil {
			logging.Warn("[GraySlot Check]Nil traffic controller found", "resourceName", res)
			continue
		}
		r := tc.PerformSelecting(ctx)
		if r == nil {
			continue
		}
		return r

	}
	return result
}
