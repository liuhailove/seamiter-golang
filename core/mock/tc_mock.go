package mock

import (
	"github.com/liuhailove/seamiter-golang/core/base"
	"strconv"
	"strings"
	"time"
)

type mockTrafficShapingController struct {
	baseTrafficShapingController
}

func (m *mockTrafficShapingController) PerformCheckingFunc(ctx *base.EntryContext) *base.TokenResult {
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

// PerformCheckingArgs 执行参数检查
func (m *mockTrafficShapingController) PerformCheckingArgs(ctx *base.EntryContext) *base.TokenResult {
	return m.DoInnerCheck(ctx)
}
