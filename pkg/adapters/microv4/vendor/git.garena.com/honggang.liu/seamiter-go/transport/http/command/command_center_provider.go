package command

import (
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/transport/common/transport"
)

var (
	commandCenter transport.CommandCenter
)

func init() {
	resolveInstance()
}
func resolveInstance() {
	resolveCommandCenter := new(SimpleHttpCommandCenter)
	if resolveCommandCenter == nil {
		logging.Warn("[CommandCenterProvider] WARN: No existing CommandCenter found")
	} else {
		commandCenter = resolveCommandCenter
		logging.Info("[CommandCenterProvider] CommandCenter resolved", "CommandCenter", commandCenter)
	}
}

// GetCommandCenter
//  Get resolved {@link CommandCenter} instance.
func GetCommandCenter() transport.CommandCenter {
	return commandCenter
}
