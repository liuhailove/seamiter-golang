package handler

import (
	"errors"
	jsoniter "github.com/json-iterator/go"
	"github.com/liuhailove/seamiter-golang/core/flow"
	"github.com/liuhailove/seamiter-golang/core/system"
	"github.com/liuhailove/seamiter-golang/ext/datasource"
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/liuhailove/seamiter-golang/transport/common/command"
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
