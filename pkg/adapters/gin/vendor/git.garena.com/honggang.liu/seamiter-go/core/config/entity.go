package config

import (
	"fmt"
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

type RulePersistencetMode string

const (
	FileMode   RulePersistencetMode = "file"
	EtcdMode   RulePersistencetMode = "etcdv3"
	DbMode     RulePersistencetMode = "db"
	ZKMode     RulePersistencetMode = "zk"
	ApolloMode RulePersistencetMode = "apollo"
	ConsulMode RulePersistencetMode = "consul"
	K8sMode    RulePersistencetMode = "k8s"
	NacosMode  RulePersistencetMode = "nacos"
)

type Entity struct {
	// Version represents the format version of the entity.
	Version string
	Sea     SeaConfig
}

// FileDatasourceConfig 文件持久化存储配置
type FileDatasourceConfig struct {
	// SourceFilePath 规则文件路径
	SourceFilePath string `yaml:"sourceFilePath"`
	// FlowRuleName 流控规则名称
	FlowRuleName string `yaml:"flowRuleName"`
	// AuthorityRuleName 授权规则名称
	AuthorityRuleName string `yaml:"authorityRuleName"`
	// DegradeRuleName 降级规则名称
	DegradeRuleName string `yaml:"degradeRuleName"`
	// SystemRuleName 系统规则名称
	SystemRuleName string `yaml:"systemRuleName"`
	// HotspotRuleName 热点规则名称
	HotspotRuleName string `yaml:"hotspotRuleName"`
	// MockRuleName mock规则名称
	MockRuleName string `yaml:"mockRuleName"`
	// RetryRuleName retry规则名称
	RetryRuleName string `yaml:"retryRuleName"`
}

// EtcdV3DatasourceConfig etcdv3持久化存储配置
type EtcdV3DatasourceConfig struct {
}

// ZkDatasourceConfig zk持久化存储配置
type ZkDatasourceConfig struct {
}

// NacosDatasourceConfig nacos持久化存储配置
type NacosDatasourceConfig struct {
}

// ApolloDatasourceConfig apollo持久化存储配置
type ApolloDatasourceConfig struct {
}

// SeaConfig represent the general configuration of sea.
type SeaConfig struct {
	CloseAll bool `yaml:"closeAll"`
	// 控制台
	Dashboard struct {
		// Server 地址
		Server string `yaml:"server"`
		// Port 端口号
		Port uint32 `yaml:"port"`
		// HeartbeatClientIp 心跳客户端IP
		HeartbeatClientIp string `yaml:"heartbeatClientIp"`
		// AutoHeartbeatClientIp 自动生成ClientIp和Port
		AutoHeartbeatClientIp bool `yaml:"autoHeartbeatClientIp"`
		// HeartbeatApiPath 心跳路径
		HeartbeatApiPath string `yaml:"heartbeatApiPath"`
		// HeartBeatIntervalMs 心跳间隔，单位ms
		HeartBeatIntervalMs uint64 `yaml:"heartBeatIntervalMs"`
		// FetchRuleIntervalMs 主动拉取规则的时间间隔，单位ms
		FetchRuleIntervalMs uint64 `yaml:"fetchRuleIntervalMs"`
		// FindMaxVersionApiPath 获取当前系统的最新版本路径
		FindMaxVersionApiPath string `yaml:"findMaxVersionApiPath"`
		// QueryAllDegradeRuleApiPath 查询系统的降级规则路径
		QueryAllDegradeRuleApiPath string `yaml:"queryAllDegradeRuleApiPath"`
		// QueryAllFlowRuleApiPath 查询全部流控规则路径
		QueryAllFlowRuleApiPath string `yaml:"queryAllFlowRuleApiPath"`
		// QueryAllParamFlowRuleApiPath 查询热点参数规则路径
		QueryAllParamFlowRuleApiPath string `yaml:"queryAllParamFlowRuleApiPath"`
		// QueryAllMockRuleApiPath 查询Mock规则路径
		QueryAllMockRuleApiPath string `yaml:"queryAllMockRuleApiPath"`
		// QueryAllSystemRuleApiPath 查询系统规则路径
		QueryAllSystemRuleApiPath string `yaml:"queryAllSystemRuleApiPath"`
		// QueryAllAuthorityRuleApiPath 查询授权规则路径
		QueryAllAuthorityRuleApiPath string `yaml:"queryAllAuthorityRuleApiPath"`
		// QueryAllRetryRuleApiPath 查询重试规则路径
		QueryAllRetryRuleApiPath string `yaml:"queryAllRetryRuleApiPath"`
		// SendMetricIntervalMs 上报Metric时间间隔
		SendMetricIntervalMs uint64 `yaml:"sendMetricIntervalMs"`
		// SendMetricApiPath 上报Metric的Api路径
		SendMetricApiPath string `yaml:"sendMetricApiPath"`
		// SendRspApiPath 上报响应Api路径
		SendRspApiPath string `yaml:"sendRspApiPath"`
		// SendRspApiPathIntervalMs 上报Rsp时间间隔
		SendRspApiPathIntervalMs uint64 `yaml:"sendRspApiPathIntervalMs"`
		// SendRequestApiPath 上报请求体Api路径
		SendRequestApiPath string `yaml:"sendRequestApiPath"`
		// SendRequestApiPathIntervalMs 上报Request时间间隔
		SendRequestApiPathIntervalMs uint64 `yaml:"sendRequestApiPathIntervalMs"`
		// ProxyUrl 代理URL，在域名不能直接访问时，需要使用代理
		ProxyUrl string `yaml:"proxyUrl"`
		// OpenConnectDashboard 打开链接Dashboard开关，部分时候，我们可能会选择本地存储，不需要和dashboard通信
		OpenConnectDashboard bool `yaml:"openConnectDashboard"`
	}
	App struct {
		// Name represents the name of current running service.
		Name string
		// Type indicates the classification of the service (e.g. web service, API gateway).
		Type int32
	}
	// Exporter represents configuration items related to exporter, like metric exporter.
	Exporter ExporterConfig
	// Log represents configuration items related to logging.
	Log LogConfig
	// Stat represents configuration items related to statistics.
	Stat StatConfig
	// UseCacheTime indicates whether to cache time(ms)
	UseCacheTime bool `yaml:"useCacheTime"`
	// RulePersistentMode 规则持久化模式
	RulePersistentMode RulePersistencetMode `yaml:"rulePersistentMode"`
	// FileDatasourceConfig 持久化存储配置
	FileDatasourceConfig FileDatasourceConfig `yaml:"fileDatasourceConfig"`
	// EtcdV3DatasourceConfig 持久化存储配置
	EtcdV3DatasourceConfig EtcdV3DatasourceConfig `yaml:"etcdV3DatasourceConfig"`
	// ZkDatasourceConfig 持久化存储配置
	ZkDatasourceConfig ZkDatasourceConfig `yaml:"zkDatasourceConfig"`
	// NacosDatasourceConfig 持久化存储配置
	NacosDatasourceConfig NacosDatasourceConfig `yaml:"nacosDatasourceConfig"`
	// ApolloDatasourceConfig 持久化存储配置
	ApolloDatasourceConfig ApolloDatasourceConfig `yaml:"apolloDatasourceConfig"`
}

// ExporterConfig represents configuration items related to exporter, like metric exporter.
type ExporterConfig struct {
	Metric MetricExporterConfig
}

// MetricExporterConfig represents configuration of metric exporter.
type MetricExporterConfig struct {
	// HttpAddr is the http server listen address, like ":8080".
	HttpAddr string `yaml:"http_addr"`
	// HttpPath is the http request path of access metrics, like "/metrics".
	HttpPath string `yaml:"http_path"`
}

// LogConfig represent the configuration of logging in sea.
type LogConfig struct {
	// Logger indicates that using logger to replace default logging.
	Logger logging.Logger
	// Dir represents the log directory path.
	Dir string
	// UsePid indicates whether the filename ends with the process ID (PID).
	UsePid bool `yaml:"usePid"`
	// Metric represents the configuration items of the metric log.
	Metric MetricLogConfig
	// Level 日志级别
	Level uint8
}

// MetricLogConfig represents the configuration items of the metric log.
type MetricLogConfig struct {
	SingleFileMaxSize uint64 `yaml:"singleFileMaxSize"`
	MaxFileCount      uint32 `yaml:"maxFileCount"`
	FlushIntervalSec  uint32 `yaml:"flushIntervalSec"`
}

// StatConfig represents the configuration items of statistics.
type StatConfig struct {
	// GlobalStatisticSampleCountTotal and GlobalStatisticIntervalMsTotal is the per resource's global default statistic sliding window config
	GlobalStatisticSampleCountTotal uint32 `yaml:"globalStatisticSampleCountTotal"`
	GlobalStatisticIntervalMsTotal  uint32 `yaml:"globalStatisticIntervalMsTotal"`

	// MetricStatisticSampleCount and MetricStatisticIntervalMs is the per resource's default readonly metric statistic
	// This default readonly metric statistic must be reusable based on global statistic.
	MetricStatisticSampleCount uint32 `yaml:"metricStatisticSampleCount"`
	MetricStatisticIntervalMs  uint32 `yaml:"metricStatisticIntervalMs"`

	System SystemStatConfig `yaml:"system"`
}

// SystemStatConfig represents the configuration items of system statistics.
type SystemStatConfig struct {
	// CollectIntervalMs represents the collecting interval of the system metrics collector.
	CollectIntervalMs uint32 `yaml:"collectIntervalMs"`
	// CollectLoadIntervalMs represents the collecting interval of the system load collector.
	CollectLoadIntervalMs uint32 `yaml:"collectLoadIntervalMs"`
	// CollectCpuIntervalMs represents the collecting interval of the system cpu usage collector.
	CollectCpuIntervalMs uint32 `yaml:"collectCpuIntervalMs"`
	// CollectMemoryIntervalMs represents the collecting interval of the system memory usage collector.
	CollectMemoryIntervalMs uint32 `yaml:"collectMemoryIntervalMs"`
}

// NewDefaultConfig creates a new default config entity.
func NewDefaultConfig() *Entity {
	return &Entity{
		Version: "v1",
		Sea: SeaConfig{
			CloseAll: true,
			App: struct {
				Name string
				Type int32
			}{
				Name: UnknownProjectName,
				Type: DefaultAppType,
			},
			Dashboard: struct {
				// Server 地址
				Server string `yaml:"server"`
				// Port 端口号
				Port uint32 `yaml:"port"`
				// HeartbeatClientIp 心跳客户端IP
				HeartbeatClientIp string `yaml:"heartbeatClientIp"`
				// AutoHeartbeatClientIp 自动生成ClientIp和Port
				AutoHeartbeatClientIp bool `yaml:"autoHeartbeatClientIp"`
				// HeartbeatApiPath 心跳路径
				HeartbeatApiPath string `yaml:"heartbeatApiPath"`
				// HeartBeatIntervalMs 心跳间隔，单位ms
				HeartBeatIntervalMs uint64 `yaml:"heartBeatIntervalMs"`
				// FetchRuleIntervalMs 主动拉取规则的时间间隔，单位ms
				FetchRuleIntervalMs uint64 `yaml:"fetchRuleIntervalMs"`
				// FindMaxVersionApiPath 获取当前系统的最新版本路径
				FindMaxVersionApiPath string `yaml:"findMaxVersionApiPath"`
				// QueryAllDegradeRuleApiPath 查询系统的降级规则路径
				QueryAllDegradeRuleApiPath string `yaml:"queryAllDegradeRuleApiPath"`
				// QueryAllFlowRuleApiPath 查询全部流控规则路径
				QueryAllFlowRuleApiPath string `yaml:"queryAllFlowRuleApiPath"`
				// QueryAllParamFlowRuleApiPath 查询热点参数规则路径
				QueryAllParamFlowRuleApiPath string `yaml:"queryAllParamFlowRuleApiPath"`
				// QueryAllMockRuleApiPath 查询Mock规则路径
				QueryAllMockRuleApiPath string `yaml:"queryAllMockRuleApiPath"`
				// QueryAllSystemRuleApiPath 查询系统规则路径
				QueryAllSystemRuleApiPath string `yaml:"queryAllSystemRuleApiPath"`
				// QueryAllAuthorityRuleApiPath 查询授权规则路径
				QueryAllAuthorityRuleApiPath string `yaml:"queryAllAuthorityRuleApiPath"`
				// QueryAllRetryRuleApiPath 查询重试规则路径
				QueryAllRetryRuleApiPath string `yaml:"queryAllRetryRuleApiPath"`
				// SendMetricIntervalMs 上报Metric时间间隔
				SendMetricIntervalMs uint64 `yaml:"sendMetricIntervalMs"`
				// SendMetricApiPath 上报Metric的Api路径
				SendMetricApiPath string `yaml:"sendMetricApiPath"`
				// SendRspApiPath 上报响应Api路径
				SendRspApiPath string `yaml:"sendRspApiPath"`
				// SendRspApiPathIntervalMs 上报Rsp时间间隔
				SendRspApiPathIntervalMs uint64 `yaml:"sendRspApiPathIntervalMs"`
				// SendRequestApiPath 上报请求体Api路径
				SendRequestApiPath string `yaml:"sendRequestApiPath"`
				// SendRequestApiPathIntervalMs 上报Request时间间隔
				SendRequestApiPathIntervalMs uint64 `yaml:"sendRequestApiPathIntervalMs"`
				// ProxyUrl 代理URL，在域名不能直接访问时，需要使用代理
				ProxyUrl string `yaml:"proxyUrl"`
				// OpenConnectDashboard 打开链接Dashboard开关，部分时候，我们可能会选择本地存储，不需要和dashboard通信
				OpenConnectDashboard bool `yaml:"openConnectDashboard"`
			}{
				Server:                       DefaultDashServer,
				Port:                         DefaultHeartbeatPort,
				HeartbeatClientIp:            DefaultHeartbeatClintIp,
				HeartbeatApiPath:             DefaultHeartbeatPath,
				HeartBeatIntervalMs:          DefaultHeartbeatIntervalMs,
				FetchRuleIntervalMs:          DefaultFetchRuleIntervalMs,
				AutoHeartbeatClientIp:        true,
				FindMaxVersionApiPath:        DefaultFindMaxVersionApiPath,
				QueryAllDegradeRuleApiPath:   DefaultQueryAllDegradeRuleApiPath,
				QueryAllFlowRuleApiPath:      DefaultQueryAllFlowRuleApiPath,
				QueryAllParamFlowRuleApiPath: DefaultQueryAllParamFlowRuleApiPath,
				QueryAllMockRuleApiPath:      DefaultQueryAllMockRuleApiPath,
				QueryAllSystemRuleApiPath:    DefaultQueryAllSystemRuleApiPath,
				QueryAllAuthorityRuleApiPath: DefaultQueryAllAuthorityApiPath,
				QueryAllRetryRuleApiPath:     DefaultQueryAllRetryApiPath,
				SendMetricIntervalMs:         DefaultSendIntervalMs,
				SendMetricApiPath:            DefaultSendMetricsApiPath,
				SendRspApiPathIntervalMs:     DefaultSendRspIntervalMs,
				SendRspApiPath:               DefaultSendRspApiPath,
				SendRequestApiPath:           DefaultSendRequestApiPath,
				SendRequestApiPathIntervalMs: DefaultSendRequestIntervalMs,
				ProxyUrl:                     "",
				OpenConnectDashboard:         true,
			},
			Log: LogConfig{
				Logger: nil,
				Dir:    GetDefaultLogDir(),
				UsePid: false,
				Metric: MetricLogConfig{
					SingleFileMaxSize: DefaultMetricLogSingleFileMaxSize,
					MaxFileCount:      DefaultMetricLogMaxFileAmount,
					FlushIntervalSec:  DefaultMetricLogFlushIntervalSec,
				},
				Level: DefaultLogLevel,
			},
			Stat: StatConfig{
				GlobalStatisticSampleCountTotal: base.DefaultSampleCountTotal,
				GlobalStatisticIntervalMsTotal:  base.DefaultIntervalMsTotal,
				MetricStatisticSampleCount:      base.DefaultSampleCount,
				MetricStatisticIntervalMs:       base.DefaultIntervalMs,
				System: SystemStatConfig{
					CollectIntervalMs:       DefaultSystemStatCollectIntervalMs,
					CollectLoadIntervalMs:   DefaultLoadStatCollectIntervalMs,
					CollectCpuIntervalMs:    DefaultCpuStatCollectIntervalMs,
					CollectMemoryIntervalMs: DefaultMemoryStatCollectIntervalMs,
				},
			},
			UseCacheTime:       false,
			RulePersistentMode: FileMode,
			FileDatasourceConfig: FileDatasourceConfig{
				SourceFilePath:    DefaultSourceFilePath,
				FlowRuleName:      DefaultFlowRuleName,
				AuthorityRuleName: DefaultAuthorityRuleName,
				DegradeRuleName:   DefaultDegradeRuleName,
				SystemRuleName:    DefaultSystemRuleName,
				HotspotRuleName:   DefaultHotspotRuleName,
				MockRuleName:      DefaultMockRuleName,
				RetryRuleName:     DefaultRetryRuleName,
			},
		},
	}
}

func CheckValid(entity *Entity) error {
	if entity == nil {
		return errors.New("Nil entity")
	}
	if len(entity.Version) == 0 {
		return errors.New("Empty version")
	}
	return checkConfValid(&entity.Sea)
}

func checkConfValid(conf *SeaConfig) error {
	if conf == nil {
		return errors.New("Nil globalCfg")
	}
	if conf.App.Name == "" {
		return errors.New("App.Name is empty")
	}
	mc := conf.Log.Metric
	if mc.MaxFileCount <= 0 {
		return errors.New("Illegal metric log globalCfg: maxFileCount <= 0")
	}
	if mc.SingleFileMaxSize <= 0 {
		return errors.New("Illegal metric log globalCfg: singleFileMaxSize <= 0")
	}
	if err := base.CheckValidityForReuseStatistic(conf.Stat.MetricStatisticSampleCount, conf.Stat.MetricStatisticIntervalMs,
		conf.Stat.GlobalStatisticSampleCountTotal, conf.Stat.GlobalStatisticIntervalMsTotal); err != nil {
		return err
	}
	return nil
}

func (entity *Entity) String() string {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	e, err := json.Marshal(entity)
	if err != nil {
		return fmt.Sprintf("%+v", *entity)
	}
	return string(e)
}

func (entity *Entity) AppName() string {
	return entity.Sea.App.Name
}

func (entity *Entity) AppType() int32 {
	return entity.Sea.App.Type
}

func (entity *Entity) LogBaseDir() string {
	return entity.Sea.Log.Dir
}

func (entity *Entity) LogLevel() uint8 {
	return entity.Sea.Log.Level
}

func (entity *Entity) Logger() logging.Logger {
	return entity.Sea.Log.Logger
}

// LogUsePid returns whether the log file name contains the PID suffix.
func (entity *Entity) LogUsePid() bool {
	return entity.Sea.Log.UsePid
}

func (entity *Entity) MetricExporterHTTPAddr() string {
	return entity.Sea.Exporter.Metric.HttpAddr
}

func (entity *Entity) MetricLogFlushIntervalSec() uint32 {
	return entity.Sea.Log.Metric.FlushIntervalSec
}

func (entity *Entity) MetricExportHTTPPath() string {
	return entity.Sea.Exporter.Metric.HttpPath
}

func (entity *Entity) MetricExportHTTPAddr() string {
	return entity.Sea.Exporter.Metric.HttpAddr
}

func (entity *Entity) MetricLogSingleFileMaxSize() uint64 {
	return entity.Sea.Log.Metric.SingleFileMaxSize
}

func (entity *Entity) MetricLogMaxFileAmount() uint32 {
	return entity.Sea.Log.Metric.MaxFileCount
}

func (entity *Entity) SystemStatCollectIntervalMs() uint32 {
	return entity.Sea.Stat.System.CollectIntervalMs
}

func (entity *Entity) LoadStatCollectIntervalMs() uint32 {
	return entity.Sea.Stat.System.CollectLoadIntervalMs
}

func (entity *Entity) CpuStatCollectIntervalMs() uint32 {
	return entity.Sea.Stat.System.CollectCpuIntervalMs
}

func (entity *Entity) UseCacheTime() bool {
	return entity.Sea.UseCacheTime
}

func (entity *Entity) MemoryStatCollectIntervalMs() uint32 {
	return entity.Sea.Stat.System.CollectMemoryIntervalMs
}

func (entity *Entity) GlobalStatisticIntervalMsTotal() uint32 {
	return entity.Sea.Stat.GlobalStatisticIntervalMsTotal
}

func (entity *Entity) GlobalStatisticSampleCountTotal() uint32 {
	return entity.Sea.Stat.GlobalStatisticSampleCountTotal
}

func (entity *Entity) MetricStatisticIntervalMs() uint32 {
	return entity.Sea.Stat.MetricStatisticIntervalMs
}

func (entity *Entity) MetricStatisticSampleCount() uint32 {
	return entity.Sea.Stat.MetricStatisticSampleCount
}

func (entity *Entity) DashboardServer() string {
	return entity.Sea.Dashboard.Server
}

func (entity *Entity) DashboardPort() uint32 {
	return entity.Sea.Dashboard.Port
}

func (entity *Entity) HeartbeatClintIp() string {
	return entity.Sea.Dashboard.HeartbeatClientIp
}

func (entity *Entity) HeartbeatApiPath() string {
	return entity.Sea.Dashboard.HeartbeatApiPath
}

func (entity *Entity) HeartBeatIntervalMs() uint64 {
	return entity.Sea.Dashboard.HeartBeatIntervalMs
}

func (entity *Entity) CloseAll() bool {
	return entity.Sea.CloseAll
}

func (entity *Entity) RuleConsistentMode() RulePersistencetMode {
	return entity.Sea.RulePersistentMode
}

func (entity *Entity) SourceFilePath() string {
	return entity.Sea.FileDatasourceConfig.SourceFilePath
}

func (entity *Entity) FlowRuleName() string {
	return entity.Sea.FileDatasourceConfig.FlowRuleName
}

func (entity *Entity) AuthorityRuleName() string {
	return entity.Sea.FileDatasourceConfig.AuthorityRuleName
}

func (entity *Entity) DegradeRuleName() string {
	return entity.Sea.FileDatasourceConfig.DegradeRuleName
}

func (entity *Entity) SystemRuleName() string {
	return entity.Sea.FileDatasourceConfig.SystemRuleName
}

func (entity *Entity) HotspotRuleName() string {
	return entity.Sea.FileDatasourceConfig.HotspotRuleName
}

func (entity *Entity) MockRuleName() string {
	return entity.Sea.FileDatasourceConfig.MockRuleName
}
