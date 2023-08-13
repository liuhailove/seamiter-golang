package handler

import (
	"errors"
	"git.garena.com/honggang.liu/seamiter-go/core/system"
	"git.garena.com/honggang.liu/seamiter-go/ext/datasource"
	"git.garena.com/honggang.liu/seamiter-go/ext/datasource/util"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/transport/common/command"
	"strings"
)

const (
	WriteDsFailureMsg = "partial success (write data source failed)"
	FlowRuleType      = "flow"
	DegradeRuleType   = "degrade"
	SystemRuleType    = "system"
	AuthorityRuleType = "authority"
)

var (
	modifyRulesCommandHandlerInst = new(modifyRulesCommandHandler)
)

func init() {
	command.RegisterHandler(modifyRulesCommandHandlerInst.Name(), modifyRulesCommandHandlerInst)
}

type modifyRulesCommandHandler struct {
}

func (m modifyRulesCommandHandler) Name() string {
	return "setRules"
}

func (m modifyRulesCommandHandler) Desc() string {
	return "modify the rules, accept param: type={ruleType}&data={ruleJson}"
}

func (m modifyRulesCommandHandler) Handle(request command.Request) *command.Response {
	var typ = request.GetParam("type")
	// rule data in get parameter
	var data = request.GetParam("data")
	logging.Info("Receiving rule change", "type", typ, "data", data)
	var result = "success"
	if strings.EqualFold(FlowRuleType, typ) {
		rules, err := datasource.FlowRuleJsonArrayParser([]byte(data))
		if err != nil {
			logging.Warn("[modifyRulesCommandHandler] unmarshall error", "data", data, "err", err)
			return command.OfFailure(err)
		}
		err = datasource.FlowRulesUpdater(rules)
		if err != nil {
			logging.Warn("[modifyParamFlowRulesCommandHandler] FlowRulesUpdater error", "data", data, "err", err)
			return command.OfFailure(err)
		}
		if !m.writeToDataSource(util.GetFlowDataSource(), []byte(data)) {
			result = WriteDsFailureMsg
		}
		return command.OfSuccess(result)
	} else if strings.EqualFold(AuthorityRuleType, typ) {
		// TODO
		return command.OfSuccess(result)
	} else if strings.EqualFold(DegradeRuleType, typ) {
		rules, err := datasource.CircuitBreakerRuleJsonArrayParser([]byte(data))
		if err != nil {
			logging.Warn("[modifyRulesCommandHandler] unmarshall error", "data", data, "err", err)
			return command.OfFailure(err)
		}
		err = datasource.CircuitBreakerRulesUpdater(rules)
		if err != nil {
			logging.Warn("[modifyParamFlowRulesCommandHandler] CircuitBreakerRulesUpdater error", "data", data, "err", err)
			return command.OfFailure(err)
		}
		if !m.writeToDataSource(util.GetDegradeDataSource(), []byte(data)) {
			result = WriteDsFailureMsg
		}
		return command.OfSuccess(result)
	} else if strings.EqualFold(SystemRuleType, typ) {
		systemRulesInf, err := datasource.SystemRuleJsonArrayParser([]byte(data))
		if err != nil {
			logging.Warn("[modifyRulesCommandHandler] unmarshall error", "data", data, "err", err)
			return command.OfFailure(err)
		}
		var systemRules = systemRulesInf.([]*system.Rule)
		err = datasource.SystemRulesUpdater(systemRules)
		if err != nil {
			logging.Warn("[modifyParamFlowRulesCommandHandler] SystemRulesUpdater error", "data", data, "err", err)
			return command.OfFailure(err)
		}
		var result = "success"
		if !m.writeToDataSource(util.GetSystemSource(), []byte(data)) {
			result = WriteDsFailureMsg
		}
		return command.OfSuccess(result)
	}
	return command.OfFailure(errors.New("invalid type"))
}

func (m modifyRulesCommandHandler) writeToDataSource(source datasource.DataSource, data []byte) bool {
	if source != nil {
		err := source.Write(data)
		if err != nil {
			logging.Warn("Write data source failed", "err", err)
			return false
		}
		return true
	}
	return true
}
