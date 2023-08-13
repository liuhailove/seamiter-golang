package handler

import "git.garena.com/honggang.liu/seamiter-go/transport/common/command"

var (
	fetchSystemStatusCommandHandlerInst = new(fetchSystemStatusCommandHandler)
)

func init() {
	command.RegisterHandler(fetchSystemStatusCommandHandlerInst.Name(), fetchSystemStatusCommandHandlerInst)
}

type fetchSystemStatusCommandHandler struct {
}

func (f fetchSystemStatusCommandHandler) Name() string {
	return "systemStatus"
}

func (f fetchSystemStatusCommandHandler) Desc() string {
	return "get system status"
}

func (f fetchSystemStatusCommandHandler) Handle(request command.Request) *command.Response {
	return command.OfSuccess("")
}
