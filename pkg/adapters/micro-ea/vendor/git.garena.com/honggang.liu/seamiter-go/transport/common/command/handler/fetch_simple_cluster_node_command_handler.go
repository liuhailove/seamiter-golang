package handler

import "git.garena.com/honggang.liu/seamiter-go/transport/common/command"

var (
	fetchSimpleClusterNodeCommandHandlerInst = new(fetchSimpleClusterNodeCommandHandler)
)

func init() {
	command.RegisterHandler(fetchSimpleClusterNodeCommandHandlerInst.Name(), fetchSimpleClusterNodeCommandHandlerInst)
}

type fetchSimpleClusterNodeCommandHandler struct {
}

func (f fetchSimpleClusterNodeCommandHandler) Name() string {
	return "clusterNode"
}

func (f fetchSimpleClusterNodeCommandHandler) Desc() string {
	return "get all clusterNode VO, use type=notZero to ignore those nodes with totalRequest <=0"
}

func (f fetchSimpleClusterNodeCommandHandler) Handle(request command.Request) *command.Response {
	return command.OfSuccess("")
}
