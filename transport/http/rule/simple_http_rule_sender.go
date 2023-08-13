package rule

import (
	"github.com/liuhailove/seamiter-golang/core/config"
	"github.com/liuhailove/seamiter-golang/ext/datasource"
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/liuhailove/seamiter-golang/transport/common/command"
	config2 "github.com/liuhailove/seamiter-golang/transport/common/transport/config"
	"github.com/liuhailove/seamiter-golang/transport/common/transport/endpoint"
	"github.com/liuhailove/seamiter-golang/transport/http/heartbeat/client"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	FlowRuleType      int32 = 1
	DegradeRuleType   int32 = 2
	HotParamRuleType  int32 = 3
	MockRuleType      int32 = 4
	SystemRuleType    int32 = 5
	AuthorityRuleType int32 = 6
	RetryRuleType     int32 = 7
	GrayRuleType      int32 = 8
	IsolationRuleType int32 = 9

	OkStatus = 200
)

var (
	ErrFetchNoData = errors.New("fetch no data")
)

var (
	handlerMap     = make(map[string]command.Handler)
	ruleTypeKeyMap = make(map[int32]string)

	FlowRuleTypeCurrentVersion      = "v_0"
	DegradeRuleTypeCurrentVersion   = "v_0"
	HotParamRuleTypeCurrentVersion  = "v_0"
	MockRuleTypeCurrentVersion      = "v_0"
	SystemRuleTypeCurrentVersion    = "v_0"
	AuthorityRuleTypeCurrentVersion = "v_0"
	RetryRuleTypeCurrentVersion     = "v_0"
	GrayRuleTypeCurrentVersion      = "v_0"
	IsolationRuleTypeCurrentVersion = "v_0"

	versionMux     = new(sync.RWMutex)
	findVersionMux = new(sync.RWMutex)
	findRuleMux    = new(sync.RWMutex)
	handleRuleMux  = new(sync.RWMutex)
)

type simpleHttpRuleSender struct {
	addressList       []endpoint.Endpoint
	fetchMaxVersion   *Message
	currentAddressIdx int
	httpClient        *client.SimpleHttpClient
}

func NewSimpleHttpRuleSender() *simpleHttpRuleSender {
	var dashboardList = config2.GetConsoleServerList()
	if len(dashboardList) == 0 {
		logging.Warn("[SimpleHttpRuleSender] Dashboard server address not configured or not available")
	} else {
		logging.Info("[SimpleHttpRuleSender] Default console address list retrieved:", "addrs", dashboardList)
	}
	sender := new(simpleHttpRuleSender)
	sender.addressList = dashboardList
	sender.fetchMaxVersion = NewFetchMessage()
	sender.httpClient = new(client.SimpleHttpClient)
	sender.httpClient.Initial(config.ProxyUrl())
	return sender
}

// SetRuleTypeCurrentVersion 设置规则的当前版本
func (s simpleHttpRuleSender) SetRuleTypeCurrentVersion(ruleType int32, version string) {
	versionMux.Lock()
	defer versionMux.Unlock()

	switch ruleType {
	case FlowRuleType:
		FlowRuleTypeCurrentVersion = version
	case DegradeRuleType:
		DegradeRuleTypeCurrentVersion = version
	case HotParamRuleType:
		HotParamRuleTypeCurrentVersion = version
	case MockRuleType:
		MockRuleTypeCurrentVersion = version
	case SystemRuleType:
		SystemRuleTypeCurrentVersion = version
	case AuthorityRuleType:
		AuthorityRuleTypeCurrentVersion = version
	case RetryRuleType:
		RetryRuleTypeCurrentVersion = version
	case GrayRuleType:
		GrayRuleTypeCurrentVersion = version
	case IsolationRuleType:
		IsolationRuleTypeCurrentVersion = version
	default:
	}
}

// GetRuleTypeCurrentVersion 获取规则的当前版本
func (s simpleHttpRuleSender) GetRuleTypeCurrentVersion(ruleType int32) string {
	versionMux.Lock()
	defer versionMux.Unlock()

	switch ruleType {
	case FlowRuleType:
		return FlowRuleTypeCurrentVersion
	case DegradeRuleType:
		return DegradeRuleTypeCurrentVersion
	case HotParamRuleType:
		return HotParamRuleTypeCurrentVersion
	case MockRuleType:
		return MockRuleTypeCurrentVersion
	case SystemRuleType:
		return SystemRuleTypeCurrentVersion
	case AuthorityRuleType:
		return AuthorityRuleTypeCurrentVersion
	case RetryRuleType:
		return RetryRuleTypeCurrentVersion
	case GrayRuleType:
		return GrayRuleTypeCurrentVersion
	case IsolationRuleType:
		return IsolationRuleTypeCurrentVersion
	default:
		return "v_0"
	}
}

// GetRulesCurrentVersionStr
//
// 获取当前业务中规则的版本号
func GetRulesCurrentVersionStr() string {
	versionMux.Lock()
	defer versionMux.Unlock()
	var currentVersionStr string
	currentVersionStr += "AuthorityRuleType" + ":" + AuthorityRuleTypeCurrentVersion + "\r\n"
	currentVersionStr += "DegradeRuleType" + ":" + DegradeRuleTypeCurrentVersion + "\r\n"
	currentVersionStr += "FlowRuleType" + ":" + FlowRuleTypeCurrentVersion + "\r\n"
	currentVersionStr += "GrayRuleType" + ":" + GrayRuleTypeCurrentVersion + "\r\n"
	currentVersionStr += "HotParamRuleType" + ":" + HotParamRuleTypeCurrentVersion + "\r\n"
	currentVersionStr += "MockRuleType" + ":" + MockRuleTypeCurrentVersion + "\r\n"
	currentVersionStr += "RetryRuleType" + ":" + RetryRuleTypeCurrentVersion + "\r\n"
	currentVersionStr += "SystemRuleType" + ":" + SystemRuleTypeCurrentVersion + "\r\n"
	currentVersionStr += "IsolationRuleType" + ":" + IsolationRuleTypeCurrentVersion + "\r\n"

	return currentVersionStr
}

// Rest2ImpossibleVersion 版本复原到不可能版本，目的主要是能够使SDK重新全量拉取规则
func (s simpleHttpRuleSender) Rest2ImpossibleVersion() {
	versionMux.Lock()
	defer versionMux.Unlock()

	FlowRuleTypeCurrentVersion = "v_00"
	DegradeRuleTypeCurrentVersion = "v_00"
	HotParamRuleTypeCurrentVersion = "v_00"
	MockRuleTypeCurrentVersion = "v_00"
	SystemRuleTypeCurrentVersion = "v_00"
	AuthorityRuleTypeCurrentVersion = "v_00"
	RetryRuleTypeCurrentVersion = "v_00"
	GrayRuleTypeCurrentVersion = "v_00"
	IsolationRuleTypeCurrentVersion = "v_00"

}

// RuleTypes 返回规则类
func (s simpleHttpRuleSender) RuleTypes() []int32 {
	return []int32{FlowRuleType, DegradeRuleType, HotParamRuleType, MockRuleType, SystemRuleType, RetryRuleType, GrayRuleType, IsolationRuleType}
}

func (s simpleHttpRuleSender) RuleTypeStr() string {
	var ruleTypeStr string
	for _, ruleType := range s.RuleTypes() {
		ruleTypeStr += strconv.Itoa(int(ruleType)) + ","
	}
	return strings.TrimRight(ruleTypeStr, ",")
}

func (s simpleHttpRuleSender) GetRuleTypeCurrentVersions() string {
	var ruleTypeVersionStr string
	for _, ruleType := range s.RuleTypes() {
		ruleTypeVersionStr += s.GetRuleTypeCurrentVersion(ruleType) + ","
	}
	return strings.TrimRight(ruleTypeVersionStr, ",")
}
func (s simpleHttpRuleSender) BeforeStart() {
	// Register handlers
	handlerMap = command.ProviderInst().NamedHandlers()

	ruleTypeKeyMap[FlowRuleType] = "flow"
	ruleTypeKeyMap[DegradeRuleType] = "degrade"
	ruleTypeKeyMap[HotParamRuleType] = "hotParam"
	ruleTypeKeyMap[MockRuleType] = "mock"
	ruleTypeKeyMap[AuthorityRuleType] = "authority"
	ruleTypeKeyMap[RetryRuleType] = "retry"
	ruleTypeKeyMap[SystemRuleType] = "system"
	ruleTypeKeyMap[GrayRuleType] = "gray"
	ruleTypeKeyMap[IsolationRuleType] = "isolation"

}

// FindMaxVersion 查找规则的最大版本
func (s simpleHttpRuleSender) FindMaxVersion() (map[int32]string, error) {
	findVersionMux.Lock()
	defer findVersionMux.Unlock()

	// 一次查找所有规则的最新版本
	var data []byte
	var err error
	var publishesItf interface{}
	data, err = s.FetchData(config.FindMaxVersionApiPath())
	if err != nil || data == nil {
		return nil, err
	}
	publishesItf, err = datasource.RulePublishJsonArrayParser(data)
	if err != nil {
		logging.Warn("[simpleHttpRuleSender] FindMaxVersion RulePublishJsonArrayParser error", "err", err.Error())
		return nil, errors.Wrap(err, "FindMaxVersion RulePublishJsonArrayParser error")
	}
	var publishes = publishesItf.([]*datasource.Publish)
	var maxVersionMap = make(map[int32]string)
	for _, publish := range publishes {
		maxVersionMap[publish.RuleType] = publish.Version
	}
	return maxVersionMap, nil
}

// Check 检查ruleType对应的CurrentVersion是否为最新版本
func (s simpleHttpRuleSender) Check(ruleType int32, latestVersion string) bool {
	return s.GetRuleTypeCurrentVersion(ruleType) == latestVersion
}

func (s simpleHttpRuleSender) FindRule(ruleType int32) (string, error) {
	findRuleMux.Lock()
	defer findRuleMux.Unlock()

	// 一次查找虽有规则的最新版本
	var data []byte
	var err error
	switch ruleType {
	case FlowRuleType:
		data, err = s.FetchData(config.QueryAllFlowRuleApiPath())
	case DegradeRuleType:
		data, err = s.FetchData(config.QueryAllDegradeRuleApiPath())
	case HotParamRuleType:
		data, err = s.FetchData(config.QueryAllParamFlowRuleApiPath())
	case MockRuleType:
		data, err = s.FetchData(config.QueryAllMockRuleApiPath())
	case SystemRuleType:
		data, err = s.FetchData(config.QueryAllSystemRuleApiPath())
	case AuthorityRuleType:
		data, err = s.FetchData(config.QueryAllAuthorityRuleApiPath())
	case RetryRuleType:
		data, err = s.FetchData(config.QueryAllRetryRuleApiPath())
	case GrayRuleType:
		data, err = s.FetchData(config.QueryAllGrayRuleApiPath())
	case IsolationRuleType:
		data, err = s.FetchData(config.QueryAllIsolationRuleApiPath())
	default:
	}
	if err != nil {
		return "", err
	}
	if data == nil {
		return "", ErrFetchNoData
	}
	return string(data), nil
}

func (s simpleHttpRuleSender) HandleRule(ruleType int32, data string) error {
	handleRuleMux.Lock()
	defer handleRuleMux.Unlock()

	var start = time.Now().UnixNano()
	// 获取规则的最新版本
	// 和当前版本对比，如果不一致，需要拉取规则变更
	var commandRequest = command.NewRequest()
	_ = commandRequest.AddParam("type", ruleTypeKeyMap[ruleType])
	_ = commandRequest.AddParam("data", data)
	var h command.Handler
	switch ruleType {
	case FlowRuleType, DegradeRuleType, SystemRuleType, AuthorityRuleType, IsolationRuleType:
		h = handlerMap["setRules"]
	case HotParamRuleType:
		h = handlerMap["setParamFlowRules"]
	case MockRuleType:
		h = handlerMap["setMockRules"]
	case RetryRuleType:
		h = handlerMap["setRetryRules"]
	case GrayRuleType:
		h = handlerMap["setGrayRules"]
	default:
	}
	if h == nil {
		logging.Warn("[simpleHttpRuleSender] h not exist", "handler", h)
		return nil
	}
	var rsp = h.Handle(*commandRequest)
	if rsp == nil || !rsp.IsSuccess() {
		logging.Warn("[simpleHttpRuleSender] handler error", "rsp", rsp)
		return errors.New("[simpleHttpRuleSender] handler error")
	}
	var cost = time.Now().UnixNano() - start
	logging.Debug("[simpleHttpRuleSender] ", "Deal request", h.Name(), "cost(ms)", time.Duration(cost)/time.Millisecond)
	return nil
}

func (s simpleHttpRuleSender) FetchData(apiPath string) ([]byte, error) {
	if config2.GetRuntimePort() <= 0 {
		logging.Info("[SimpleHttpRuleSender] Command server port not initialized, won't send heartbeat")
		return nil, nil
	}
	var addrInfo = s.GetAvailableAddress()
	if addrInfo == nil {
		return nil, nil
	}
	request := client.NewSimpleHttpRequest(*addrInfo, apiPath)
	request.SetParams(s.fetchMaxVersion.GenerateCurrentMessage(s))
	response, err := s.httpClient.Post(request)
	if err != nil {
		logging.Warn("[SimpleHttpRuleSender] Failed to send heartbeat to "+addrInfo.String(), "err", err)
		return nil, err
	}
	if response.GetStatusCode() == OkStatus {
		return response.GetBody(), nil
	} else if s.ClientErrorCode(response.GetStatusCode()) || s.ServerErrorCode(response.GetStatusCode()) {
		logging.Warn("[SimpleHttpRuleSender] Failed to send heartbeat to " + addrInfo.String() + ", http status code: " + strconv.Itoa(response.GetStatusCode()))
	}
	return nil, nil
}

func (s simpleHttpRuleSender) GetAvailableAddress() *endpoint.Endpoint {
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

func (s simpleHttpRuleSender) ClientErrorCode(code int) bool {
	return code > 399 && code < 500
}

func (s simpleHttpRuleSender) ServerErrorCode(code int) bool {
	return code > 499 && code < 600
}
