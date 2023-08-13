package endpoint

import "strconv"

type Protocol string

const (
	HTTP  Protocol = "http"
	HTTPS Protocol = "https"
)

type Endpoint struct {
	Protocol Protocol // 协议
	Host     string   // 主机
	Port     uint32   // 端口
}

func (e Endpoint) String() string {
	return "Endpoint{" + "protocol=" + string(e.Protocol) + ", host='" + e.Host + ", port=" + strconv.Itoa(int(int32(e.Port))) + "}"
}
