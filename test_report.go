package main

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

type TestReport struct {
	waitWorker       *sync.WaitGroup
	recordChan       chan *Record
	quitChan         chan bool
	doneChan         chan bool
	begin            time.Time
	end              time.Time
	responseTimeData []time.Duration

	totalRequests       int
	totalExecutionTime  time.Duration
	totalResponseTime   time.Duration
	totalReceived       int64
	totalFailedReqeusts int

	errConnect   int
	errTimeout   int
	errReceive   int
	errException int
	errResponse  int
}

func NewTestReport(waitWorker *sync.WaitGroup, recordChan chan *Record) *TestReport {
	return &TestReport{
		waitWorker: waitWorker,
		recordChan: recordChan,
		quitChan:   make(chan bool),
		doneChan:   make(chan bool),
	}
}

func (r *TestReport) Start() {
	fmt.Println("test report starting")
	r.waitWorker.Wait()
	fmt.Println("test report start")
	r.begin = time.Now()
}

func (r *TestReport) Stop() {
	fmt.Println("test report stop")
	r.end = time.Now()
	r.totalExecutionTime = r.end.Sub(r.begin)
	fmt.Println("test report totalExecutionTime", r.totalExecutionTime)
}

func (r *TestReport) Quit() {
	r.quitChan <- true
}

func (r *TestReport) Wait() {
	<-r.doneChan
}

func (r *TestReport) Done() {
	r.doneChan <- true
}

func (r *TestReport) Run() {
	r.waitWorker.Done()
	r.waitWorker.Wait()
	r.Start()
	defer func() {
		r.Stop()
		r.Done()
	}()
	for {
		select {
		case record := <-r.recordChan:
			//			fmt.Println("test report run", record)
			r.updateReport(record)
		case <-r.quitChan:
			fmt.Println("test report quit")
			return
		}
	}
}

func (r *TestReport) updateReport(record *Record) {
	r.totalRequests += 1

	if record.Error != nil {
		r.totalFailedReqeusts += 1

		switch record.Error.(type) {
		case *ConnectError:
			r.errConnect += 1
		case *ResponseTimeoutError:
			r.errTimeout += 1
		case *ExceptionError:
			fmt.Println(record.Error)
			r.errException += 1
		case *ReceiveError:
			r.errReceive += 1
		case *ResponseError:
			r.errResponse += 1
		default:
			fmt.Println(record.Error)
			r.errException += 1
		}

	} else {
		r.totalResponseTime += record.Elapse
		r.totalReceived += record.ContentSize
		r.responseTimeData = append(r.responseTimeData, record.Elapse)
	}
}

func (r *TestReport) PrintReport(profile *TestProfile, api *TestAPI) {
	var buffer bytes.Buffer

	responseTimeData := r.responseTimeData
	totalFailedReqeusts := r.totalFailedReqeusts
	totalRequests := r.totalRequests
	//	totalExecutionTime := r.totalExecutionTime
	totalReceived := r.totalReceived

	fmt.Fprint(&buffer, "\n\n")
	fmt.Fprintf(&buffer, "Server Hostname:        %s\n", profile.Host)
	fmt.Fprintf(&buffer, "Server Port:            %d\n\n", int(profile.Port))

	fmt.Fprintf(&buffer, "Document Path:          %s\n", api.Path)

	fmt.Fprintf(&buffer, "Concurrency Level:      %d\n", int(api.Concurrency))
	fmt.Fprintf(&buffer, "Execution time for tests:   %.2f seconds\n", r.totalExecutionTime)
	fmt.Fprintf(&buffer, "Complete requests:      %d\n", totalRequests)
	if totalFailedReqeusts == 0 {
		fmt.Fprintln(&buffer, "Failed requests:        0")
	} else {
		fmt.Fprintf(&buffer, "Success requests:        %d\n", totalRequests-totalFailedReqeusts)
		fmt.Fprintf(&buffer, "Failed requests:        %d\n", totalFailedReqeusts)
		fmt.Fprintf(&buffer, "   (Connect: %d, Receive: %d, Timeout: %d, Exceptions: %d)\n", r.errConnect, r.errReceive, r.errTimeout, r.errException)
	}
	if r.errResponse > 0 {
		fmt.Fprintf(&buffer, "Non-2xx responses:      %d\n", r.errResponse)
	}
	fmt.Fprintf(&buffer, "HTML transferred:       %d bytes\n", totalReceived)

	if len(responseTimeData) > 0 {
		stdDevOfResponseTime := stdDev(responseTimeData) / float64(time.Millisecond)
		sort.Sort(durationSlice(responseTimeData))

		meanOfResponseTime := int64(r.totalExecutionTime) / int64(totalRequests) / int64(time.Millisecond)
		medianOfResponseTime := responseTimeData[len(responseTimeData)/2] / time.Millisecond
		minResponseTime := responseTimeData[0] / time.Millisecond
		maxResponseTime := responseTimeData[len(responseTimeData)-1] / time.Millisecond

		fmt.Fprintf(&buffer, "Requests per second:    %.2f [#/sec] (mean)\n", float64(totalRequests)/r.totalExecutionTime.Seconds())
		fmt.Fprintf(&buffer, "Time per request:       %.3f [ms] (mean)\n", float64(api.Concurrency)*float64(r.totalExecutionTime.Nanoseconds())/float64(time.Millisecond)/float64(totalRequests))
		fmt.Fprintf(&buffer, "Time per request:       %.3f [ms] (mean, across all concurrent requests)\n", float64(r.totalExecutionTime.Nanoseconds())/float64(time.Millisecond)/float64(totalRequests))
		fmt.Fprintf(&buffer, "Data Transfer rate:     %.2f [Kbytes/sec] received\n\n", float64(totalReceived/1024)/r.totalExecutionTime.Seconds())

		fmt.Fprint(&buffer, "Connection Times (ms)\n")
		fmt.Fprint(&buffer, "              min\tmean[+/-sd]\tmedian\tmax\n")
		fmt.Fprintf(&buffer, "Total:        %d     \t%d   %.2f \t%d \t%d\n\n",
			minResponseTime,
			meanOfResponseTime,
			stdDevOfResponseTime,
			medianOfResponseTime,
			maxResponseTime)

		fmt.Fprintln(&buffer, "Percentage of the requests served within a certain time (ms)")

		percentages := []int{50, 66, 75, 80, 90, 95, 98, 99}

		for _, percentage := range percentages {
			fmt.Fprintf(&buffer, " %d%%\t %d\n", percentage, responseTimeData[percentage*len(responseTimeData)/100]/1000000)
		}
		fmt.Fprintf(&buffer, " %d%%\t %d (longest request)\n", 100, maxResponseTime)
	}
	fmt.Println(buffer.String())
}

type durationSlice []time.Duration

func (s durationSlice) Len() int           { return len(s) }
func (s durationSlice) Less(i, j int) bool { return s[i] < s[j] }
func (s durationSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// StdDev calculate standard deviation
func stdDev(data []time.Duration) float64 {
	var sum int64
	for _, d := range data {
		sum += int64(d)
	}
	avg := float64(sum / int64(len(data)))

	sumOfSquares := 0.0
	for _, d := range data {

		sumOfSquares += math.Pow(float64(d)-avg, 2)
	}
	return math.Sqrt(sumOfSquares / float64(len(data)))

}
