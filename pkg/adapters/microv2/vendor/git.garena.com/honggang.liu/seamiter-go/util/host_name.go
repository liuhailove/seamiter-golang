package util

import (
	"fmt"
	"net"
	"os"
)

var host string
var ip string

func init() {
	host, _ = os.Hostname()
	if host == "" {
		host = "unknown"
	}
	ip, _ = getRealIp()
}

// GetIP 获取服务名称
func GetIP() string {
	return ip
}

// GetHostName 获取服务hostname
func GetHostName() string {
	return host
}

// GetConfigString 获取服务配置信息
func GetConfigString() string {
	return `{"machine":"` + host + `","ip":"` + ip + `"}`
}

// 获取应用host
func getRealIp() (ip string, err error) {
	adds, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	var localIPV4 string
	var nonLocalIPV4 string
	for _, addr := range adds {
		if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.To4() != nil {
			if ipNet.IP.IsLoopback() {
				localIPV4 = ipNet.IP.String()
			} else {
				nonLocalIPV4 = ipNet.IP.String()
			}
		}
	}
	if nonLocalIPV4 != "" {
		ip = fmt.Sprintf("%s", nonLocalIPV4)
	} else {
		ip = fmt.Sprintf("%s", localIPV4)
	}
	return
}
