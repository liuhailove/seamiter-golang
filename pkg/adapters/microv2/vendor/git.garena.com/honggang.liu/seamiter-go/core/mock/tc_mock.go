package mock

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
)

type mockTrafficShapingController struct {
	baseTrafficShapingController
}

func (m *mockTrafficShapingController) PerformCheckingFunc(ctx *base.EntryContext) *base.TokenResult {
	//if !m.HeadersCheck(ctx) {
	//	return nil
	//}
	//if !m.ContextCheck(ctx) {
	//	return nil
	//}
	return base.NewTokenResultBlockedWithCause(base.BlockTypeMock, "", m.BoundRule(), m.BoundRule().ThenReturnMockData)
}

// PerformCheckingArgs 执行参数检查
func (m *mockTrafficShapingController) PerformCheckingArgs(ctx *base.EntryContext) *base.TokenResult {
	item := m.ArgsCheck(ctx)
	if item == nil {
		return nil
	}
	if item.MockReplace == Req {
		return base.NewTokenResultBlockedWithCause(base.BlockTypeMockRequest, "", m.BoundRule(), item.TmpData)
	}
	return base.NewTokenResultBlockedWithCause(base.BlockTypeMock, "", m.BoundRule(), item.ThenReturnMockData)
}
