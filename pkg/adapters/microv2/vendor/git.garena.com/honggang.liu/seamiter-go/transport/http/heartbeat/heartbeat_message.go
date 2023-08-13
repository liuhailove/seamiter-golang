package heartbeat

import (
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	config2 "git.garena.com/honggang.liu/seamiter-go/transport/common/transport/config"
	"git.garena.com/honggang.liu/seamiter-go/util"
	jsoniter "github.com/json-iterator/go"
	"strconv"
)

// Message
// Heart beat message entity.
// The message consists of key-value pair parameters.
type Message struct {
	message map[string]string
}

func NewHeartbeatMessage() *Message {
	h := new(Message)
	h.message = make(map[string]string)
	return h
}

func (h *Message) RegisterInformation(key, value string) *Message {
	h.message[key] = value
	return h
}

func (h *Message) GenerateCurrentMessage() map[string]string {
	h.message["hostname"] = util.GetHostName()
	h.message["ip"] = config.HeartbeatClintIp()
	h.message["app"] = config.AppName()
	// Put application type (since 1.6.0).
	h.message["app_type"] = strconv.Itoa(int(config.AppType()))
	h.message["port"] = strconv.Itoa(config2.GetRuntimePort())
	// Version of sea.
	h.message["v"] = config.Version()
	// Actually timestamp.
	h.message["version"] = strconv.FormatInt(int64(util.CurrentTimeMillis()), 10)
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
