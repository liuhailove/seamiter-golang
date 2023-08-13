package client

import "github.com/liuhailove/seamiter-golang/transport/common/command"

// CommandClient
//  Basic interface for clients that sending commands
type CommandClient interface {

	// SendCommand
	// Send a command to target destination.
	//
	// @param host    target host
	// @param port    target port
	// @param request command request
	// @return the response from target command server
	// @throws Exception when unexpected error occurs
	//
	SendCommand(host string, port int32, request command.Request) (*command.Response, error)
}
