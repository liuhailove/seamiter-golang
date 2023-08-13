package http

import (
	"errors"
	"github.com/liuhailove/seamiter-golang/core/config"
	"github.com/liuhailove/seamiter-golang/logging"
	"github.com/liuhailove/seamiter-golang/transport/common/transport"
	"github.com/liuhailove/seamiter-golang/transport/http/heartbeat"
	"github.com/liuhailove/seamiter-golang/util"
	"runtime"
	"strconv"
	"time"
)

var (
	heartBeatSenderInitFuncInst = new(heartBeatSenderInitFunc)
)

type heartBeatSenderInitFunc struct {
	isInitialized util.AtomicBool
}

func (h heartBeatSenderInitFunc) Initial() error {
	if !h.isInitialized.CompareAndSet(false, true) {
		return nil
	}
	if config.CloseAll() {
		logging.Warn("[fetchRuleInitFunc] WARN: Sdk closeAll is set true")
		return nil
	}
	if !config.OpenConnectDashboard() {
		logging.Warn("[HeartbeatSenderInitFunc] WARN: OpenConnectDashboard is false")
		return nil
	}
	sender := heartbeat.GetHeartbeatSender()
	if sender == nil {
		logging.Warn("[HeartbeatSenderInitFunc] WARN: No HeartbeatSender loaded")
		return errors.New("[HeartbeatSenderInitFunc] WARN: No HeartbeatSender loaded")
	}
	interval := h.retrieveInterval(sender)
	//延迟5s执行，等待配置文件的初始化
	var heartbeatTimer = time.NewTimer(time.Millisecond * 5000)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				var buf [4096]byte
				n := runtime.Stack(buf[:], false)
				logging.Warn("heartBeatSenderInitFunc worker exit from panic", "err", string(buf[:n]))
			}
		}()
		for {
			<-heartbeatTimer.C
			heartbeatTimer.Reset(time.Millisecond * time.Duration(interval)) //interval秒心跳防止过期
			util.Try(func() {
				_, err := sender.SendHeartbeat()
				if err != nil {
					logging.Warn("[HeartbeatSender] Send heartbeat error", "error", err)
				}
			}).CatchAll(func(err error) {
				logging.Error(err, "[HeartbeatSender] WARN: error", "err", err.Error())
			})

		}
	}()
	return nil
}

func (h heartBeatSenderInitFunc) Order() int {
	return -1
}

func (h heartBeatSenderInitFunc) ImmediatelyLoadOnce() error {
	return nil
}

func (h heartBeatSenderInitFunc) retrieveInterval(sender transport.HeartBeatSender) uint64 {
	intervalInConfig := config.HeartBeatIntervalMs()
	if intervalInConfig > 0 {
		logging.Info("[HeartbeatSenderInitFunc] Using heartbeat interval in sea config property: " + strconv.FormatUint(intervalInConfig, 10))
		return intervalInConfig
	}
	senderInterval := sender.IntervalMs()
	logging.Info("[HeartbeatSenderInit] Heartbeat interval not configured in config property or invalid, using sender default: " + strconv.FormatUint(senderInterval, 10))
	return senderInterval
}

func GetHeartBeatSenderInitFuncInst() *heartBeatSenderInitFunc {
	return heartBeatSenderInitFuncInst
}
