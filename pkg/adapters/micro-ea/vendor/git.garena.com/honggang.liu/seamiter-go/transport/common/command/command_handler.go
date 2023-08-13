package command

type Handler interface {
	Name() string
	//
	// Desc
	// Get brief description of the command.
	//
	// @return brief description of the command
	// @since 1.5.0
	//
	Desc() string

	//
	// Handle the given Courier command request.
	//
	// @param  the request to handle
	// @return the response
	//
	Handle(request Request) *Response
}
