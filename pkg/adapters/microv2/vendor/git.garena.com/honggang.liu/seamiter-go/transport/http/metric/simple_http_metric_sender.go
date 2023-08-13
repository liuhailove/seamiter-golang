package metric

import (
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/transport/common/command"
	config2 "git.garena.com/honggang.liu/seamiter-go/transport/common/transport/config"
	"git.garena.com/honggang.liu/seamiter-go/transport/common/transport/endpoint"
	"git.garena.com/honggang.liu/seamiter-go/transport/http/heartbeat/client"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

const (
	OkStatus        = 200
	DefaultInterval = 1000 * 1

	// FetchIntervalSecond 拉取时间间隔
	FetchIntervalSecond = 2 * 1000

	MaxLastFetchIntervalMs = 15 * 1000

	// MaxRetryNum 最大重试次数
	MaxRetryNum = 5
)

var (
	handlerMap = make(map[string]command.Handler)
	// 最近一次sendTime
	latestSendTime uint64
)

// simpleHttpMetricSender http sender
type simpleHttpMetricSender struct {
	addressList       []endpoint.Endpoint
	currentAddressIdx int
	message           *Message
	httpClient        *client.SimpleHttpClient
}

func NewSimpleHttpMetricSender() *simpleHttpMetricSender {
	var dashboardList = config2.GetConsoleServerList()
	if len(dashboardList) == 0 {
		logging.Warn("[SimpleHttpMetricSender] Dashboard server address not configured or not available")
	} else {
		logging.Info("[SimpleHttpMetricSender] Default console address list retrieved:", "addrs", dashboardList)
	}
	sender := new(simpleHttpMetricSender)
	sender.message = NewMetricMessage()
	sender.addressList = dashboardList
	sender.httpClient = new(client.SimpleHttpClient)
	sender.httpClient.Initial(config.ProxyUrl())
	return sender
}

func (s simpleHttpMetricSender) BeforeStart() {
	// Register handlers
	handlerMap = command.ProviderInst().NamedHandlers()
}

func (s simpleHttpMetricSender) SendMetric() (bool, error) {
	var now = util.CurrentTimeMillis()
	// trim milliseconds
	var lastFetchMs = now - MaxLastFetchIntervalMs
	// 和发送的下一秒对比
	if lastFetchMs < latestSendTime+1000*1 {
		lastFetchMs = latestSendTime + 1000*1
	}
	lastFetchMs = lastFetchMs / 1000 * 1000
	var endTime = lastFetchMs + FetchIntervalSecond
	if endTime > now-1000*2 {
		// too near
		return true, nil
	}
	latestSendTime = endTime
	var finalLastFetchMs = lastFetchMs
	var finalEndTime = endTime
	var start = time.Now().UnixNano()
	if config2.GetRuntimePort() <= 0 {
		logging.Info("[SimpleHttpHeartbeatSender] Command server port not initialized, won't send heartbeat")
		return false, nil
	}
	var addrInfo = s.GetAvailableAddress()
	if addrInfo == nil {
		return false, nil
	}
	request := client.NewSimpleHttpRequest(*addrInfo, config.SendMetricApiPath())

	// 获取规则的最新版本
	// 和当前版本对比，如果不一致，需要拉取规则变更
	var commandRequest = command.NewRequest()
	_ = commandRequest.AddParam("startTime", strconv.FormatUint(finalLastFetchMs, 10))
	_ = commandRequest.AddParam("endTime", strconv.FormatUint(finalEndTime, 10))
	var h = handlerMap["metric"]
	var rsp = h.Handle(*commandRequest)
	if rsp == nil || !rsp.IsSuccess() {
		logging.Warn("[SimpleHttpMetricSender] handler error", "rsp", rsp)
		return false, errors.New("[SimpleHttpMetricSender] handler error")
	}
	var cost = time.Now().UnixNano() - start
	logging.Debug("[SimpleHttpMetricSender] ", "Deal request", h.Name(), "cost(ms)", time.Duration(cost)/time.Millisecond)
	request.SetParams(s.message.GenerateCurrentMessage(rsp.GetResult()))
	var i = 0
	for i = 0; i < MaxRetryNum; i++ {
		response, err := s.httpClient.Post(request)
		if err != nil {
			if i == MaxRetryNum-1 {
				logging.Warn("[SimpleHttpMetricSender] Failed to send metric to "+addrInfo.String(), "err", err)
				return false, err
			}
			// 休眠50ms后重试
			util.Sleep(50 * time.Millisecond)
			continue
		}
		if response.GetStatusCode() == OkStatus {
			return true, nil
		}
		// 休眠50ms后重试
		util.Sleep(50 * time.Millisecond)
		if i == MaxRetryNum-1 {
			if s.ClientErrorCode(response.GetStatusCode()) || s.ServerErrorCode(response.GetStatusCode()) {
				logging.Warn("[SimpleHttpMetricSender] Failed to send metric to " + addrInfo.String() + ", http status code: " + strconv.Itoa(response.GetStatusCode()))
			}
		}
	}
	//response, err := s.httpClient.Post(request)
	//if err != nil {
	//	logging.Warn("[SimpleHttpMetricSender] Failed to send metric to "+addrInfo.String(), "err", err)
	//	return false, err
	//}
	//if response.GetStatusCode() == OkStatus {
	//	return true, nil
	//} else if s.ClientErrorCode(response.GetStatusCode()) || s.ServerErrorCode(response.GetStatusCode()) {
	//	logging.Warn("[SimpleHttpMetricSender] Failed to send metric to " + addrInfo.String() + ", http status code: " + strconv.Itoa(response.GetStatusCode()))
	//}
	return false, nil
}

func (s simpleHttpMetricSender) IntervalMs() uint64 {
	return DefaultInterval
}

func (s simpleHttpMetricSender) GetAvailableAddress() *endpoint.Endpoint {
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

func (s simpleHttpMetricSender) ClientErrorCode(code int) bool {
	return code > 399 && code < 500
}

func (s simpleHttpMetricSender) ServerErrorCode(code int) bool {
	return code > 499 && code < 600
}
