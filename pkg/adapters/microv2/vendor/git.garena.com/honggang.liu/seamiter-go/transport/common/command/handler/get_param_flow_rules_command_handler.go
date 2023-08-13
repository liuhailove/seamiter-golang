package handler

import (
	"git.garena.com/honggang.liu/seamiter-go/core/hotspot"
	"git.garena.com/honggang.liu/seamiter-go/ext/datasource"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/transport/common/command"
)

var (
	getParamFlowRulesCommandHandlerInst = new(getParamFlowRulesCommandHandler)
)

func init() {
	command.RegisterHandler(getParamFlowRulesCommandHandlerInst.Name(), getParamFlowRulesCommandHandlerInst)
}

// getParamFlowRulesCommandHandler 获取热点参数限流规则
type getParamFlowRulesCommandHandler struct {
}

func (g getParamFlowRulesCommandHandler) Name() string {
	return "getParamFlowRules"
}

func (g getParamFlowRulesCommandHandler) Desc() string {
	return "Get all parameter flow rules"
}

func (g getParamFlowRulesCommandHandler) Handle(request command.Request) *command.Response {
	rules := hotspot.GetRules()
	rulesBytes, err := datasource.HotSpotParamRuleTrans(rules)
	if err != nil {
		logging.Error(err, "[getParamFlowRulesCommandHandler] handler error")
		return command.OfFailure(err)
	}
	return command.OfSuccess(string(rulesBytes))
}
