package gray

import (
	"github.com/liuhailove/seamiter-golang/core/base"
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/liuhailove/seamiter-golang/util"
	"math"
	"strings"
)

const (
	// RuleCheckSlotOrder 灰度策略需要在最后一个，因为执行完此结果就返回了
	RuleCheckSlotOrder = math.MaxInt32
)

var (
	DefaultSlot = &Slot{}
)

type Slot struct {
}

func (s Slot) Order() uint32 {
	return RuleCheckSlotOrder
}

// Initial
//
// 初始化，如果有初始化工作放入其中
func (s Slot) Initial() {
}

func (s Slot) Router(ctx *base.EntryContext) *base.TokenResult {
	res := ctx.Resource.Name()
	tcs := getTrafficControllerListFor(res)
	result := ctx.RuleCheckResult

	// 按序检查规则
	for _, tc := range tcs {
		if tc == nil {
			logging.Warn("[GraySlot Check]Nil traffic controller found", "resourceName", res)
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
		r := tc.PerformSelecting(ctx)
		if r == nil {
			continue
		}
		return r

	}
	return result
}
