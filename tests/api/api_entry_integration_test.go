package api

import (
	"github.com/liuhailove/seamiter-golang/api"
	"github.com/liuhailove/seamiter-golang/core/base"
	"github.com/liuhailove/seamiter-golang/core/config"
	"github.com/liuhailove/seamiter-golang/core/flow"
	"github.com/liuhailove/seamiter-golang/core/system_metric"
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/liuhailove/seamiter-golang/util"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func initsea() {
	// We should initialize sea first.
	conf := config.NewDefaultConfig()
	// for testing, logging output to console
	conf.Sea.Log.Logger = logging.NewConsoleLogger()
	conf.Sea.Log.Metric.FlushIntervalSec = 0
	conf.Sea.Stat.System.CollectIntervalMs = 0
	conf.Sea.Stat.System.CollectMemoryIntervalMs = 0
	conf.Sea.Stat.System.CollectCpuIntervalMs = 0
	conf.Sea.Stat.System.CollectLoadIntervalMs = 0
	err := api.InitWithConfig(conf)
	if err != nil {
		log.Fatal(err)
	}
}

func TestAdaptiveFlowControl(t *testing.T) {
	initsea()
	util.SetClock(util.NewMockClock())

	rs := "hello0"
	rule := flow.Rule{
		Resource:               rs,
		TokenCalculateStrategy: flow.MemoryAdaptive,
		ControlBehavior:        flow.Reject,
		StatIntervalInMs:       1000,
		LowMemUsageThreshold:   5,
		HighMemUsageThreshold:  1,
		MemLowWaterMarkBytes:   1 * 1024,
		MemHighWaterMarkBytes:  2 * 1024,
	}
	rule1 := rule
	ok, err := flow.LoadRules([]*flow.Rule{&rule1})
	assert.True(t, ok)
	assert.Nil(t, err)
	// mock memory usage < MemLowWaterMarkBytes, QPS threshold is 2
	system_metric.SetSystemMemoryUsage(512)
	for i := 0; i < 5; i++ {
		entry, blockError := api.Entry(rs, api.WithTrafficType(base.Inbound))
		assert.Nil(t, blockError)
		if blockError != nil {
			t.Errorf("entry error: %+v", blockError)
		}
		entry.Exit()
	}

	_, blockError := api.Entry(rs, api.WithTrafficType(base.Inbound))
	assert.NotNil(t, blockError)
	if blockError != nil {
		t.Logf("entry error:%+v, caused: %+v", blockError.Error(), blockError.TriggeredRule())
	}

	// clear statistics
	util.Sleep(time.Second * 2)
	// QPS threshold is 3
	system_metric.SetSystemMemoryUsage(1536)
	for i := 0; i < 3; i++ {
		entry, blockError := api.Entry(rs, api.WithTrafficType(base.Inbound))
		assert.Nil(t, blockError)
		if blockError != nil {
			t.Errorf("entry error:%+v", blockError)
		}
		entry.Exit()
	}
	_, blockError = api.Entry(rs, api.WithTrafficType(base.Inbound))
	assert.NotNil(t, blockError)
	if blockError != nil {
		t.Logf("entry error:%+v, caused: %+v", blockError.Error(), blockError.TriggeredRule())
	}

	// clear statistic
	util.Sleep(time.Second * 2)
	t.Log("start to test memory based adaptive flow control")
	// QPS threshold is 3
	system_metric.SetSystemMemoryUsage(2049)
	for i := 0; i < 1; i++ {
		entry, blockError := api.Entry(rs, api.WithTrafficType(base.Inbound))
		assert.Nil(t, blockError)
		if blockError != nil {
			t.Errorf("entry error:%+v", blockError)
		}
		entry.Exit()
	}
	_, blockError = api.Entry(rs, api.WithTrafficType(base.Inbound))
	assert.NotNil(t, blockError)
	if blockError != nil {
		t.Logf("entry error:%+v, caused: %+v", blockError.Error(), blockError.TriggeredRule())
	}
}

func TestInitWithConfig(t *testing.T) {
	config.InitConfigWithYaml("./sea.yml")
}
