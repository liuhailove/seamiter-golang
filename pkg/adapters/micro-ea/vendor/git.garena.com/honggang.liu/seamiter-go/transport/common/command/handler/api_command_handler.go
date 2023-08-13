package handler

import (
	"git.garena.com/honggang.liu/seamiter-go/transport/common/command"
	jsoniter "github.com/json-iterator/go"
)

var (
	apiCommandHandlerInst = new(apiCommandHandler)
)

func init() {
	command.RegisterHandler(apiCommandHandlerInst.Name(), apiCommandHandlerInst)
}

//  List all available command handlers by request:
// {@code curl http://ip:commandPort/api}
type apiCommandHandler struct {
}

func (a apiCommandHandler) Name() string {
	return "api"
}

func (a apiCommandHandler) Desc() string {
	return "get all available command handlers"
}

func (a apiCommandHandler) Handle(request command.Request) *command.Response {
	handlers := command.ProviderInst().NamedHandlers()
	if len(handlers) == 0 {
		return command.OfSuccess(`[]`)
	}
	var data []map[string]string
	for _, handler := range handlers {
		var item = make(map[string]string, 2)
		item["url"] = handler.Name()
		item["desc"] = handler.Desc()
		data = append(data, item)
	}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	bytes, _ := json.Marshal(data)
	return command.OfSuccess(string(bytes))
}
