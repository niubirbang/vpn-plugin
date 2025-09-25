package main

import "C"
import (
	"encoding/json"
	"errors"
)

const (
	CodeOK    = 0
	CodeError = 1
)

var handlers = map[string]Handler{
	"SUCCESS_TEST": func(rb RequestBinder) (interface{}, error) {
		return "Hello", nil
	},
	"ERROR_TEST": func(rb RequestBinder) (interface{}, error) {
		return "This is error data", errors.New("test error")
	},
}

type (
	RequestBinder func(interface{}) error
	Request       struct {
		Action string      `json:"action"` // 操作名
		Params interface{} `json:"params"` // 参数
	}
	Response struct {
		Code int         `json:"code"` // 0=成功，非0=错误
		Msg  string      `json:"msg"`
		Data interface{} `json:"data,omitempty"`
	}
	Handler func(RequestBinder) (interface{}, error)
)

func ParseRequest(cReq *C.char) (*Request, error) {
	var req Request
	if err := json.Unmarshal([]byte(C.GoString(cReq)), &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func NewResponse(data interface{}, err error) *Response {
	code := CodeOK
	msg := "SUCCESS"
	if err != nil {
		code = CodeError
		msg = err.Error()
	}
	return &Response{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

func (r Request) GetBinder() RequestBinder {
	return func(i interface{}) error {
		b, _ := json.Marshal(r.Params)
		return json.Unmarshal(b, i)
	}
}

func (r Response) CResponse() *C.char {
	b, _ := json.Marshal(r)
	return C.CString(string(b))
}

//export Call
func Call(cReq *C.char) *C.char {
	req, err := ParseRequest(cReq)
	if err != nil {
		return NewResponse(nil, err).CResponse()
	}
	handler, ok := handlers[req.Action]
	if !ok {
		return NewResponse(nil, errors.New("action not found")).CResponse()
	}
	data, err := handler(req.GetBinder())
	return NewResponse(data, err).CResponse()
}

func main() {}
