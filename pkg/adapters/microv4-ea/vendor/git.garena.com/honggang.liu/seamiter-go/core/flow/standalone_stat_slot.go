package flow

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"github.com/pkg/errors"
)

const (
	StatSlotOrder = 3000
)

var (
	DefaultStandaloneStatSlot = &StandaloneStatSlot{}
)

type StandaloneStatSlot struct {
}

func (s StandaloneStatSlot) Order() uint32 {
	return StatSlotOrder
}

func (s StandaloneStatSlot) OnEntryPassed(ctx *base.EntryContext) {
	res := ctx.Resource.Name()
	for _, tc := range getTrafficControllerListFor(res) {
		if !tc.boundStat.reuseResourceStat {
			if tc.boundStat.writeOnlyMetric != nil {
				tc.boundStat.writeOnlyMetric.AddCount(base.MetricEventPass, int64(ctx.Input.BatchCount))
			} else {
				logging.Error(errors.New("nil independent write statistic"), "Nil statistic for traffic control in StandaloneStatSlot.OnEntryPassed()", "rule", tc.rule)
			}
		}
	}
}

func (s StandaloneStatSlot) OnEntryBlocked(ctx *base.EntryContext, blockError *base.BlockError) {
	// Do nothing
}

func (s StandaloneStatSlot) OnCompleted(ctx *base.EntryContext) {
	// Do nothing
}
