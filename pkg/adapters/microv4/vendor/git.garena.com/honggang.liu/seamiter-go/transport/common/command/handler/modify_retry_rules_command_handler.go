package handler

import (
	"fmt"
	retry "git.garena.com/honggang.liu/seamiter-go/core/retry/rule"
	"git.garena.com/honggang.liu/seamiter-go/ext/datasource"
	"git.garena.com/honggang.liu/seamiter-go/ext/datasource/util"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/transport/common/command"
)

var (
	modifyRetryRulesCommandHandlerInst = new(modifyRetryRulesCommandHandler)
)

func init() {
	command.RegisterHandler(modifyRetryRulesCommandHandlerInst.Name(), modifyRetryRulesCommandHandlerInst)
}

// modifyRetryRulesCommandHandler 更新重试规则handler
type modifyRetryRulesCommandHandler struct {
}

func (m *modifyRetryRulesCommandHandler) Name() string {
	return "setRetryRules"
}

func (m *modifyRetryRulesCommandHandler) Desc() string {
	return "Set retry rules, while previous rules will be replaced."
}

func (m *modifyRetryRulesCommandHandler) Handle(request command.Request) *command.Response {
	// rule data in get parameter
	var data = request.GetParam("data")
	logging.Info("Receiving retry rule change", "data", data)
	var mockRules []*retry.Rule
	var ok bool

	retryRulesInf, err := datasource.RetryRuleJsonArrayParser([]byte(data))
	if err != nil {
		logging.Warn("[modifyRetryRulesCommandHandler] unmarshall error", "data", data, "err", err)
		return command.OfFailure(err)
	}
	if mockRules, ok = retryRulesInf.([]*retry.Rule); !ok {
		logging.Warn("[modifyRetryRulesCommandHandler] assert to []*retry.Rule", "data", data)
		err = fmt.Errorf("[modifyRetryRulesCommandHandler] assert to []*retry.Rule")
		return command.OfFailure(err)
	}

	err = datasource.RetryRulesUpdater(mockRules)
	if err != nil {
		logging.Warn("[modifyRetryRulesCommandHandler] RetryRulesUpdater error", "data", data, "err", err)
		return command.OfFailure(err)
	}
	var result = "success"
	if !m.writeToDataSource(util.GetRetrySource(), []byte(data)) {
		result = WriteDsFailureMsg
	}
	return command.OfSuccess(result)
}
func (m *modifyRetryRulesCommandHandler) writeToDataSource(source datasource.DataSource, data []byte) bool {
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
