package api

import (
	"context"
	"fmt"
	"github.com/liuhailove/seamiter-golang/core/config"
	"github.com/liuhailove/seamiter-golang/core/log/metric"
	"github.com/liuhailove/seamiter-golang/core/system_metric"
	metric_exporter "github.com/liuhailove/seamiter-golang/exporter/metric"
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/liuhailove/seamiter-golang/util"
	"github.com/pkg/errors"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func init() {
	// 种子初始化
	rand.Seed(time.Now().UnixNano())
}

// InitDefault 初始化semiater的环境变量，包含：
//      1. 覆盖全局配置，可以来自手动的配置，YAML文件或者环境变量
//      2. 覆盖全局的Logger
//      3. 初始化全局组件，包含:metric log, system statistic，transport ....
func InitDefault() error {
	return initsea("")
}

// InitWithParser 使用给定的解析方法反序列化configBytes并返回config.Entity
func InitWithParser(configBytes []byte, parser func([]byte) (*config.Entity, error)) (err error) {
	if parser == nil {
		return errors.New("nil parser")
	}
	confEntity, err := parser(configBytes)
	if err != nil {
		return err
	}
	return InitWithConfig(confEntity)
}

// InitWithConfig 使用配置初始化Seamiter
func InitWithConfig(confEntity *config.Entity) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	err = config.CheckValid(confEntity)
	if err != nil {
		return err
	}
	config.ResetGlobalConfig(confEntity)
	if err = config.OverrideConfigFromEnvAndInitLog(); err != nil {
		return err
	}
	return initCoreComponents()
}

// InitWithConfigFile 根据给定的YAML文件路径初始化seamiter
func InitWithConfigFile(configPath string) error {
	return initsea(configPath)
}

// initCoreComponents 根据全局配置文件初始化系统环境变量
func initCoreComponents() error {
	if config.MetricLogFlushIntervalSec() > 0 {
		if err := metric.InitTask(); err != nil {
			return err
		}
	}

	systemStatInterval := config.SystemStatCollectIntervalMs()
	loadStatInterval := systemStatInterval
	cpuStatInterval := systemStatInterval
	memStatInterval := systemStatInterval

	if config.LoadStatCollectIntervalMs() > 0 {
		loadStatInterval = config.LoadStatCollectIntervalMs()
	}
	if config.CpuStatCollectIntervalMs() > 0 {
		cpuStatInterval = config.CpuStatCollectIntervalMs()
	}
	if config.MemoryStatCollectIntervalMs() > 0 {
		memStatInterval = config.MemoryStatCollectIntervalMs()
	}

	if loadStatInterval > 0 {
		system_metric.InitLoadCollector(loadStatInterval)
	}
	if cpuStatInterval > 0 {
		system_metric.InitCpuCollector(cpuStatInterval)
	}
	if memStatInterval > 0 {
		system_metric.InitMemoryCollector(memStatInterval)
	}

	if config.UseCacheTime() {
		util.StartTimeTicker()
	}

	// 如果prometheus配置不为空，初始化server
	if config.MetricExportHTTPAddr() != "" {
		httpAddr := config.MetricExportHTTPAddr()
		httpPath := config.MetricExportHTTPPath()
		util.Try(func() {
			l, err := net.Listen("tcp", httpAddr)
			if err != nil {
				if errOp, ok := err.(*net.OpError); ok {
					if errOp.Err != nil {
						if errSys, ok := errOp.Err.(*os.SyscallError); ok {
							if errSys.Err == syscall.EADDRINUSE {
								l2 := &net.ListenConfig{Control: reusePortControl}
								// 使用本地端口
								_, err = strconv.ParseInt(httpAddr, 10, 64)
								if err == nil {
									l, err = l2.Listen(context.Background(), "tcp", "localhost:"+httpAddr)
								} else {
									l, err = l2.Listen(context.Background(), "tcp", "localhost:"+httpAddr[strings.LastIndex(":", httpAddr)+1:])
								}
							}
						}
					}
				} else {
					panic(fmt.Errorf("init metric exporter http server err: %s", err.Error()))
				}
			}
			if err != nil {
				panic(fmt.Errorf("init metric exporter http server err: %s", err.Error()))
			}
			http.Handle(httpPath, metric_exporter.HTTPHandler())
			go util.RunWithRecover(func() {
				err = http.Serve(l, nil)
				if err != nil {
					panic(err)
				}
			})
		}).CatchAll(func(err error) {
			logging.Warn("listen metric exporter", "err", err)
		})
	}
	// 通信初始化
	err := doInit()
	if err != nil {
		return err
	}
	return nil
}

func initsea(configPath string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	// Initialize general config and logging module.
	if err = config.InitConfigWithYaml(configPath); err != nil {
		return err
	}
	return initCoreComponents()
}

// reusePortControl 端口重用
func reusePortControl(network, address string, c syscall.RawConn) error {
	var opErr error
	err := c.Control(func(fd uintptr) {
		// syscall.SO_REUSEPORT ,在Linux下还可以指定端口重用
		opErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	})
	if err != nil {
		return err
	}
	return opErr
}
