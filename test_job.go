package main

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"io"
)

type HTTPJob struct {
	waitJob *sync.WaitGroup
	request *http.Request
}

func NewHTTPJob(waitJob *sync.WaitGroup, api *TestAPI, url string, bodyByte []byte) (*HTTPJob, error) {
	var body io.Reader

	if api.HTTPMethod == "post" || api.HTTPMethod == "put" {
		body = bytes.NewReader(bodyByte)
		//	fmt.Println("new request body", body)
	}

	request, err := http.NewRequest(api.HTTPMethod, url, body)
	if err != nil {
		fmt.Println("new request", err)
		return nil, err
	}
	//	fmt.Println("new request", *request)

	if ProtocolFromString(api.Protocol) == JSONRPC {
		request.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	if api.UserAgent != "" {
		request.Header.Set("User-Agent", api.UserAgent)
	}

	if api.KeepAlive {
		request.Header.Set("Connection", "keep-alive")
	}

	return &HTTPJob{waitJob: waitJob, request: request}, nil
}

func (j *HTTPJob)Done() {
	j.waitJob.Done()
}
