package handler

import "git.garena.com/honggang.liu/seamiter-go/transport/common/command"

var (
	onOffSetCommandHandlerInst = new(onOffSetCommandHandler)
)

func init() {
	command.RegisterHandler(onOffSetCommandHandlerInst.Name(), onOffSetCommandHandlerInst)
}

type onOffSetCommandHandler struct {
}

func (o onOffSetCommandHandler) Name() string {
	return "setSwitch"
}

func (o onOffSetCommandHandler) Desc() string {
	return "set sea switch, accept param: value={true|false}"
}

func (o onOffSetCommandHandler) Handle(request command.Request) *command.Response {
	return command.OfSuccess("")
}
