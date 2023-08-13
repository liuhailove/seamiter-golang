package config

import (
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

var (
	globalCfg   = NewDefaultConfig()
	initLogOnce sync.Once
)

func ResetGlobalConfig(config *Entity) {
	globalCfg = config
}

// InitConfigWithYaml loads general configuration from the YAML file under provided path.
func InitConfigWithYaml(filePath string) (err error) {
	// Initialize general config and logging module.
	if err = applyYamlConfigFile(filePath); err != nil {
		return err
	}
	return OverrideConfigFromEnvAndInitLog()
}

// applyYamlConfigFile loads general configuration from the given YAML file.
func applyYamlConfigFile(configPath string) error {
	// Priority: system environment > YAML file > default config
	if util.IsBlank(configPath) {
		// If the config file path is absent, sea will try to resolve it from the system env.
		configPath = os.Getenv(ConfFilePathEnvKey)
	}
	if util.IsBlank(configPath) {
		configPath = DefaultConfigFilename
	}
	// First sea will try to load config from the given file.
	// If the path is empty (not set), sea will use the default config.
	return loadGlobalConfigFromYamlFile(configPath)
}

func OverrideConfigFromEnvAndInitLog() error {
	// Then sea will try to get fundamental config items from system environment.
	// If present, the value in system env will override the value in config file.
	err := overrideItemsFromSystemEnv()
	if err != nil {
		return err
	}
	defer logging.Info("[Config] Print effective global config", "globalConfig", *globalCfg)
	// Configured Logger is the highest priority
	if configLogger := Logger(); configLogger != nil {
		err = logging.ResetGlobalLogger(configLogger)
		if err != nil {
			return err
		}
		return nil
	}

	logDir := LogBaseDir()
	if len(logDir) == 0 {
		logDir = GetDefaultLogDir()
	}
	if err := initializeLogConfig(logDir, LogUsePid()); err != nil {
		return err
	}

	// 日志级别
	logging.ResetGlobalLoggerLevel(logging.Level(LogLevel()))

	logging.Info("[Config] App name resolved", "appName", AppName())
	return nil
}

func loadGlobalConfigFromYamlFile(filePath string) error {
	if filePath == DefaultConfigFilename {
		if _, err := os.Stat(DefaultConfigFilename); err != nil {
			//use default globalCfg.
			return nil
		}
	}
	_, err := os.Stat(filePath)
	if err != nil && !os.IsExist(err) {
		return err
	}
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(content, globalCfg)
	if err != nil {
		return err
	}
	logging.Info("[Config] Resolving sea config from file", "file", filePath)
	return checkConfValid(&(globalCfg.Sea))
}

func overrideItemsFromSystemEnv() error {
	if appName := os.Getenv(AppNameEnvKey); !util.IsBlank(appName) {
		globalCfg.Sea.App.Name = appName
	}

	if appTypeStr := os.Getenv(AppTypeEnvKey); !util.IsBlank(appTypeStr) {
		appType, err := strconv.ParseInt(appTypeStr, 10, 32)
		if err != nil {
			return err
		}
		globalCfg.Sea.App.Type = int32(appType)
	}

	if addPidStr := os.Getenv(LogNamePidEnvKey); !util.IsBlank(addPidStr) {
		addPid, err := strconv.ParseBool(addPidStr)
		if err != nil {
			return err
		}
		globalCfg.Sea.Log.UsePid = addPid
	}
	if logDir := os.Getenv(LogDirEnvKey); !util.IsBlank(logDir) {
		globalCfg.Sea.Log.Dir = logDir
	}
	return checkConfValid(&(globalCfg.Sea))
}

func initializeLogConfig(logDir string, usePid bool) (err error) {
	if logDir == "" {
		return errors.New("invalid empty log path")
	}
	initLogOnce.Do(func() {
		if err = util.CreateDirIfNotExists(logDir); err != nil {
			return
		}
		err = reconfigureRecordLogger(logDir, usePid)
	})
	return err
}

func reconfigureRecordLogger(logBaseDir string, withPid bool) error {
	filePath := filepath.Join(logBaseDir, logging.RecordLogFileName)
	if withPid {
		filePath = filePath + ".pid" + strconv.Itoa(os.Getppid())
	}
	fileLogger, err := logging.NewSimpleFileLogger(filePath)
	if err != nil {
		return err
	}
	// Note: not thread-safe!
	if err = logging.ResetGlobalLogger(fileLogger); err != nil {
		return err
	}
	logging.Info("[Config] Log base directory", "baseDir", logBaseDir)
	return nil
}

func GetDefaultLogDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, logging.DefaultDirName)
}

func AppName() string {
	return globalCfg.AppName()
}

func AppType() int32 {
	return globalCfg.AppType()
}

func Logger() logging.Logger {
	return globalCfg.Logger()
}

func LogBaseDir() string {
	return globalCfg.LogBaseDir()
}

func LogLevel() uint8 {
	return globalCfg.LogLevel()
}

// LogUsePid returns whether the log file name contains the PID suffix.
func LogUsePid() bool {
	return globalCfg.LogUsePid()
}

func MetricExportHTTPAddr() string {
	return globalCfg.MetricExportHTTPAddr()
}

func MetricExportHTTPPath() string {
	return globalCfg.MetricExportHTTPPath()
}

func MetricLogFlushIntervalSec() uint32 {
	return globalCfg.MetricLogFlushIntervalSec()
}

func MetricLogSingleFileMaxSize() uint64 {
	return globalCfg.MetricLogSingleFileMaxSize()
}

func MetricLogMaxFileAmount() uint32 {
	return globalCfg.MetricLogMaxFileAmount()
}

func SystemStatCollectIntervalMs() uint32 {
	return globalCfg.SystemStatCollectIntervalMs()
}

func LoadStatCollectIntervalMs() uint32 {
	return globalCfg.LoadStatCollectIntervalMs()
}

func CpuStatCollectIntervalMs() uint32 {
	return globalCfg.CpuStatCollectIntervalMs()
}

func MemoryStatCollectIntervalMs() uint32 {
	return globalCfg.MemoryStatCollectIntervalMs()
}

func UseCacheTime() bool {
	return globalCfg.UseCacheTime()
}

func GlobalStatisticIntervalMsTotal() uint32 {
	return globalCfg.GlobalStatisticIntervalMsTotal()
}

func GlobalStatisticSampleCountTotal() uint32 {
	return globalCfg.GlobalStatisticSampleCountTotal()
}

func GlobalStatisticBucketLengthInMs() uint32 {
	return globalCfg.GlobalStatisticIntervalMsTotal() / GlobalStatisticSampleCountTotal()
}

func MetricStatisticIntervalMs() uint32 {
	return globalCfg.MetricStatisticIntervalMs()
}
func MetricStatisticSampleCount() uint32 {
	return globalCfg.MetricStatisticSampleCount()
}

func ConsoleServer() string {
	return globalCfg.Sea.Dashboard.Server
}

func ConsolePort() uint32 {
	return globalCfg.Sea.Dashboard.Port
}

func HeartbeatClintIp() string {
	if AutoHeartbeatClientIp() {
		return util.GetIP()
	}
	return globalCfg.Sea.Dashboard.HeartbeatClientIp
}

func AutoHeartbeatClientIp() bool {
	return globalCfg.Sea.Dashboard.AutoHeartbeatClientIp
}

func HeartbeatApiPath() string {
	return globalCfg.Sea.Dashboard.HeartbeatApiPath
}

func HeartBeatIntervalMs() uint64 {
	return globalCfg.Sea.Dashboard.HeartBeatIntervalMs
}

func FetchRuleIntervalMs() uint64 {
	return globalCfg.Sea.Dashboard.FetchRuleIntervalMs
}

func Version() string {
	return globalCfg.Version
}

func FindMaxVersionApiPath() string {
	return globalCfg.Sea.Dashboard.FindMaxVersionApiPath
}

func QueryAllDegradeRuleApiPath() string {
	return globalCfg.Sea.Dashboard.QueryAllDegradeRuleApiPath
}

func QueryAllFlowRuleApiPath() string {
	return globalCfg.Sea.Dashboard.QueryAllFlowRuleApiPath
}

func QueryAllParamFlowRuleApiPath() string {
	return globalCfg.Sea.Dashboard.QueryAllParamFlowRuleApiPath
}

func QueryAllMockRuleApiPath() string {
	return globalCfg.Sea.Dashboard.QueryAllMockRuleApiPath
}

func QueryAllSystemRuleApiPath() string {
	return globalCfg.Sea.Dashboard.QueryAllSystemRuleApiPath
}

func QueryAllAuthorityRuleApiPath() string {
	return globalCfg.Sea.Dashboard.QueryAllAuthorityRuleApiPath
}

func QueryAllRetryRuleApiPath() string {
	return globalCfg.Sea.Dashboard.QueryAllRetryRuleApiPath
}

func SendMetricIntervalMs() uint64 {
	return globalCfg.Sea.Dashboard.SendMetricIntervalMs
}

func SendMetricApiPath() string {
	return globalCfg.Sea.Dashboard.SendMetricApiPath
}

func SendRspApiPathIntervalMs() uint64 {
	return globalCfg.Sea.Dashboard.SendRspApiPathIntervalMs
}

func SendRequestApiPathIntervalMs() uint64 {
	return globalCfg.Sea.Dashboard.SendRequestApiPathIntervalMs
}

func SendRspApiPath() string {
	return globalCfg.Sea.Dashboard.SendRspApiPath
}

func SendRequestApiPath() string {
	return globalCfg.Sea.Dashboard.SendRequestApiPath
}

func ProxyUrl() string {
	return globalCfg.Sea.Dashboard.ProxyUrl
}

func OpenConnectDashboard() bool {
	return globalCfg.Sea.Dashboard.OpenConnectDashboard
}

func CloseAll() bool {
	return globalCfg.Sea.CloseAll
}

func RuleConsistentModeType() RulePersistencetMode {
	return globalCfg.Sea.RulePersistentMode
}

func SourceFilePath() string {
	return globalCfg.Sea.FileDatasourceConfig.SourceFilePath
}

func FlowRuleName() string {
	return globalCfg.Sea.FileDatasourceConfig.FlowRuleName
}

func AuthorityRuleName() string {
	return globalCfg.Sea.FileDatasourceConfig.AuthorityRuleName
}

func DegradeRuleName() string {
	return globalCfg.Sea.FileDatasourceConfig.DegradeRuleName
}

func SystemRuleName() string {
	return globalCfg.Sea.FileDatasourceConfig.SystemRuleName
}

func HotspotRuleName() string {
	return globalCfg.Sea.FileDatasourceConfig.HotspotRuleName
}

func MockRuleName() string {
	return globalCfg.Sea.FileDatasourceConfig.MockRuleName
}

func RetryRuleName() string {
	return globalCfg.Sea.FileDatasourceConfig.RetryRuleName
}
