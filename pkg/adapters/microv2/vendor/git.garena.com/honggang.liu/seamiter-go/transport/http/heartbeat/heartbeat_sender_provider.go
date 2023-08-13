package heartbeat

import (
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/transport/common/transport"
	"sync"
)

var (
	heartbeatSender transport.HeartBeatSender
	heartbeatMux    = new(sync.Mutex)
)

func resolveInstance() {
	resolved := NewSimpleHttpHeartbeatSender()
	if resolved == nil {
		logging.Warn("[HeartbeatSenderProvider] WARN: No existing HeartbeatSender found")
		return
	}
	heartbeatSender = resolved
	logging.Info("[HeartbeatSenderProvider] HeartbeatSender activated:", "name", "SimpleHttpHeartbeatSender")

}

func GetHeartbeatSender() transport.HeartBeatSender {
	if heartbeatSender == nil {
		heartbeatMux.Lock()
		defer heartbeatMux.Unlock()
		resolveInstance()
	}
	return heartbeatSender
}
