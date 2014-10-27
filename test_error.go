package main

type ExceptionError struct {
	err error
}

func (e *ExceptionError) Error() string {
	return e.err.Error()
}

type ResponseError struct {
	err error
}

func (e *ResponseError) Error() string {
	return e.err.Error()
}

type ResponseTimeoutError struct {
	err error
}

func (e *ResponseTimeoutError) Error() string {
	return e.err.Error()
}

type ConnectError struct {
	err error
}

func (e *ConnectError) Error() string {
	return e.err.Error()
}

type ReceiveError struct {
	err error
}

func (e *ReceiveError) Error() string {
	return e.err.Error()
}
