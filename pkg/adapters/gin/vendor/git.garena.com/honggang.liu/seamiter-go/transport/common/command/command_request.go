package command

import (
	"errors"
	"strings"
)

// Request Command request representation of command center.
type Request struct {
	metadata   map[string]string // 元数据信息
	parameters map[string]string // 参数信息
	body       []byte            // 请求体
}

func NewRequest() *Request {
	var req = new(Request)
	req.metadata = make(map[string]string)
	req.parameters = make(map[string]string)
	return req
}

func (c Request) GetBody() []byte {
	return c.body
}

func (c Request) GetParameters() map[string]string {
	return c.parameters
}

func (c Request) GetParam(key string) string {
	return c.parameters[key]
}
func (c Request) GetParamAndDefault(key, defaultValue string) string {
	value := c.parameters[key]
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}

func (c Request) AddParam(key, value string) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("parameter key cannot be empty")
	}
	c.parameters[key] = value
	return nil
}

func (c Request) GetMetadata() map[string]string {
	return c.metadata
}

func (c Request) AddMetaData(key, value string) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("metadata key cannot be empty")
	}
	c.metadata[key] = value
	return nil
}
