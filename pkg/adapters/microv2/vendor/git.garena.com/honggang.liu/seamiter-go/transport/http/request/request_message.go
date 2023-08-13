package request

import (
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	config2 "git.garena.com/honggang.liu/seamiter-go/transport/common/transport/config"
	"git.garena.com/honggang.liu/seamiter-go/util"
	jsoniter "github.com/json-iterator/go"
	"strconv"
)

// Message
// Request beat message entity.
// The message consists of key-value pair parameters.
type Message struct {
	message map[string]string
}

func NewRspMessage() *Message {
	h := new(Message)
	h.message = make(map[string]string)
	return h
}

func (h *Message) RegisterInformation(key, value string) *Message {
	h.message[key] = value
	return h
}

func (h *Message) GenerateCurrentMessage(reps string) map[string]string {
	h.message["app"] = config.AppName()
	h.message["hostname"] = util.GetHostName()
	h.message["ip"] = config.HeartbeatClintIp()
	// Put application type (since 1.6.0).
	h.message["port"] = strconv.Itoa(config2.GetRuntimePort())
	// metric data
	h.message["reps"] = reps
	return h.message
}

func (h *Message) String() string {
	if h.message == nil {
		return ""
	}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	data, _ := json.Marshal(h.message)
	return string(data)
}
