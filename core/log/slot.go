package log

import "github.com/liuhailove/seamiter-golang/core/base"

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

// Initial
//
// 初始化，如果有初始化工作放入其中
func (s Slot) Initial() {
}

func (s Slot) OnEntryPassed(_ *base.EntryContext) {
}

func (s Slot) OnEntryBlocked(ctx *base.EntryContext, blockError *base.BlockError) {
}

func (s Slot) OnCompleted(ctx *base.EntryContext) {
}
