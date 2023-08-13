package mock

import (
	"github.com/liuhailove/seamiter-golang/core/base"
	"github.com/liuhailove/seamiter-golang/util"
	"strconv"
	"strings"
	"time"
)

type waitingThenMockTrafficShapingController struct {
	baseTrafficShapingController
}

func (m *waitingThenMockTrafficShapingController) PerformCheckingFunc(ctx *base.EntryContext) *base.TokenResult {
	if nanosToWait := m.r.ThenReturnWaitingTimeMs * time.Millisecond.Nanoseconds(); nanosToWait > 0 {
		// Handle waiting action.
		util.Sleep(time.Duration(nanosToWait))
	}
	var thenReturnMockData = m.BoundRule().ThenReturnMockData
	// 先替换，无论是否匹配，都可以先替换
	// nano方法替换
	thenReturnMockData = strings.ReplaceAll(thenReturnMockData, TimeNanoFunc, strconv.FormatInt(time.Now().UnixNano(), 10))
	// 毫秒方法替换
	thenReturnMockData = strings.ReplaceAll(thenReturnMockData, TimeMillisFunc, strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	// 秒方法替换
	thenReturnMockData = strings.ReplaceAll(thenReturnMockData, TimeSecFunc, strconv.FormatInt(time.Now().Unix(), 10))
	return base.NewTokenResultBlockedWithCause(base.BlockTypeMock, "", m.BoundRule(), thenReturnMockData)
}

func (m *waitingThenMockTrafficShapingController) PerformCheckingArgs(ctx *base.EntryContext) *base.TokenResult {
	return m.DoInnerCheck(ctx)
}
