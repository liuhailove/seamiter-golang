package mock

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
)

type mockTrafficShapingController struct {
	baseTrafficShapingController
}

func (m *mockTrafficShapingController) PerformCheckingFunc(ctx *base.EntryContext) *base.TokenResult {
	return base.NewTokenResultBlockedWithCause(base.BlockTypeMock, "", m.BoundRule(), m.BoundRule().ThenReturnMockData)
}

// PerformCheckingArgs 执行参数检查
func (m *mockTrafficShapingController) PerformCheckingArgs(ctx *base.EntryContext) *base.TokenResult {
	return m.DoInnerCheck(ctx)
}
