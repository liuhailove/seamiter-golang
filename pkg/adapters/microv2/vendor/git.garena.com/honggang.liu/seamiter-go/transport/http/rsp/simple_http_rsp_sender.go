package rsp

import (
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/transport/common/command"
	config2 "git.garena.com/honggang.liu/seamiter-go/transport/common/transport/config"
	"git.garena.com/honggang.liu/seamiter-go/transport/common/transport/endpoint"
	"git.garena.com/honggang.liu/seamiter-go/transport/http/heartbeat/client"
	"github.com/pkg/errors"
	"strconv"
	"sync"
	"time"
)

const (
	OkStatus        = 200
	DefaultInterval = 1000 * 1
)

var (
	handlerMap = make(map[string]command.Handler)
	senderMux  = new(sync.RWMutex)

	// 最近的一次响应数据
	latestRspData string
)

// simpleHttpRspSender http sender
type simpleHttpRspSender struct {
	addressList       []endpoint.Endpoint
	currentAddressIdx int
	message           *Message
	httpClient        *client.SimpleHttpClient
}

func (s simpleHttpRspSender) BeforeStart() {
	// Register handlers
	handlerMap = command.ProviderInst().NamedHandlers()
	latestRspData = "[]"
}

func NewSimpleHttpRspSender() *simpleHttpRspSender {
	var dashboardList = config2.GetConsoleServerList()
	if len(dashboardList) == 0 {
		logging.Warn("[SimpleHttpRspSender] Dashboard server address not configured or not available")
	} else {
		logging.Info("[SimpleHttpRspSender] Default console address list retrieved:", "addrs", dashboardList)
	}
	sender := new(simpleHttpRspSender)
	sender.message = NewRspMessage()
	sender.addressList = dashboardList
	sender.httpClient = new(client.SimpleHttpClient)
	sender.httpClient.Initial(config.ProxyUrl())
	return sender
}

func (s simpleHttpRspSender) SendRsp() (bool, error) {
	senderMux.Lock()
	defer senderMux.Unlock()

	var start = time.Now().UnixNano()
	if config2.GetRuntimePort() <= 0 {
		logging.Info("[SimpleHttpRspSender] Command server port not initialized, won't send heartbeat")
		return false, nil
	}
	var addrInfo = s.GetAvailableAddress()
	if addrInfo == nil {
		return false, nil
	}
	request := client.NewSimpleHttpRequest(*addrInfo, config.SendRspApiPath())

	// 获取规则的最新版本
	// 和当前版本对比，如果不一致，需要拉取规则变更
	var commandRequest = command.NewRequest()
	var h = handlerMap["fetchRsp"]
	var rsp = h.Handle(*commandRequest)
	if rsp == nil || !rsp.IsSuccess() {
		logging.Warn("[SimpleHttpRspSender] handler error", "rsp", rsp)
		return false, errors.New("[SimpleHttpRspSender] handler error")
	}
	if latestRspData != rsp.GetResult().(string) {
		var cost = time.Now().UnixNano() - start
		logging.Debug("[SimpleHttpRspSender] ", "Deal request", h.Name(), "cost(ms)", time.Duration(cost)/time.Millisecond)
		request.SetParams(s.message.GenerateCurrentMessage(rsp.GetResult()))
		response, err := s.httpClient.Post(request)
		if err != nil {
			logging.Warn("[SimpleHttpRspSender] Failed to send metric to "+addrInfo.String(), "err", err)
			return false, err
		}
		if response.GetStatusCode() == OkStatus {
			latestRspData = rsp.GetResult().(string)
			return true, nil
		} else if s.ClientErrorCode(response.GetStatusCode()) || s.ServerErrorCode(response.GetStatusCode()) {
			logging.Warn("[SimpleHttpRspSender] Failed to send metric to " + addrInfo.String() + ", http status code: " + strconv.Itoa(response.GetStatusCode()))
		}
		return false, nil
	}
	return true, nil
}

func (s simpleHttpRspSender) IntervalMs() uint64 {
	return DefaultInterval
}

func (s simpleHttpRspSender) GetAvailableAddress() *endpoint.Endpoint {
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

func (s simpleHttpRspSender) ClientErrorCode(code int) bool {
	return code > 399 && code < 500
}

func (s simpleHttpRspSender) ServerErrorCode(code int) bool {
	return code > 499 && code < 600
}
