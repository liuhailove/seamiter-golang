package handler

import "git.garena.com/honggang.liu/seamiter-go/transport/common/command"

var (
	fetchOriginCommandHandlerInst = new(fetchOriginCommandHandler)
)

func init() {
	command.RegisterHandler(fetchOriginCommandHandlerInst.Name(), fetchOriginCommandHandlerInst)
}

type fetchOriginCommandHandler struct {
}

func (f fetchOriginCommandHandler) Name() string {
	return "origin"
}

func (f fetchOriginCommandHandler) Desc() string {
	return "get origin clusterNode by id, request param: id={resourceName}"
}

func (f fetchOriginCommandHandler) Handle(request command.Request) *command.Response {
	return command.OfSuccess("")
}
