package handler

import (
	"git.garena.com/honggang.liu/seamiter-go/core/log/metric"
	"git.garena.com/honggang.liu/seamiter-go/ext/datasource"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/transport/common/command"
)

var (
	fetchJsonTreeCommandHandlerInst = new(fetchJsonTreeCommandHandler)
)

func init() {
	command.RegisterHandler(fetchJsonTreeCommandHandlerInst.Name(), fetchJsonTreeCommandHandlerInst)
}

// fetchJsonTreeCommandHandler 抓取json tree
type fetchJsonTreeCommandHandler struct {
}

func (f fetchJsonTreeCommandHandler) Name() string {
	return "jsonTree"
}

func (f fetchJsonTreeCommandHandler) Desc() string {
	return "get tree node VO start from root node"
}

func (f fetchJsonTreeCommandHandler) Handle(request command.Request) *command.Response {
	data, err := datasource.NodeStatTrans(metric.CurrentMetricItems())
	if err != nil {
		logging.Warn("[fetchJsonTreeCommandHandler] NodeStatTrans error", "data", data, "err", err)
		return command.OfFailure(err)
	}
	return command.OfSuccess(string(data))
}
