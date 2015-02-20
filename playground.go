package gopher

import (
//"github.com/go-on/gopherjslib"
)

func playGroundHandler(handler *Handler) {
	if handler.Method != "POST" {
		handler.render("templates/playground.html")
		return
	}

}
