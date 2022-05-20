package main

import (
	"github.com/lihao20110/simple-douyin/server/initialize"
)

func main() {
	r := initialize.InitRouters()
	// public directory is used to serve static resources
	r.Static("/static", "./public")

	r.Run(":8080")
}
