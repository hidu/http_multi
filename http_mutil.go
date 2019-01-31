package main

import (
	"github.com/hidu/http_mutil/http_mutil"
)

func main() {
	wp := http_mutil.NewWorkerPool()
	defer wp.Close()
	wp.Start()
}
