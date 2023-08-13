package config

import (
	"fmt"
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

type Entity struct {
	// Version represents the format version of the entity.
	Version  string
	Sentinel SentinelConfig
}

// SentinelConfig represent the general configuration of Sentinel.
type SentinelConfig struct {
	// 控制台
	Dashboard struct {
		// Server 地址
		Server string
		// Port 端口号
		Port uint32
		// HeartbeatClintIp 心跳客户端IP
		HeartbeatClintIp string
		// HeartbeatApiPath 心跳路径
		HeartbeatApiPath string
		// HeartBeatIntervalMs 心跳间隔，单位ms
		HeartBeatIntervalMs uint64
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

// LogConfig represent the configuration of logging in Sentinel.
type LogConfig struct {
	// Logger indicates that using logger to replace default logging.
	Logger logging.Logger
	// Dir represents the log directory path.
	Dir string
	// UsePid indicates whether the filename ends with the process ID (PID).
	UsePid bool `yaml:"usePid"`
	// Metric represents the configuration items of the metric log.
	Metric MetricLogConfig
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
		Sentinel: SentinelConfig{
			App: struct {
				Name string
				Type int32
			}{
				Name: UnknownProjectName,
				Type: DefaultAppType,
			},
			Dashboard: struct {
				// Server 地址
				Server string
				// Port 端口号
				Port uint32
				// HeartbeatClintIp 心跳客户端IP
				HeartbeatClintIp string
				// HeartbeatApiPath 心跳路径
				HeartbeatApiPath string
				// HeartBeatIntervalMs 心跳间隔，单位ms
				HeartBeatIntervalMs uint64
			}{
				Server:              DefaultDashServer,
				Port:                DefaultHeartbeatPort,
				HeartbeatClintIp:    DefaultHeartbeatClintIp,
				HeartbeatApiPath:    DefaultHeartbeatPath,
				HeartBeatIntervalMs: DefaultHeartbeatIntervalMs,
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
			UseCacheTime: false,
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
	return checkConfValid(&entity.Sentinel)
}

func checkConfValid(conf *SentinelConfig) error {
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
	return entity.Sentinel.App.Name
}

func (entity *Entity) AppType() int32 {
	return entity.Sentinel.App.Type
}

func (entity *Entity) LogBaseDir() string {
	return entity.Sentinel.Log.Dir
}

func (entity *Entity) Logger() logging.Logger {
	return entity.Sentinel.Log.Logger
}

// LogUsePid returns whether the log file name contains the PID suffix.
func (entity *Entity) LogUsePid() bool {
	return entity.Sentinel.Log.UsePid
}

func (entity *Entity) MetricExporterHTTPAddr() string {
	return entity.Sentinel.Exporter.Metric.HttpAddr
}

func (entity *Entity) MetricLogFlushIntervalSec() uint32 {
	return entity.Sentinel.Log.Metric.FlushIntervalSec
}

func (entity *Entity) MetricExportHTTPPath() string {
	return entity.Sentinel.Exporter.Metric.HttpPath
}

func (entity *Entity) MetricExportHTTPAddr() string {
	return entity.Sentinel.Exporter.Metric.HttpAddr
}

func (entity *Entity) MetricLogSingleFileMaxSize() uint64 {
	return entity.Sentinel.Log.Metric.SingleFileMaxSize
}

func (entity *Entity) MetricLogMaxFileAmount() uint32 {
	return entity.Sentinel.Log.Metric.MaxFileCount
}

func (entity *Entity) SystemStatCollectIntervalMs() uint32 {
	return entity.Sentinel.Stat.System.CollectIntervalMs
}

func (entity *Entity) LoadStatCollectIntervalMs() uint32 {
	return entity.Sentinel.Stat.System.CollectLoadIntervalMs
}

func (entity *Entity) CpuStatCollectIntervalMs() uint32 {
	return entity.Sentinel.Stat.System.CollectCpuIntervalMs
}

func (entity *Entity) UseCacheTime() bool {
	return entity.Sentinel.UseCacheTime
}

func (entity *Entity) MemoryStatCollectIntervalMs() uint32 {
	return entity.Sentinel.Stat.System.CollectMemoryIntervalMs
}

func (entity *Entity) GlobalStatisticIntervalMsTotal() uint32 {
	return entity.Sentinel.Stat.GlobalStatisticIntervalMsTotal
}

func (entity *Entity) GlobalStatisticSampleCountTotal() uint32 {
	return entity.Sentinel.Stat.GlobalStatisticSampleCountTotal
}

func (entity *Entity) MetricStatisticIntervalMs() uint32 {
	return entity.Sentinel.Stat.MetricStatisticIntervalMs
}

func (entity *Entity) MetricStatisticSampleCount() uint32 {
	return entity.Sentinel.Stat.MetricStatisticSampleCount
}

func (entity *Entity) DashboardServer() string {
	return entity.Sentinel.Dashboard.Server
}

func (entity *Entity) DashboardPort() uint32 {
	return entity.Sentinel.Dashboard.Port
}

func (entity *Entity) HeartbeatClintIp() string {
	return entity.Sentinel.Dashboard.HeartbeatClintIp
}

func (entity *Entity) HeartbeatApiPath() string {
	return entity.Sentinel.Dashboard.HeartbeatApiPath
}

func (entity *Entity) HeartBeatIntervalMs() uint64 {
	return entity.Sentinel.Dashboard.HeartBeatIntervalMs
}
