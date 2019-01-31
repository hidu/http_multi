package http_mutil

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Request struct {
	LineNo      uint64
	URL         string
	ID          string
	HTTPRequest *http.Request
}

//解析单行url形式的请求输入
func parserInputFormatURlListGet(config *Config, line []byte) (*Request, error) {
	reqURL := string(line)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	myReq := &Request{
		HTTPRequest: req,
		ID:          req.Host,
		URL:         reqURL,
	}
	return myReq, nil
}

type InputFormatJsonRequest struct {
	ID     string            `json:"id"`
	Method string            `json:"method"`
	URL    string            `json:"url"`
	Header map[string]string `json:"header"`
	Body   string            `json:"body"`
}

func (jr *InputFormatJsonRequest) String() string {
	bf, _ := json.Marshal(jr)
	return string(bf)
}

var _defaultJsonRequest = &InputFormatJsonRequest{
	ID:     "update_uid_1",
	Method: "post",
	URL:    "http://127.0.0.1:8088/user/save/1",
	Header: map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	},
	Body: "name=HanMeiMei&age=12",
}

//解析json格式的请求
func parserInputFormatJson(config *Config, line []byte) (*Request, error) {
	var data *InputFormatJsonRequest
	err := json.Unmarshal(line, &data)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(data.Method, data.URL, strings.NewReader(data.Body))
	if err != nil {
		return nil, err
	}

	if data.Header != nil {
		for k, v := range data.Header {
			req.Header.Add(k, v)
		}
		host := req.Header.Get("Host")
		if host != "" {
			req.Host = host
		}
	}

	myReq := &Request{
		HTTPRequest: req,
		ID:          data.ID,
		URL:         data.URL,
	}
	return myReq, nil

}
