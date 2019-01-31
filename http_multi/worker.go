package http_multi

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"sync/atomic"
	"time"
)

//Worker 一个worker(实际进行网络操作的worker)
type Worker struct {
	config *Config

	id         int
	httpClient *http.Client

	talkTotal    uint64
	successTotal uint64

	startTime time.Time

	workerPool *WorkerPool
}

//NewWorker 创建新worker
func NewWorker(id int, workerPool *WorkerPool) *Worker {
	worker := &Worker{
		id:         id,
		config:     workerPool.config,
		startTime:  time.Now(),
		workerPool: workerPool,
	}

	worker.LogfBase("started")

	go func() {
		for range time.Tick(10 * time.Second) {
			worker.printQPS()
		}
	}()

	return worker
}

//Talk 和远端进行网络交互-同步
func (w *Worker) Talk(req *Request) error {
	atomic.AddUint64(&w.talkTotal, 1)

	return w.TalkWithHTTP(req)
}

//TalkWithHTTP HTTP协议的交互
func (w *Worker) TalkWithHTTP(req *Request) error {
	if w.httpClient == nil {
		//		timeout := time.Duration(w.config.ConnectTimeoutMs+w.config.WriteTimeoutMs+w.config.ReadTimeMs) * time.Millisecond
		timeout := time.Duration(w.config.TimeoutMs) * time.Millisecond
		w.httpClient = &http.Client{
			Timeout: timeout,
		}
	}

	if w.config.Trace {
		bf, _ := httputil.DumpRequest(req.HTTPRequest, true)
		w.Log("dump http request:\n", string(bf))
	}

	tryMax := 1 + w.config.Retry

	var try uint = 0

	var status int
	var body []byte
	var err error

	for ; try < tryMax; try++ {
		status, body, err = w.doRequest(req)
		if err == nil {
			break
		}
	}
	resp := &Response{
		ID:         req.ID,
		URL:        req.URL,
		StatusCode: status,
		RespBody:   string(body),
		Error:      fmt.Sprintf("%v", err),
	}

	saveErr := w.workerPool.saveResponse(resp)

	return saveErr
}

func (w *Worker) doRequest(req *Request) (status int, respBody []byte, err error) {
	resp, err := w.httpClient.Do(req.HTTPRequest)
	if err != nil {
		return -1, nil, err
	}

	defer resp.Body.Close()

	if w.config.Trace {
		bf, _ := httputil.DumpResponse(resp, true)
		w.Log("dump http response:\n", string(bf))
	}

	httpBody, err := ioutil.ReadAll(resp.Body)

	return resp.StatusCode, httpBody, err
}

//LogfBase 打印日志
func (w *Worker) LogfBase(format string, v ...interface{}) {
	log.Printf("worker_id=%d %s",
		w.id,
		fmt.Sprintf(format, v...),
	)
}

//Logf 打印日志
func (w *Worker) Logf(format string, v ...interface{}) {
	w.LogfBase("%s", fmt.Sprintf(format, v...))
}

//Log 打印日志
func (w *Worker) Log(v ...interface{}) {
	w.Logf(fmt.Sprint(v...))
}

//printQPS 打印worker的qps信息
func (w *Worker) getQPS() *qpsInfo {
	num := atomic.LoadUint64(&w.successTotal)
	cost := time.Since(w.startTime)
	qps := float64(num) / float64(cost.Nanoseconds()/1e9)

	return &qpsInfo{
		total:   atomic.LoadUint64(&w.talkTotal),
		success: num,
		qps:     qps,
	}
}

func (w *Worker) printQPS() {
	qps := w.getQPS()
	w.LogfBase("worker_qps_info: total=%d success=%d worker_qps=%.2f", qps.total, qps.success, qps.qps)
}

//Close 回收资源
func (w *Worker) Close() {
	cost := time.Since(w.startTime)
	qps := w.getQPS()
	defer w.LogfBase("stoped,worker_cost=%s %s", cost.String(), qps.String())
}

type qpsInfo struct {
	total   uint64
	success uint64
	qps     float64
}

func (q *qpsInfo) String() string {
	return fmt.Sprintf("total=%d success=%d qps=%.2f", q.total, q.success, q.qps)
}
