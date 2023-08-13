package request

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
	DefaultInterval = 2000 * 1

	// NilMapJson 空map转的json
	NilMapJson = "{}"
)

var (
	handlerMap = make(map[string]command.Handler)
	senderMux  = new(sync.RWMutex)
	// 最近一次的请求数据
	latestRequestData = NilMapJson
)

// simpleHttpRequestSender http sender
type simpleHttpRequestSender struct {
	addressList       []endpoint.Endpoint
	currentAddressIdx int
	message           *Message
	httpClient        *client.SimpleHttpClient
}

func (s simpleHttpRequestSender) BeforeStart() {
	// Register handlers
	handlerMap = command.ProviderInst().NamedHandlers()
}

func NewSimpleHttpRequestSender() *simpleHttpRequestSender {
	var dashboardList = config2.GetConsoleServerList()
	if len(dashboardList) == 0 {
		logging.Warn("[NewSimpleHttpRequestSender] Dashboard server address not configured or not available")
	} else {
		logging.Info("[NewSimpleHttpRequestSender] Default console address list retrieved:", "addrs", dashboardList)
	}
	sender := new(simpleHttpRequestSender)
	sender.message = NewRspMessage()
	sender.addressList = dashboardList
	sender.httpClient = new(client.SimpleHttpClient)
	sender.httpClient.Initial(config.ProxyUrl())
	return sender
}

func (s simpleHttpRequestSender) SendRequest() (bool, error) {
	senderMux.Lock()
	defer senderMux.Unlock()

	var start = time.Now().UnixNano()
	if config2.GetRuntimePort() <= 0 {
		logging.Info("[simpleHttpRequestSender] Command server port not initialized, won't send heartbeat")
		return false, nil
	}
	var addrInfo = s.GetAvailableAddress()
	if addrInfo == nil {
		return false, nil
	}
	request := client.NewSimpleHttpRequest(*addrInfo, config.SendRequestApiPath())

	// 获取规则的最新版本
	// 和当前版本对比，如果不一致，需要拉取规则变更
	var commandRequest = command.NewRequest()
	var h = handlerMap["fetchRequest"]
	var rsp = h.Handle(*commandRequest)
	if rsp == nil || !rsp.IsSuccess() {
		logging.Warn("[simpleHttpRequestSender] handler error", "rsp", rsp)
		return false, errors.New("[simpleHttpRequestSender] handler error")
	}
	var result = rsp.GetResult().(string)
	if latestRequestData != result && NilMapJson != result {
		var cost = time.Now().UnixNano() - start
		logging.Debug("[simpleHttpRequestSender] ", "Deal request", h.Name(), "cost(ms)", time.Duration(cost)/time.Millisecond)
		request.SetParams(s.message.GenerateCurrentMessage(rsp.GetResult().(string)))
		response, err := s.httpClient.Post(request)
		if err != nil {
			logging.Warn("[simpleHttpRequestSender] Failed to send metric to "+addrInfo.String(), "err", err)
			return false, err
		}
		if response.GetStatusCode() == OkStatus {
			latestRequestData = result
			return true, nil
		} else if s.ClientErrorCode(response.GetStatusCode()) || s.ServerErrorCode(response.GetStatusCode()) {
			logging.Warn("[simpleHttpRequestSender] Failed to send metric to " + addrInfo.String() + ", http status code: " + strconv.Itoa(response.GetStatusCode()))
		}
		return false, nil
	}
	return true, nil
}

func (s simpleHttpRequestSender) IntervalMs() uint64 {
	return DefaultInterval
}

func (s simpleHttpRequestSender) GetAvailableAddress() *endpoint.Endpoint {
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

func (s simpleHttpRequestSender) ClientErrorCode(code int) bool {
	return code > 399 && code < 500
}

func (s simpleHttpRequestSender) ServerErrorCode(code int) bool {
	return code > 499 && code < 600
}
