package http

import (
	"errors"
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	_ "git.garena.com/honggang.liu/seamiter-go/transport/common/command/handler" // 强制初始化
	"git.garena.com/honggang.liu/seamiter-go/transport/http/command"
	"git.garena.com/honggang.liu/seamiter-go/util"
)

var (
	commandCenterInitFuncInst = new(commandCenterInitFunc)
)

type commandCenterInitFunc struct {
	isInitialized util.AtomicBool
}

func (c commandCenterInitFunc) Initial() error {
	if !c.isInitialized.CompareAndSet(false, true) {
		return nil
	}
	if config.CloseAll() {
		logging.Warn("[fetchRuleInitFunc] WARN: Sdk closeAll is set true")
		return nil
	}
	var commandCenter = command.GetCommandCenter()
	if commandCenter == nil {
		logging.Warn("[CommandCenterInitFunc] Cannot resolve CommandCenter")
		return errors.New("[CommandCenterInitFunc] Cannot resolve CommandCenter")
	}
	err := commandCenter.BeforeStart()
	if err != nil {
		return err
	}
	err = commandCenter.Start()
	if err != nil {
		return err
	}
	return nil
}

func (c commandCenterInitFunc) Order() int {
	return -1
}

func GetCommandCenterInitFuncInst() *commandCenterInitFunc {
	return commandCenterInitFuncInst
}
