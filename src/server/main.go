package main

import (
	"gopher"
)

func main() {
	go gopher.RssRefresh()
	gopher.StartServer()
}
