package command

var (
	handlerMap = make(map[string]Handler)
)

// commandHandlerProvider Provides and filters command handlers registered via SPI.
type commandHandlerProvider struct {
}

func RegisterHandler(name string, handler Handler) {
	handlerMap[name] = handler
}

// NamedHandlers Get all command handlers annotated with {@link CommandMapping} with command name.
func (c commandHandlerProvider) NamedHandlers() map[string]Handler {
	return handlerMap
}

var _commandHandlerProviderInst commandHandlerProvider

// ProviderInst 单例获取
func ProviderInst() *commandHandlerProvider {
	return &_commandHandlerProviderInst
}
