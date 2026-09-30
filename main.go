package main

import (
	"github.com/beego/beego/v2/server/web"

	_ "event-explorer/routers"
)

func main() {
	web.SetStaticPath("/static", "static")
	web.Run()
}