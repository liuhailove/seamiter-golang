package flow

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/core/stat"
	metric_exporter "git.garena.com/honggang.liu/seamiter-go/exporter/metric"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"github.com/pkg/errors"
)

const (
	RuleCheckSlotOrder = 2000
)

var (
	DefaultSlot   = &Slot{}
	flowWaitCount = metric_exporter.NewCounter(
		"flow_wait_total",
		"Flow wait count",
		[]string{"resource"})
)

func init() {
	metric_exporter.Register(flowWaitCount)
}

type Slot struct {
}

func (s Slot) Order() uint32 {
	return RuleCheckSlotOrder
}

func (s Slot) Check(ctx *base.EntryContext) *base.TokenResult {
	res := ctx.Resource.Name()
	tcs := getTrafficControllerListFor(res)
	result := ctx.RuleCheckResult

	// Check rules in order
	for _, tc := range tcs {
		if tc == nil {
			logging.Warn("[FlowSlot Check]Nil traffic controller found", "resourceName", res)
			continue
		}
		r := canPassCheck(tc, ctx.StatNode, ctx.Input.BatchCount)
		if r == nil {
			// nil means pass
			continue
		}
		if r.Status() == base.ResultStatusBlocked {
			return r
		}
		if r.Status() == base.ResultStatusShouldWait {
			if nanosToWait := r.NanosToWait(); nanosToWait > 0 {
				flowWaitCount.Add(float64(ctx.Input.BatchCount), ctx.Resource.Name())
				// Handle waiting action.
				util.Sleep(nanosToWait)
			}
			continue
		}
	}
	return result
}

func canPassCheck(tc *TrafficShapingController, node base.StatNode, batchCount uint32) *base.TokenResult {
	return canPassCheckWithFlag(tc, node, batchCount, 0)
}

func canPassCheckWithFlag(tc *TrafficShapingController, node base.StatNode, batchCount uint32, flag int32) *base.TokenResult {
	return checkInLocal(tc, node, batchCount, flag)
}

func selectNodeByRelStrategy(rule *Rule, node base.StatNode) base.StatNode {
	if rule.RelationStrategy == AssociatedResource {
		return stat.GetResourceNode(rule.RefResource)
	}
	return node
}

func checkInLocal(tc *TrafficShapingController, resStat base.StatNode, batchCount uint32, flag int32) *base.TokenResult {
	actual := selectNodeByRelStrategy(tc.rule, resStat)
	if actual == nil {
		logging.FrequentErrorOnce.Do(func() {
			logging.Error(errors.Errorf("nil resource node"), "No resource node for flow rule in FlowSlot.checkInLocal()", "rule", tc.rule)
		})
		return base.NewTokenResultPass()
	}
	return tc.PerformChecking(actual, batchCount, flag)
}
