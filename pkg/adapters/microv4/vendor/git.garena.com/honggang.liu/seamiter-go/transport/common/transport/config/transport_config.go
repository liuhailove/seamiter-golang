package config

import (
	"git.garena.com/honggang.liu/seamiter-go/core/config"
	"git.garena.com/honggang.liu/seamiter-go/logging"
	"git.garena.com/honggang.liu/seamiter-go/transport/common/transport/endpoint"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"strconv"
	"strings"
)

const (
	ConsoleServer        = "csp.sea.dashboard.server"
	ServerPort           = "csp.sea.api.port"
	HeartbeatIntervalMs  = "csp.sea.heartbeat.interval.ms"
	HeartbeatClientIp    = "csp.sea.heartbeat.client.ip"
	HeartbeatApiPath     = "csp.sea.heartbeat.api.path"
	HeartbeatDefaultPath = "/registry/machine"
)

var (
	RuntimePort = -1
)

//
// GetHeartbeatIntervalMs
// Get heartbeat interval in milliseconds.
//
// @return heartbeat interval in milliseconds if exists, or null if not configured or invalid config
func GetHeartbeatIntervalMs() uint64 {
	return config.HeartBeatIntervalMs()
}

//
// GetConsoleServerList
// Get a list of Endpoint(protocol, ip/domain, port) indicating sea Dashboard's address.<br>
// NOTE: only support <b>HTTP</b> and <b>HTTPS</b> protocol
//
// @return list of Endpoint(protocol, ip/domain, port). <br>
//        <b>May not be null</b>. <br>
//        An empty list returned when not configured.
//
func GetConsoleServerList() []endpoint.Endpoint {
	consoleServer := config.ConsoleServer()
	var arr []endpoint.Endpoint
	if strings.TrimSpace(consoleServer) == "" {
		return arr
	}
	var consoleArr = strings.Split(consoleServer, ",")
	for _, ipPortStr := range consoleArr {
		if strings.TrimSpace(ipPortStr) == "" {
			continue
		}
		ipPortStr = strings.TrimSpace(ipPortStr)
		port := 80
		protocol := endpoint.HTTP
		if strings.HasPrefix(ipPortStr, "http://") {
			ipPortStr = ipPortStr[7:]
		} else if strings.HasPrefix(ipPortStr, "https://") {
			ipPortStr = ipPortStr[8:]
			port = 443
			protocol = endpoint.HTTPS
		}
		index := strings.Index(ipPortStr, ":")
		if index == 0 {
			// skip
			continue
		}
		host := ipPortStr
		if index >= 0 {
			portParser, err := strconv.Atoi(ipPortStr[index+1:])
			if err != nil {
				logging.Warn("Parse port of dashboard server failed:", "ipPortStr", ipPortStr, "error", err)
				continue
			}
			if portParser <= 1 || portParser >= 65535 {
				logging.Warn("Port number [", portParser, "] over range")
				continue
			}
			host = ipPortStr[0:index]
			port = portParser
		}
		arr = append(arr, endpoint.Endpoint{
			Protocol: protocol,
			Host:     host,
			Port:     uint32(port),
		})
	}
	return arr
}

func GetRuntimePort() int {
	return RuntimePort
}

// GetPort Get Server port of this HTTP server.
func GetPort() string {
	if RuntimePort > 0 {
		return strconv.Itoa(RuntimePort)
	}
	return strconv.Itoa(int(config.ConsolePort()))
}

// SetRuntimePort Set real port this HTTP server uses.
func SetRuntimePort(port int) {
	RuntimePort = port
}

// GetHeartbeatClientIp
// Get heartbeat client local ip.
// If the client ip not configured,it will be the address of local host
func GetHeartbeatClientIp() string {
	ip := config.HeartbeatClintIp()
	if ip == "" {
		ip = util.GetIP()
	}
	return ip
}

// GetHeartbeatApiPath
// Get the heartbeat api path. If the machine registry path of the dashboard
// is modified, then the API path should also be consistent with the API path of the dashboard.
func GetHeartbeatApiPath() string {
	var apiPath = config.HeartbeatApiPath()
	if apiPath == "" {
		return HeartbeatDefaultPath
	}
	if !strings.HasPrefix(apiPath, "/") {
		apiPath = "/" + apiPath
	}
	return apiPath
}
