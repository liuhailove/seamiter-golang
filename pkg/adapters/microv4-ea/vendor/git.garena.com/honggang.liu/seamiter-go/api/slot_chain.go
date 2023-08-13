package api

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/core/circuitbreaker"
	"git.garena.com/honggang.liu/seamiter-go/core/flow"
	"git.garena.com/honggang.liu/seamiter-go/core/gray"
	"git.garena.com/honggang.liu/seamiter-go/core/hotspot"
	"git.garena.com/honggang.liu/seamiter-go/core/isolation"
	"git.garena.com/honggang.liu/seamiter-go/core/log"
	"git.garena.com/honggang.liu/seamiter-go/core/mock"
	"git.garena.com/honggang.liu/seamiter-go/core/stat"
	"git.garena.com/honggang.liu/seamiter-go/core/system"
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
