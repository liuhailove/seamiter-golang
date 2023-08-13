package http

import (
	"errors"
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/transport/common/transport"
	"git.garena.com/honggang.liu/seamiter-go/transport/http/heartbeat"
	"git.garena.com/honggang.liu/seamiter-go/util"
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
		for {
			<-heartbeatTimer.C
			heartbeatTimer.Reset(time.Millisecond * time.Duration(interval)) //interval秒心跳防止过期
			_, err := sender.SendHeartbeat()
			if err != nil {
				logging.Warn("[HeartbeatSender] Send heartbeat error", "error", err)
			}
		}
	}()
	return nil
}

func (h heartBeatSenderInitFunc) Order() int {
	return -1
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
