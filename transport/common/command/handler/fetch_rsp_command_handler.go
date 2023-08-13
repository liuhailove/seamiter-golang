package handler

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/liuhailove/seamiter-golang/core/base"
	"github.com/liuhailove/seamiter-golang/ext/datasource"
	"github.com/liuhailove/seamiter-golang/transport/common/command"
)

var (
	fetchRspCommandHandlerInst = new(fetchRspCommandHandler)
)

func init() {
	command.RegisterHandler(fetchRspCommandHandlerInst.Name(), fetchRspCommandHandlerInst)
}

// fetchRspCommandHandler 抓取临时Rsp，这个rsp是一次正确请求的临时存储
type fetchRspCommandHandler struct {
}

func (f *fetchRspCommandHandler) Name() string {
	return "fetchRsp"
}

func (f *fetchRspCommandHandler) Desc() string {
	return "get tmp rsp"
}

func (f *fetchRspCommandHandler) Handle(request command.Request) *command.Response {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if data, err := json.Marshal(base.ResourceNodeList()); err != nil {
		desc := fmt.Sprintf("Fail to trans rsp to bytes, err: %s", err.Error())
		return command.OfFailure(datasource.NewError(datasource.ConvertSourceError, desc))
	} else {
		return command.OfSuccess(string(data))
	}
}
