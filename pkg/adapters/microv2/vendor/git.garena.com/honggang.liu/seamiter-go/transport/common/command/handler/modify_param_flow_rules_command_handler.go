package handler

import (
	"git.garena.com/honggang.liu/seamiter-go/core/hotspot"
	"git.garena.com/honggang.liu/seamiter-go/ext/datasource"
	"git.garena.com/honggang.liu/seamiter-go/ext/datasource/util"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/transport/common/command"
)

var (
	modifyParamFlowRulesCommandHandlerInst = new(modifyParamFlowRulesCommandHandler)
)

func init() {
	command.RegisterHandler(modifyParamFlowRulesCommandHandlerInst.Name(), modifyParamFlowRulesCommandHandlerInst)
}

// modifyParamFlowRulesCommandHandler 更新热点参数限流规则
type modifyParamFlowRulesCommandHandler struct {
}

func (m modifyParamFlowRulesCommandHandler) Name() string {
	return "setParamFlowRules"
}

func (m modifyParamFlowRulesCommandHandler) Desc() string {
	return "Set parameter flow rules, while previous rules will be replaced."
}

func (m modifyParamFlowRulesCommandHandler) Handle(request command.Request) *command.Response {
	// rule data in get parameter
	var data = request.GetParam("data")
	logging.Info("Receiving rule change", "data", data)
	var hotspotRules []*hotspot.Rule
	hotspotRulesInf, err := datasource.HotSpotParamRuleJsonArrayParser([]byte(data))
	if err != nil {
		logging.Warn("[modifyParamFlowRulesCommandHandler] unmarshall error", "data", data, "err", err)
		return command.OfFailure(err)
	}
	hotspotRules = hotspotRulesInf.([]*hotspot.Rule)
	err = datasource.HotSpotParamRulesUpdater(hotspotRules)
	if err != nil {
		logging.Warn("[modifyParamFlowRulesCommandHandler] HotSpotParamRulesUpdater error", "data", data, "err", err)
		return command.OfFailure(err)
	}
	var result = "success"
	if !m.writeToDataSource(util.GetFlowDataSource(), []byte(data)) {
		result = WriteDsFailureMsg
	}
	return command.OfSuccess(result)
}
func (m modifyParamFlowRulesCommandHandler) writeToDataSource(source datasource.DataSource, data []byte) bool {
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
