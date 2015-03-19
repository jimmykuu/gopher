package main

import (
	"github.com/jimmykuu/gopher"
)

func main() {
	go gopher.RssRefresh()
	gopher.StartServer()
}
