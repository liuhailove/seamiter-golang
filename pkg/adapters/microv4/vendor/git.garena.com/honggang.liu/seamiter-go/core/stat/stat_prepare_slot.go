package stat

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
)

const (
	PrepareSlotOrder = 1000
)

var (
	DefaultResourceNodePrepareSlot = &ResourceNodePrepareSlot{}
)

type ResourceNodePrepareSlot struct {
}

func (s *ResourceNodePrepareSlot) Order() uint32 {
	return PrepareSlotOrder
}

// Initial
//
// 初始化，如果有初始化工作放入其中
func (s ResourceNodePrepareSlot) Initial() {
}

func (s *ResourceNodePrepareSlot) Prepare(ctx *base.EntryContext) {
	node := GetOrCreateResourceNode(ctx.Resource.Name(), ctx.Resource.Classification())
	// Set the resource node to the context.
	ctx.StatNode = node
}
