package handler

import (
	"errors"
	"git.garena.com/honggang.liu/seamiter-go/core/flow"
	"git.garena.com/honggang.liu/seamiter-go/core/system"
	"git.garena.com/honggang.liu/seamiter-go/ext/datasource"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/transport/common/command"
	jsoniter "github.com/json-iterator/go"
	"strings"
)

var (
	fetchActiveRuleCommandHandlerInst = new(fetchActiveRuleCommandHandler)
)

func init() {
	command.RegisterHandler(fetchActiveRuleCommandHandlerInst.Name(), fetchActiveRuleCommandHandlerInst)
}

// fetchActiveRuleCommandHandler 抓取活跃的规则Handler
type fetchActiveRuleCommandHandler struct {
}

func (f fetchActiveRuleCommandHandler) Name() string {
	return "getRules"
}

func (f fetchActiveRuleCommandHandler) Desc() string {
	return "get all active rules by type, request param: type={ruleType}"
}

func (f fetchActiveRuleCommandHandler) Handle(request command.Request) *command.Response {
	typ := request.GetParam("type")
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if strings.EqualFold("flow", typ) {
		rules := flow.GetRules()
		rulesBytes, _ := json.Marshal(rules)
		return command.OfSuccess(string(rulesBytes))
	} else if strings.EqualFold("degrade", typ) {
		rules := datasource.GetCircuitBreakerRules()
		rulesBytes, _ := json.Marshal(rules)
		return command.OfSuccess(string(rulesBytes))
	} else if strings.EqualFold("authority", typ) {
		return nil
	} else if strings.EqualFold("system", typ) {
		data, err := datasource.SystemRuleTrans(system.GetRules())
		if err != nil {
			logging.Warn("[fetchActiveRuleCommandHandler] SystemRuleTrans error", "data", data, "err", err)
			return command.OfFailure(err)
		}
		return command.OfSuccess(string(data))
	} else {
		return command.OfFailure(errors.New("invalid type"))
	}
}
