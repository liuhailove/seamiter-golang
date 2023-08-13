package handler

import (
	"fmt"
	"git.garena.com/honggang.liu/seamiter-go/core/gray"
	"git.garena.com/honggang.liu/seamiter-go/ext/datasource"
	"git.garena.com/honggang.liu/seamiter-go/ext/datasource/util"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/transport/common/command"
)

var (
	modifyGrayRulesCommandHandlerInst = new(modifyGrayRulesCommandHandler)
)

func init() {
	command.RegisterHandler(modifyGrayRulesCommandHandlerInst.Name(), modifyGrayRulesCommandHandlerInst)
}

// modifyGrayRulesCommandHandler 更新灰度规则
type modifyGrayRulesCommandHandler struct {
}

func (m modifyGrayRulesCommandHandler) Name() string {
	return "setGrayRules"
}

func (m modifyGrayRulesCommandHandler) Desc() string {
	return "Set gray rules, while previous rules will be replaced."
}

func (m modifyGrayRulesCommandHandler) Handle(request command.Request) *command.Response {
	// rule data in get parameter
	var data = request.GetParam("data")
	logging.Info("Receiving rule change", "data", data)
	var grayRules []*gray.Rule
	var ok bool
	grayRulesInf, err := datasource.GrayRuleJsonArrayParser([]byte(data))
	if err != nil {
		logging.Warn("[modifyGrayRulesCommandHandler] unmarshall error", "data", data, "err", err)
		return command.OfFailure(err)
	}
	if grayRules, ok = grayRulesInf.([]*gray.Rule); !ok {
		logging.Warn("[modifyGrayRulesCommandHandler] assert to []*gray.Rule", "data", data)
		err = fmt.Errorf("[modifyGrayRulesCommandHandler] assert to []*gray.Rule")
		return command.OfFailure(err)
	}
	err = datasource.GrayRulesUpdater(grayRules)
	if err != nil {
		logging.Warn("[modifyGrayRulesCommandHandler] GrayRulesUpdater error", "data", data, "err", err)
		return command.OfFailure(err)
	}
	var result = "success"
	if !m.writeToDataSource(util.GetGraySource(), []byte(data)) {
		result = WriteDsFailureMsg
	}
	return command.OfSuccess(result)
}
func (m modifyGrayRulesCommandHandler) writeToDataSource(source datasource.DataSource, data []byte) bool {
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
