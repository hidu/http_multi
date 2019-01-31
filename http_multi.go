package main

import (
	"github.com/hidu/http_multi/http_multi"
)

func main() {
	wp := http_multi.NewWorkerPool()
	defer wp.Close()
	wp.Start()
}
