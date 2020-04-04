package internal

import (
	"encoding/json"
)

type Response struct {
	ID  string `json:"id"`
	URL string `json:"url"`

	// http status code
	StatusCode int    `json:"status"`
	Error      string `json:"error"`
	RespBody   string `json:"body"`

	// 耗时
	Cost int64 `json:"cost_ms"`

	// 请求所在文件行数
	LineNo uint64 `json:"line_no"`
}

func (r *Response) Bytes() ([]byte, error) {
	return json.Marshal(r)
}
