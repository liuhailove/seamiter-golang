package handler

import (
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/transport/common/command"
)

var (
	versionCommandHandlerInst = new(versionCommandHandler)
)

func init() {
	command.RegisterHandler(versionCommandHandlerInst.Name(), versionCommandHandlerInst)
}

type versionCommandHandler struct {
}

func (v versionCommandHandler) Name() string {
	return "version"
}

func (v versionCommandHandler) Desc() string {
	return "get sea version"
}

func (v versionCommandHandler) Handle(request command.Request) *command.Response {
	return command.OfSuccess(config.Version())
}
