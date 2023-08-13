package handler

import (
	"github.com/liuhailove/seamiter-golang/transport/common/command"
	"github.com/liuhailove/seamiter-golang/util"
)

var (
	basicInfoCommandHandlerInst = new(basicInfoCommandHandler)
)

func init() {
	command.RegisterHandler(basicInfoCommandHandlerInst.Name(), basicInfoCommandHandlerInst)
}

type basicInfoCommandHandler struct {
}

func (b basicInfoCommandHandler) Name() string {
	return "basicInfo"
}

func (b basicInfoCommandHandler) Desc() string {
	return "get sea config info"
}

func (b basicInfoCommandHandler) Handle(request command.Request) *command.Response {
	return command.OfSuccess(util.GetConfigString())
}
