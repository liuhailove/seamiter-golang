package heartbeat

import (
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	config2 "git.garena.com/honggang.liu/seamiter-go/transport/common/transport/config"
	"git.garena.com/honggang.liu/seamiter-go/transport/common/transport/endpoint"
	"git.garena.com/honggang.liu/seamiter-go/transport/http/heartbeat/client"
	"strconv"
)

const (
	OkStatus        = 200
	DefaultInterval = 1000 * 10
)

// SimpleHttpHeartbeatSender
// The heartbeat sender provides basic API for sending heartbeat request to provided target.
// This implementation is based on a trivial HTTP client.
type SimpleHttpHeartbeatSender struct {
	addressList       []endpoint.Endpoint
	currentAddressIdx int
	heartBeat         *Message
	httpClient        *client.SimpleHttpClient
}

func NewSimpleHttpHeartbeatSender() *SimpleHttpHeartbeatSender {
	var dashboardList = config2.GetConsoleServerList()
	if len(dashboardList) == 0 {
		logging.Warn("[SimpleHttpHeartbeatSender] Dashboard server address not configured or not available")
	} else {
		logging.Info("[SimpleHttpHeartbeatSender] Default console address list retrieved:", "addrs", dashboardList)
	}
	sender := new(SimpleHttpHeartbeatSender)
	sender.addressList = dashboardList
	sender.heartBeat = NewHeartbeatMessage()
	sender.httpClient = new(client.SimpleHttpClient)
	sender.httpClient.Initial(config.ProxyUrl())
	return sender
}
func (s SimpleHttpHeartbeatSender) SendHeartbeat() (bool, error) {
	if config2.GetRuntimePort() <= 0 {
		logging.Info("[SimpleHttpHeartbeatSender] Command server port not initialized, won't send heartbeat")
		return false, nil
	}
	var addrInfo = s.GetAvailableAddress()
	if addrInfo == nil {
		return false, nil
	}
	request := client.NewSimpleHttpRequest(*addrInfo, config2.GetHeartbeatApiPath())
	request.SetParams(s.heartBeat.GenerateCurrentMessage())
	response, err := s.httpClient.Post(request)
	if err != nil {
		logging.Warn("[SimpleHttpHeartbeatSender] Failed to send heartbeat to "+addrInfo.String(), "err", err)
		return false, err
	}
	if response.GetStatusCode() == OkStatus {
		return true, nil
	} else if s.ClientErrorCode(response.GetStatusCode()) || s.ServerErrorCode(response.GetStatusCode()) {
		logging.Warn("[SimpleHttpHeartbeatSender] Failed to send heartbeat to " + addrInfo.String() + ", http status code: " + strconv.Itoa(response.GetStatusCode()))
	}
	return false, nil
}

func (s SimpleHttpHeartbeatSender) IntervalMs() uint64 {
	return DefaultInterval
}

func (s SimpleHttpHeartbeatSender) GetAvailableAddress() *endpoint.Endpoint {
	if s.addressList == nil || len(s.addressList) == 0 {
		return nil
	}
	if s.currentAddressIdx < 0 {
		s.currentAddressIdx = 0
	}
	idx := s.currentAddressIdx % len(s.addressList)
	s.currentAddressIdx++
	return &s.addressList[idx]
}

func (s SimpleHttpHeartbeatSender) ClientErrorCode(code int) bool {
	return code > 399 && code < 500
}

func (s SimpleHttpHeartbeatSender) ServerErrorCode(code int) bool {
	return code > 499 && code < 600
}
