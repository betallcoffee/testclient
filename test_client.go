package main

import (
	"fmt"
	"strconv"
	"sync"
	"runtime"
)

func main() {
	NCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(NCPU)
	fmt.Println("GOMAXPROCS", NCPU)
	profileFilename := "device.prof"
	profile, err := ParseProfileFromFile(profileFilename)
	if err != nil {
		fmt.Println(err)
	}
	client := NewTestClient(profile)
	go client.Test()
	client.Wait()
}

type TestClient struct {
	doneChan chan bool
	profile  *TestProfile
}

func NewTestClient(p *TestProfile) *TestClient {
	return &TestClient{doneChan: make(chan bool), profile: p}
}

func (c *TestClient) Wait() {
	fmt.Println("test client waiting")
	<-c.doneChan
}

func (c *TestClient) Done() {
	fmt.Println("test client done")
	c.doneChan <- true
}

func (c *TestClient) Test() {
	for _, api := range c.profile.API {
		fmt.Println(api)
		switch ProtocolFromString(api.Protocol) {
		case JSONRPC:
			testCases, parseErr := ParseJSONRPCTestCaseFromFile(api.Case)
			if parseErr != nil {
				fmt.Println("parse case", api.Case, parseErr)
				continue
			}
			c.testAPI(&api, testCases)
		}
	}
	c.Done()
}

func (c *TestClient) testAPI(api *TestAPI, testCases []jsonrpcTestCase) {
	recordChan := make(chan *Record, int64(len(testCases))*int64(api.Repeat))
	var waitWorker *sync.WaitGroup
	var waitJob *sync.WaitGroup
	switch NetFromString(api.Net) {
	case HTTP:
		waitWorker, waitJob = c.runHTTPWorker(api, testCases, recordChan, int64(api.Concurrency))
	}

	waitWorker.Add(1)
	testReport := NewTestReport(waitWorker, recordChan)
	fmt.Println("testclient runing report")
	go testReport.Run()
	fmt.Println("testclient waiting jobs")
	waitJob.Wait()
	fmt.Println("testclient done jobs")
	testReport.Quit()
	fmt.Println("testclient quit report")
	testReport.PrintReport(c.profile, api)
	fmt.Println("testclient print report")
}

func (c *TestClient) runHTTPWorker(api *TestAPI, testCases []jsonrpcTestCase, recordChan chan *Record, concurrency int64) (*sync.WaitGroup, *sync.WaitGroup) {
	jobChan := make(chan *HTTPJob, concurrency)
	waitWorker := &sync.WaitGroup{}
	waitWorker.Add(int(concurrency))
	for i := int64(0); i < concurrency; i++ {
		go NewHTTPWorker(int(i), waitWorker, api, jobChan, recordChan).Run()
	}
	var url string
	if int32(c.profile.Port) != 0 {
		url = "http://" + c.profile.Host + ":" + strconv.Itoa(int(c.profile.Port)) + api.Path
	} else {
		url = "http://" + c.profile.Host + api.Path
	}

	waitJob := &sync.WaitGroup{}
	waitJob.Add(len(testCases) * int(api.Repeat))
	fmt.Println("runHTTPWorker url", url)
	for _, testCase := range testCases {
		for i := int64(0); i < int64(api.Repeat); i++ {
			body, bodyErr := testCase.RequestBody()
			if bodyErr != nil {
				continue
			}
			job, err := NewHTTPJob(waitJob, api, url, body)
			if err != nil {
				continue
			}
			jobChan <- job
		}
	}
	return waitWorker, waitJob
}
