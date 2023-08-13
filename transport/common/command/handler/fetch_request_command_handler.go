package handler

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/liuhailove/seamiter-golang/core/mock"
	"github.com/liuhailove/seamiter-golang/ext/datasource"
	"github.com/liuhailove/seamiter-golang/transport/common/command"
)

var (
	fetchRequestCommandHandlerInst = new(fetchRequestCommandHandler)
)

func init() {
	command.RegisterHandler(fetchRequestCommandHandlerInst.Name(), fetchRequestCommandHandlerInst)
}

// fetchRequestCommandHandler 抓取临时Request，这个request是一次请求的临时存储
type fetchRequestCommandHandler struct {
}

func (f *fetchRequestCommandHandler) Name() string {
	return "fetchRequest"
}

func (f *fetchRequestCommandHandler) Desc() string {
	return "get tmp request"
}

func (f *fetchRequestCommandHandler) Handle(request command.Request) *command.Response {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if data, err := json.Marshal(mock.NeedReportRequest()); err != nil {
		desc := fmt.Sprintf("Fail to trans request to bytes, err: %s", err.Error())
		return command.OfFailure(datasource.NewError(datasource.ConvertSourceError, desc))
	} else {
		return command.OfSuccess(string(data))
	}
}
