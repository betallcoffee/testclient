package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type Record struct {
	Begin    time.Time
	End      time.Time
	Elapse  time.Duration
	Error    error
	Response []byte
	ContentSize int64
}

func (r *Record) Start() {
	r.Begin = time.Now()
//	fmt.Println("Record begin", r.Begin)
}

func (r *Record) Stop() {
	r.End = time.Now()
	r.Elapse = r.End.Sub(r.Begin)
//	fmt.Println("Record end", r.End)
}

type HTTPWorker struct {
	tag        int
	start      *sync.WaitGroup
	api        *TestAPI
	client     *http.Client
	jobChan    chan *HTTPJob
	recordChan chan *Record
}

func NewHTTPWorker(tag int, start *sync.WaitGroup, api *TestAPI, jobChan chan *HTTPJob, recordChan chan *Record) *HTTPWorker {
	return &HTTPWorker{
		tag:        tag,
		start:      start,
		api:        api,
		client:     newClient(api),
		jobChan:    jobChan,
		recordChan: recordChan,
	}
}

func newClient(api *TestAPI) *http.Client {
	transport := &http.Transport{
		DisableCompression: false,
		DisableKeepAlives:  api.KeepAlive,
	}

	return &http.Client{Transport: transport}
}

func (w *HTTPWorker) Run() {
//	fmt.Println("HTTPWorker Run", w.tag)
	w.start.Done()
	w.start.Wait()
//	fmt.Println("HTTPWorker Runing", w.tag)
	timer := time.NewTimer(time.Duration(w.api.Timeout)*time.Second)

	for job := range w.jobChan {

		timer.Reset(time.Duration(w.api.Timeout)*time.Second)
		asyncResult := w.send(job.request)

		select {
		case record := <-asyncResult:
//			fmt.Println("HTTPWorker send result", record)
			w.recordChan <- record
		case <-timer.C:
			w.recordChan <- &Record{Error: &ResponseTimeoutError{errors.New("execution timeout")}}
			w.client.Transport.(*http.Transport).CancelRequest(job.request)
		}
		job.Done()
	}
	timer.Stop()
}

func (w *HTTPWorker) send(request *http.Request) chan *Record {
	asyncResult := make(chan *Record, 1)
	go func() {
		record := &Record{}
		record.Start()

		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					record.Error = err
				} else {
					record.Error = &ExceptionError{errors.New(fmt.Sprint(r))}
				}
			}
			asyncResult <- record
		}()

//		fmt.Println("HTTPWorker sending")
		resp, err := w.client.Do(request)
		if err != nil {
//			fmt.Println("HTTPWorker send err", err)
			record.Error = &ConnectError{err}
			return
		}

//		fmt.Println("HTTPWorker send ok")
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode > 300 {
//			fmt.Println("HTTPWork StatusCode err", resp)
			record.Error = &ResponseError{err}
			return
		}
//		fmt.Println("HTTPWorker reading")
		response, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			record.Error = &ReceiveError{err}
			return
		}
//		fmt.Println("HTTPWorker read ok")
		record.Response = response
		record.ContentSize = int64(len(response))
//		fmt.Println(string(response))

		record.Stop()
	}()
	return asyncResult
}
