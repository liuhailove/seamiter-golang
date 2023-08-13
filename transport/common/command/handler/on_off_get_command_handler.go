package handler

import "github.com/liuhailove/seamiter-golang/transport/common/command"

var (
	onOffGetCommandHandlerInst = new(onOffGetCommandHandler)
)

func init() {
	command.RegisterHandler(onOffGetCommandHandlerInst.Name(), onOffGetCommandHandlerInst)
}

type onOffGetCommandHandler struct {
}

func (o onOffGetCommandHandler) Name() string {
	return "getSwitch"
}

func (o onOffGetCommandHandler) Desc() string {
	return "get sea switch status"
}

func (o onOffGetCommandHandler) Handle(request command.Request) *command.Response {
	return command.OfSuccess("")
}
