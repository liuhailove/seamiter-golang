package flow

import (
	"github.com/liuhailove/seamiter-golang/core/base"
	"github.com/liuhailove/seamiter-golang/core/stat"
	metric_exporter "github.com/liuhailove/seamiter-golang/exporter/metric"
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/liuhailove/seamiter-golang/util"
	"github.com/pkg/errors"
	"strings"
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
	isInitialized util.AtomicBool
}

func (s Slot) Order() uint32 {
	return RuleCheckSlotOrder
}

// Initial
//
// 初始化，如果有初始化工作放入其中
func (s Slot) Initial() {
}

func (s Slot) Check(ctx *base.EntryContext) *base.TokenResult {

	res := ctx.Resource.Name()
	tcs := getTrafficControllerListFor(res)
	result := ctx.RuleCheckResult

	// 按序检查规则
	for _, tc := range tcs {
		if tc == nil {
			logging.Warn("[FlowSlot Check]Nil traffic controller found", "resourceName", res)
			continue
		}
		// 来源检查
		var needContinueCheck = true
		if tc.BoundRule().LimitApp != "" && !strings.EqualFold(tc.BoundRule().LimitApp, "default") {
			if !util.Contains(ctx.FromService, strings.Split(tc.BoundRule().LimitApp, ",")) {
				needContinueCheck = false
			}
		}
		if !needContinueCheck {
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
				if flowWaitCount != nil {
					flowWaitCount.Add(float64(ctx.Input.BatchCount), ctx.Resource.Name())
				}
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
