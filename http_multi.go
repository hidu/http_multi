/*
 * Copyright(C) 2020 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2020/4/4
 */

package main

import (
	"github.com/hidu/http_multi/internal"
)

func main() {
	wp := internal.NewWorkerPool()
	defer wp.Close()
	wp.Start()
}
