package handler

import (
	"errors"
	"fmt"
	"github.com/liuhailove/seamiter-golang/core/isolation"
	"github.com/liuhailove/seamiter-golang/core/system"
	"github.com/liuhailove/seamiter-golang/ext/datasource"
	"github.com/liuhailove/seamiter-golang/ext/datasource/util"
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/liuhailove/seamiter-golang/transport/common/command"
	"strings"
)

const (
	WriteDsFailureMsg = "partial success (write data source failed)"
	FlowRuleType      = "flow"
	DegradeRuleType   = "degrade"
	SystemRuleType    = "system"
	AuthorityRuleType = "authority"
	IsolationRuleType = "isolation"
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
		var systemRules []*system.Rule
		var ok bool
		if systemRules, ok = systemRulesInf.([]*system.Rule); !ok {
			logging.Warn("[modifyParamFlowRulesCommandHandler] assert to SystemRulesUpdater error", "data", data)
			err = fmt.Errorf("[modifyParamFlowRulesCommandHandler] assert to SystemRulesUpdater error")
			return command.OfFailure(err)
		}
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
	} else if strings.EqualFold(IsolationRuleType, typ) {
		isolationRulesInf, err := datasource.IsolationRuleJsonArrayParser([]byte(data))
		if err != nil {
			logging.Warn("[modifyRulesCommandHandler] unmarshall error", "data", data, "err", err)
			return command.OfFailure(err)
		}
		var isolationRules []*isolation.Rule
		var ok bool
		if isolationRules, ok = isolationRulesInf.([]*isolation.Rule); !ok {
			logging.Warn("[modifyIsolationRulesCommandHandler] assert to IsolationRulesUpdater error", "data", data)
			err = fmt.Errorf("[modifyIsolationRulesCommandHandler] assert to IsolationRulesUpdater error")
			return command.OfFailure(err)
		}
		err = datasource.IsolationRulesUpdater(isolationRules)
		if err != nil {
			logging.Warn("[modifyIsolationRulesCommandHandler] IsolationRulesUpdater error", "data", data, "err", err)
			return command.OfFailure(err)
		}
		var result = "success"
		if !m.writeToDataSource(util.GetIsolationSource(), []byte(data)) {
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
