package api

import (
	"github.com/liuhailove/seamiter-golang/core/base"
	"github.com/liuhailove/seamiter-golang/core/circuitbreaker"
	"github.com/liuhailove/seamiter-golang/core/flow"
	"github.com/liuhailove/seamiter-golang/core/gray"
	"github.com/liuhailove/seamiter-golang/core/hotspot"
	"github.com/liuhailove/seamiter-golang/core/isolation"
	"github.com/liuhailove/seamiter-golang/core/log"
	"github.com/liuhailove/seamiter-golang/core/mock"
	"github.com/liuhailove/seamiter-golang/core/stat"
	"github.com/liuhailove/seamiter-golang/core/system"
)

var globalSlotChain = BuildDefaultSlotChain()

func GlobalSlotChain() *base.SlotChain {
	return globalSlotChain
}

func BuildDefaultSlotChain() *base.SlotChain {
	sc := base.NewSlotChain()
	sc.AddStatPrepareSlot(stat.DefaultResourceNodePrepareSlot)

	sc.AddRuleCheckSlot(system.DefaultAdaptiveSlot)
	sc.AddRuleCheckSlot(flow.DefaultSlot)
	sc.AddRuleCheckSlot(isolation.DefaultSlot)
	sc.AddRuleCheckSlot(hotspot.DefaultSlot)
	sc.AddRuleCheckSlot(circuitbreaker.DefaultSlot)
	// 数据Mock Check
	sc.AddRuleCheckSlot(mock.DefaultSlot)

	sc.AddStatSlot(stat.DefaultSlot)
	sc.AddStatSlot(log.DefaultSlot)
	sc.AddStatSlot(flow.DefaultStandaloneStatSlot)
	sc.AddStatSlot(hotspot.DefaultConcurrencyStatSlot)
	sc.AddStatSlot(circuitbreaker.DefaultMetricStatSlot)

	// 增加灰度路由策略
	sc.AddRouterSlot(gray.DefaultSlot)
	return sc
}
