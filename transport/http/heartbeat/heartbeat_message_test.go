package heartbeat

import (
	"fmt"
	"testing"
)

func TestHeartbeatMessage_NewHeartbeatMessage(t *testing.T) {
	message := NewHeartbeatMessage()
	message.RegisterInformation("test", "test")
	fmt.Println(message)
}
