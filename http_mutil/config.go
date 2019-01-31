package http_mutil

import (
	"fmt"
	"time"
)

const InputFormatURlListGet string = "url_list_get"
const InputFormatJson string = "json"
const Version string = "v1.0 20190131"

var _InputFormatDesc = `
` + InputFormatURlListGet + ` : 每行一条url地址，会以GET请求发出
    eg: http://127.0.0.1:8089/user/get/1
` + InputFormatJson + ` : 每行一条json数据
    eg:` + _defaultJsonRequest.String() + `
`

func NewConfig() *Config {
	config := &Config{}

	config.RequestQueueSize = 1024
	config.Conc = 1
	config.Retry = 3
	config.Input = "stdin"
	config.InputFormat = InputFormatURlListGet

	config.TimeoutMs = 3000

	//	config.ConnectTimeoutMs = 1000
	//	config.ReadTimeMs = 1000
	//	config.WriteTimeoutMs = 1000

	now := time.Now()

	config.LogFileName = fmt.Sprintf("./log/http.log.%s", now.Format("20060102_150405"))

	config.OutFileName = fmt.Sprintf("./data/resp_%s", now.Format("20060102_150405"))

	return config
}

type Config struct {
	//数据源
	Input string

	//数据输入格式
	InputFormat string

	//发送并发数，worker的数量
	Conc uint

	//待发送队列长度
	RequestQueueSize uint32

	//调试模式
	Trace bool

	//超时时间
	TimeoutMs uint

	//	//连接超时
	//	ConnectTimeoutMs uint
	//
	//	WriteTimeoutMs uint
	//
	//	ReadTimeMs uint

	//log前缀
	LogFileName string

	Retry uint

	//内容输出文件
	OutFileName string
}

func (c *Config) MustParse() error {
	if c.LogFileName == "" {
		return fmt.Errorf("log file name is required")
	}

	return nil
}

func (c *Config) SaveOutput(respBody []byte) error {
	return nil
}

func (c *Config) IsSTDIN() bool {
	return c.Input == "stdin"
}
