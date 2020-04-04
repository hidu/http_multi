package internal

import (
	"fmt"
	"time"
)

const (
	// InputFormatURLListGet  输入格式为url列表，一行一个url
	InputFormatURLListGet string = "url_list_get"

	// InputFormatJSON 输入格式为json，一行一条json数据
	InputFormatJSON string = "json"
)

// Version 版本信息
const Version string = "v1.0.2020040414"

var _InputFormatDesc = `
` + InputFormatURLListGet + ` : 每行一条url地址，会以GET请求发出
    eg: http://127.0.0.1:8089/user/get/1
` + InputFormatJSON + ` : 每行一条json数据
    eg:` + _defaultJSONRequest.String() + `
`

func newConfig() *Config {
	config := &Config{}

	config.RequestQueueSize = 1024
	config.Conc = 1
	config.Retry = 3
	config.Input = "stdin"
	config.InputFormat = InputFormatURLListGet

	config.TimeoutMs = 3000

	// 	config.ConnectTimeoutMs = 1000
	// 	config.ReadTimeMs = 1000
	// 	config.WriteTimeoutMs = 1000

	now := time.Now()

	config.LogFileName = fmt.Sprintf("./log/http.log.%s", now.Format("20060102_150405"))

	config.OutFileName = fmt.Sprintf("./data/resp_%s", now.Format("20060102_150405"))

	return config
}

// Config 程序配置
type Config struct {
	// 数据源
	Input string

	// 数据输入格式
	InputFormat string

	// 发送并发数，worker的数量
	Conc uint

	// 待发送队列长度
	RequestQueueSize uint32

	// 调试模式
	Trace bool

	// 超时时间
	TimeoutMs uint

	// 	//连接超时
	// 	ConnectTimeoutMs uint
	//
	// 	WriteTimeoutMs uint
	//
	// 	ReadTimeMs uint

	// log前缀
	LogFileName string

	Retry uint

	// 内容输出文件
	OutFileName string
}

// MustParse 解析、判断配置是否正确
func (c *Config) MustParse() error {
	if c.LogFileName == "" {
		return fmt.Errorf("log file name is required")
	}

	return nil
}

// func (c *Config) SaveOutput(respBody []byte) error {
// 	return nil
// }

// IsSTDIN 是否从标准输入读取信息
func (c *Config) IsSTDIN() bool {
	return c.Input == "stdin"
}
