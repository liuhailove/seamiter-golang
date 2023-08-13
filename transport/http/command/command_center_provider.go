package command

import (
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/liuhailove/seamiter-golang/transport/common/transport"
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
