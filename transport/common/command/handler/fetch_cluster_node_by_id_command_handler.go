package handler

import (
	"errors"
	"github.com/liuhailove/seamiter-golang/transport/common/command"
)

var (
	fetchClusterNodeByIdCommandHandlerInst = new(fetchClusterNodeByIdCommandHandler)
)

func init() {
	command.RegisterHandler(fetchClusterNodeByIdCommandHandlerInst.Name(), fetchClusterNodeByIdCommandHandlerInst)
}

type fetchClusterNodeByIdCommandHandler struct {
}

func (f fetchClusterNodeByIdCommandHandler) Name() string {
	return "clusterNodeById"
}

func (f fetchClusterNodeByIdCommandHandler) Desc() string {
	return "get clusterNode VO by id, request param: id={resourceName}"
}

func (f fetchClusterNodeByIdCommandHandler) Handle(request command.Request) *command.Response {
	id := request.GetParam("id")
	if id == "" {
		return command.OfFailure(errors.New("Invalid parameter: empty clusterNode name"))
	}

	return command.OfSuccess("")
}
