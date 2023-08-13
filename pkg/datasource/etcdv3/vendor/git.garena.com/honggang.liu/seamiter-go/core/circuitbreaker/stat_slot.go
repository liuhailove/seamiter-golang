package circuitbreaker

import "git.garena.com/honggang.liu/seamiter-go/core/base"

const (
	StatSlotOrder = 5000
)

var (
	DefaultMetricStatSlot = &MetricStatSlot{}
)

// MetricStatSlot records metrics for circuit breaker on invocation completed.
// MetricStatSlot must be filled into slot chain if circuit breaker is alive.
type MetricStatSlot struct {
}

func (m MetricStatSlot) Order() uint32 {
	return StatSlotOrder
}

func (m MetricStatSlot) OnEntryPassed(ctx *base.EntryContext) {
	// Do nothing
	return
}

func (m MetricStatSlot) OnEntryBlocked(ctx *base.EntryContext, blockError *base.BlockError) {
	// Do nothing
	return
}

func (m MetricStatSlot) OnCompleted(ctx *base.EntryContext) {
	res := ctx.Resource.Name()
	err := ctx.Err()
	rt := ctx.Rt()
	for _, cb := range getBreakersOfResource(res) {
		cb.OnRequestComplete(rt, err)
	}
}
