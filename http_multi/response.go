package http_multi

import (
	"encoding/json"
)

type Response struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	StatusCode int    `json:"status"`
	Error      string `json"error"`
	RespBody   string `json:"body"`
}

func (r *Response) Bytes() ([]byte, error) {
	return json.Marshal(r)
}
