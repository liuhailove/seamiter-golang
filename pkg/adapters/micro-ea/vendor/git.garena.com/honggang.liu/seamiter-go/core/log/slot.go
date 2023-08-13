package log

import "git.garena.com/honggang.liu/seamiter-go/core/base"

const (
	StatSlotOrder = 2000
)

var (
	DefaultSlot = &Slot{}
)

type Slot struct {
}

func (s Slot) Order() uint32 {
	return StatSlotOrder
}

func (s Slot) OnEntryPassed(_ *base.EntryContext) {
}

func (s Slot) OnEntryBlocked(ctx *base.EntryContext, blockError *base.BlockError) {
}

func (s Slot) OnCompleted(ctx *base.EntryContext) {
}
