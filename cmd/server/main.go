package main

import (
	"log"

	"github.com/Jedeft/xuanwu/pkg/service"

	"github.com/Jedeft/demo-micro-gw-admin/internal"
)

// @title demo展示
// @version 1.0.1
// @description.markdown
func main() {
	// 这里只是个入口而已，没必要干那么多事情
	log.Fatalln(service.Run(new(internal.Application)))
}
