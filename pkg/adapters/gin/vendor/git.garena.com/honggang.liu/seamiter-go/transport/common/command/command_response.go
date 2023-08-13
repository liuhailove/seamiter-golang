package command

// Response Command response representation of command center.
type Response struct {
	success   bool        // 处理是否成功
	result    interface{} // 处理结果
	exception error       // 异常信息
}

func (c Response) NewCommandResponseWith(result interface{}) *Response {
	return c.NewCommandResponse(result, true, nil)
}
func (c Response) NewCommandResponse(result interface{}, success bool, exception error) *Response {
	c.success = success
	c.result = result
	c.exception = exception
	return &c
}

func (c Response) IsSuccess() bool {
	return c.success
}

func (c Response) GetResult() interface{} {
	return c.result
}

func (c Response) GetException() error {
	return c.exception
}

// OfSuccess
// Construct a successful response with given object.
//
// @param  result object
// @param <T>    type of the result
// @return constructed server response
//
func OfSuccess(result interface{}) *Response {
	cr := new(Response)
	return cr.NewCommandResponseWith(result)
}

//
// OfFailure
// Construct a failed response with given exception.
//
// @param ex cause of the failure
// @return constructed server response
//
func OfFailure(ex error) *Response {
	cr := new(Response)
	return cr.NewCommandResponse(nil, false, ex)
}

//
//OfFailureWith
// Construct a failed response with given exception.
//
// @param ex     cause of the failure
// @param result additional message of the failure
// @return constructed server response
//
func OfFailureWith(ex error, result interface{}) *Response {
	cr := new(Response)
	return cr.NewCommandResponse(nil, false, ex)
}
