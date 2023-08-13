package mock

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
)

type mockTrafficShapingController struct {
	baseTrafficShapingController
}

func (m *mockTrafficShapingController) PerformCheckingFunc(_ *base.EntryContext) *base.TokenResult {
	return base.NewTokenResultBlockedWithCause(base.BlockTypeMock, "", m.BoundRule(), m.BoundRule().ThenReturnMockData)
}

func (m *mockTrafficShapingController) PerformCheckingArgs(ctx *base.EntryContext) *base.TokenResult {
	item := m.ArgsCheck(ctx)
	if item != nil {
		return base.NewTokenResultBlockedWithCause(base.BlockTypeMock, "", m.BoundRule(), item.ThenReturnMockData)
	}
	return nil
}
