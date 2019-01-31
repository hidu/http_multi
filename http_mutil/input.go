package http_mutil

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

func NewInput() *Input {
	input := &Input{
		parserHandlers: make(map[string]InputParserFn),
		requestChan:    make(chan *Request, 1024),
		config:         NewConfig(),
	}
	input.RegistPaserFn(InputFormatURlListGet, parserInputFormatURlListGet)
	input.RegistPaserFn(InputFormatJson, parserInputFormatJson)
	return input
}

type InputParserFn func(config *Config, inputLine []byte) (*Request, error)

type Input struct {
	config         *Config
	requestChan    chan *Request
	parserHandlers map[string]InputParserFn
}

func (i *Input) ParserStreamLine(inputLine []byte) (*Request, error) {
	fn, has := i.parserHandlers[i.config.InputFormat]
	if !has {
		return nil, fmt.Errorf("InputFormat not found:%s", i.config.InputFormat)
	}
	return fn(i.config, inputLine)
}

func (i *Input) RegistPaserFn(name string, fn InputParserFn) {
	i.parserHandlers[name] = fn
}

//ParseStream 解析数据流
func (i *Input) ParseStream() {

	i.requestChan = make(chan *Request, 10)

	defer close(i.requestChan)

	rd := os.Stdin

	if !i.config.IsSTDIN() {
		info, err := os.Open(i.config.Input)
		if err != nil {
			log.Println("open stream file ", i.config.Input, " failed,err=", err)
			return
		}
		rd = info
	}

	var lineNo uint64
	buf := bufio.NewReaderSize(rd, 8192)
	for {
		lineNo++
		line, err := buf.ReadBytes('\n')
		if err == io.EOF {
			log.Println("lineNo=", lineNo, "read from stream finished")
			break
		}
		line = bytes.TrimSpace(line)

		if len(line) == 0 {
			log.Println("lineNo=", lineNo, " empty line,skiped")
			continue
		}

		request, err := i.ParserStreamLine(line)

		if err != nil {
			log.Println("lineNo=", lineNo, "data=", string(line), ",build request failed:", err)
		} else {
			request.LineNo = lineNo
			i.requestChan <- request
		}
	}
}

//Next 获取一个请求
func (i *Input) Next() (req *Request, err error) {
	for req := range i.requestChan {
		return req, nil
	}
	return nil, io.EOF
}
