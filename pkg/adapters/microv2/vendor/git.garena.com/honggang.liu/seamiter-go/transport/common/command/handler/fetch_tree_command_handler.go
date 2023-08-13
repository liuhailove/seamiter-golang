package handler

import "git.garena.com/honggang.liu/seamiter-go/transport/common/command"

var (
	fetchTreeCommandHandlerInst = new(fetchTreeCommandHandler)
)

func init() {
	command.RegisterHandler(fetchTreeCommandHandlerInst.Name(), fetchTreeCommandHandlerInst)
}

type fetchTreeCommandHandler struct {
}

func (f fetchTreeCommandHandler) Name() string {
	return "tree"
}

func (f fetchTreeCommandHandler) Desc() string {
	return "get metrics in tree mode, use id to specify detailed tree root"
}

func (f fetchTreeCommandHandler) Handle(request command.Request) *command.Response {
	return command.OfSuccess("")
}
