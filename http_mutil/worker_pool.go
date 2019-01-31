package http_mutil

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/hidu/goutils/fs"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

//WorkerPool wokrer工作池
type WorkerPool struct {
	config      *Config
	requestChan chan *Request
	workers     []*Worker
	workersFree chan *Worker
	input       *Input
	talkWait    *sync.WaitGroup

	startTime time.Time

	outFile *os.File

	lock *sync.Mutex
}

//NewWorkerPool 创建wokrer工作池对象
func NewWorkerPool() *WorkerPool {
	help := flag.Usage
	flag.Usage = func() {
		help()
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Http Mutil")
		fmt.Fprintln(os.Stderr, "Site: https://github.com/hidu/http_mutil")
		fmt.Fprintln(os.Stderr, "Version: "+Version)
		fmt.Fprintln(os.Stderr, "")
	}

	input := NewInput()
	config := input.config

	flag.StringVar(&config.Input, "input", config.Input, "使用流式输入请求参数,一行一个\n可选值：stdin-使用标准输入；文件地址(eg：data/demo.txt)-从该文件读入\n")
	flag.StringVar(&config.InputFormat, "input_format", config.InputFormat, "输入数据流/文件格式，具体如下："+_InputFormatDesc)
	flag.UintVar(&config.Conc, "conc", config.Conc, "并发数")
	flag.UintVar(&config.Retry, "retry", config.Retry, "重试次数")

	flag.UintVar(&config.TimeoutMs, "timeout", config.TimeoutMs, "超时时间，单位ms")

	//	flag.UintVar(&config.ConnectTimeoutMs, "ctimeout", config.ConnectTimeoutMs, "网络超时-连接，单位ms")
	//	flag.UintVar(&config.WriteTimeoutMs, "wtimeout", config.WriteTimeoutMs, "网络超时-写数据，单位ms")
	//	flag.UintVar(&config.ReadTimeMs, "rtimeout", config.ReadTimeMs, "网络超时-读取数据，单位ms")

	flag.BoolVar(&config.Trace, "trace", config.Trace, "调试模式，会将交互的详细信息打印出来")
	flag.StringVar(&config.LogFileName, "log", config.LogFileName, "日志文件路径；若不需要文件记录日志，可输入固定值【no】")
	flag.StringVar(&config.OutFileName, "out", config.OutFileName, "response输出文件路径")

	flag.Parse()

	log.Println("starting...")

	if err := config.MustParse(); err != nil {
		log.Fatalln("check argvs with error:", err.Error())
	}

	wp := NewWorkerPoolWithConfig(config, input)

	if err := wp.initWithConfig(); err != nil {
		log.Fatalln("init failed:", err.Error())
	}

	return wp
}

//Start 启动进程
func (wp *WorkerPool) Start() {
	log.Println("workerPool start")

	wp.prepare()

	for {
		req, err := wp.input.Next()
		if req != nil {
			wp.requestChan <- req
			//新创建一个异步请求则计数加1
			wp.talkWait.Add(1)
		}

		if err == io.EOF {
			break
		}
	}

	//等待所有的请求都完成
	wp.talkWait.Wait()

	log.Println("workerPool stoped")
}

//NewWorkerPoolWithConfig  创建wokrer工作池对象
func NewWorkerPoolWithConfig(config *Config, intput *Input) *WorkerPool {
	pool := &WorkerPool{
		config:      config,
		requestChan: make(chan *Request, config.RequestQueueSize),
		workers:     make([]*Worker, 0),
		workersFree: make(chan *Worker, config.Conc),
		input:       intput,
		talkWait:    &sync.WaitGroup{},
		startTime:   time.Now(),
		lock:        new(sync.Mutex),
	}
	return pool
}

func (wp *WorkerPool) prepare() {
	log.Println("workers starting ...")

	go wp.input.ParseStream()

	time.Sleep(100 * time.Millisecond)

	for id := 0; id < int(wp.config.Conc); id++ {
		worker := NewWorker(id, wp)
		wp.workers = append(wp.workers, worker)
		wp.workersFree <- worker
	}

	go func() {
		for req := range wp.requestChan {
			worker := <-wp.workersFree
			go func(worker *Worker, req *Request) {

				if err := worker.Talk(req); err != nil {
					//todo
					log.Fatalln("talk with error:", err.Error())
				}

				//完成一个异步请求则计数减1
				wp.talkWait.Done()

				wp.workersFree <- worker
			}(worker, req)
		}

	}()

	go func() {
		for range time.Tick(5 * time.Second) {
			wp.printQPS()
		}
	}()
}

func (wp *WorkerPool) initWithConfig() error {
	log.Println("init: logFileName=", wp.config.LogFileName)

	if wp.config.LogFileName != "no" {
		if err := fs.DirCheck(wp.config.LogFileName); err != nil {
			return err
		}

		logFile, err := os.OpenFile(wp.config.LogFileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
		if err != nil {
			return err
		}

		mWriter := io.MultiWriter(os.Stderr, logFile)
		log.SetOutput(mWriter)
	}

	log.Println("init: outFileName=", wp.config.OutFileName)

	if err := fs.DirCheck(wp.config.OutFileName); err != nil {
		return err
	}

	var fErr error
	wp.outFile, fErr = os.OpenFile(wp.config.OutFileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if fErr != nil {
		return fErr
	}

	return nil
}

func (wp *WorkerPool) saveResponse(resp *Response) error {
	bf, err := resp.Bytes()
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%d\t", resp.StatusCode))
	buf.Write(bf)
	buf.Write([]byte("\n"))

	wp.lock.Lock()
	defer wp.lock.Unlock()
	n, wErr := wp.outFile.Write(buf.Bytes())
	if wErr != nil {
		return wErr
	}

	if n != buf.Len() {
		return fmt.Errorf("expect wrote %d bytes,but wrote %d bytes", buf.Len(), n)
	}

	return nil
}

func (wp *WorkerPool) getQPS() *qpsInfo {
	qps := &qpsInfo{}
	for _, worker := range wp.workers {
		workerQPS := worker.getQPS()
		qps.total += workerQPS.total
		qps.success += workerQPS.success
		qps.qps += workerQPS.qps
	}
	return qps
}
func (wp *WorkerPool) printQPS() {
	qps := wp.getQPS()
	log.Printf("pool_qps_info total=%d success=%d all_qps=%.2f\n", qps.total, qps.success, qps.qps)
}

//Close 资源回收
func (wp *WorkerPool) Close() {
	qps := wp.getQPS()
	for _, worker := range wp.workers {
		worker.Close()
	}
	cost := time.Since(wp.startTime)
	log.Println("all finished,all_cost=", cost.String(), qps.String())
}
