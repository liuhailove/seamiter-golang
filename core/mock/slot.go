package mock

import (
	"github.com/liuhailove/seamiter-golang/core/base"
	metric_exporter "github.com/liuhailove/seamiter-golang/exporter/metric"
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/liuhailove/seamiter-golang/util"
	"strings"
)

const (
	RuleCheckSlotOrder = 5000
)

var (
	DefaultSlot = &Slot{}
	mockCounter = metric_exporter.NewCounter(
		"mock_total",
		"mock count",
		[]string{"resource"})
)

func init() {
	metric_exporter.Register(mockCounter)
}

type Slot struct {
}

func (s *Slot) Order() uint32 {
	return RuleCheckSlotOrder
}

// Initial
//
// 初始化，如果有初始化工作放入其中
func (s *Slot) Initial() {
}

func (s *Slot) Check(ctx *base.EntryContext) *base.TokenResult {
	res := ctx.Resource.Name()
	result := ctx.RuleCheckResult
	if len(res) == 0 {
		return result
	}
	if logging.DebugEnabled() {
		// mock命中，打印请求和mock数据
		logging.Debug("in mock check", "resourceName", res, "request", ctx.Input.Args)
	}
	tcs := getTrafficControllersFor(res)
	var r *base.TokenResult
	var cache = false
	for _, tc := range tcs {
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
		if !cache {
			cacheRequest(tc.BoundRule(), ctx)
			cache = true
		}
		if logging.InfoEnabled() {
			logging.Info("in mock check", "resourceName", res, "metadata", ctx.Input.MetaData, "header", ctx.Input.Headers, "request", ctx.Input.Args)
		}
		if !tc.MockCheck(ctx) {
			continue
		}
		if tc.BoundRule().Strategy == Func {
			r = tc.PerformCheckingFunc(ctx)
		} else {
			r = tc.PerformCheckingArgs(ctx)
		}
		if r == nil {
			continue
		}
		// mock命中，打印请求和mock数据
		logging.Info("mock hits", "resourceName", res, "request", ctx.Input.Args, "mockData", r.String())
		mockCounter.Add(1, ctx.Resource.Name())
		return r
	}
	return result
}
