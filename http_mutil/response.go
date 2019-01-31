package http_mutil

import (
	"encoding/json"
)

type Response struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	StatusCode int    `json:"status"`
	RespBody   string `json:"body"`
}

func (r *Response) Bytes() ([]byte, error) {
	return json.Marshal(r)
}
