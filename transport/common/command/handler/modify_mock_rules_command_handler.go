package handler

import (
	"fmt"
	"github.com/liuhailove/seamiter-golang/core/mock"
	"github.com/liuhailove/seamiter-golang/ext/datasource"
	"github.com/liuhailove/seamiter-golang/ext/datasource/util"
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/liuhailove/seamiter-golang/transport/common/command"
)

var (
	modifyMockRulesCommandHandlerInst = new(modifyMockRulesCommandHandler)
)

func init() {
	command.RegisterHandler(modifyMockRulesCommandHandlerInst.Name(), modifyMockRulesCommandHandlerInst)
}

// modifyMockRulesCommandHandler 更新Mock规则
type modifyMockRulesCommandHandler struct {
}

func (m modifyMockRulesCommandHandler) Name() string {
	return "setMockRules"
}

func (m modifyMockRulesCommandHandler) Desc() string {
	return "Set mock rules, while previous rules will be replaced."
}

func (m modifyMockRulesCommandHandler) Handle(request command.Request) *command.Response {
	// rule data in get parameter
	var data = request.GetParam("data")
	logging.Info("Receiving rule change", "data", data)
	var mockRules []*mock.Rule
	var ok bool
	mockRulesInf, err := datasource.MockRuleJsonArrayParser([]byte(data))
	if err != nil {
		logging.Warn("[modifyMockRulesCommandHandler] unmarshall error", "data", data, "err", err)
		return command.OfFailure(err)
	}

	if mockRules, ok = mockRulesInf.([]*mock.Rule); !ok {
		logging.Warn("[modifyMockRulesCommandHandler] unmarshall error", "data", data, "err", err)
		err = fmt.Errorf("[modifyMockRulesCommandHandler] assert to []*mock.Rule")
		return command.OfFailure(err)
	}
	err = datasource.MockRulesUpdater(mockRules)
	if err != nil {
		logging.Warn("[modifyParamFlowRulesCommandHandler] HotSpotParamRulesUpdater error", "data", data, "err", err)
		return command.OfFailure(err)
	}
	var result = "success"
	if !m.writeToDataSource(util.GetMockSource(), []byte(data)) {
		result = WriteDsFailureMsg
	}
	return command.OfSuccess(result)
}
func (m modifyMockRulesCommandHandler) writeToDataSource(source datasource.DataSource, data []byte) bool {
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
