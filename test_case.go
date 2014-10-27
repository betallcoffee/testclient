package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type TestCase interface {
	RequestBody() []byte
}
type jsonrpcRequest struct {
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params"`
	ID     int32            `json:"id"`
}

type jsonrpcResponse struct {
	Result interface{}   `json:"result"`
	Error  *jsonrpcError `json:"error"`
	ID     int32         `json:"id"`
}

type jsonrpcError struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

type jsonrpcTestCase struct {
	Request  jsonrpcRequest
	Response jsonrpcResponse
}

func ParseJSONRPCTestCaseFromFile(filename string) ([]jsonrpcTestCase, error) {
	readByte, readErr := ioutil.ReadFile(filename)
	if readErr != nil {
		fmt.Println("read case", filename, readErr)
		return nil, readErr
	}
	var apiCases []jsonrpcTestCase
	jsonErr := json.Unmarshal(readByte, &apiCases)
	if jsonErr != nil {
		fmt.Println("json unmarshal case", jsonErr)
		return nil, jsonErr
	}
	return apiCases, nil
}

func (c *jsonrpcTestCase) RequestBody() ([]byte, error) {
	body, err := json.Marshal(c.Request)
	if err != nil {
		fmt.Println("requestBody", err)
		return nil, err
	}
	return body, nil
}
