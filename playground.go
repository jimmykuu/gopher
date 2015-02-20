package gopher

import (
	"bytes"
	"golang.org/x/net/websocket"
	"strings"

	Lib "github.com/go-on/gopherjslib"
)

type Command struct {
	Command string
	Content string //命令
}

type Res struct {
	Content string //相应内容
	Err     string //错误
}

// playground 的静态请求
func playGroundHandler(handler *Handler) {
	if handler.Method != "POST" {
		handler.render("templates/playground.html")
		return
	}

}

// 封装websocket
func playSocketHandler(handler *Handler) {
	websocket.Handler(playWebSocket).
		ServeHTTP(handler.ResponseWriter,
		handler.Request)
}

// 封装错误输出
func catch(err error) {
}

var buildOptions = new(Lib.Options)

func init() {
	buildOptions.Minify = false
}

// socket 处理主体
func playWebSocket(ws *websocket.Conn) {
	cmd := new(Command)
	for {
		err := websocket.JSON.Receive(ws, cmd)
		if err != nil {
			logger.Println(err)
			break
		}
		if cmd.Command == "close" {
			break
		}
		if cmd.Command == "run" {
			var buf bytes.Buffer
			Lib.Build(strings.NewReader(cmd.Content), &buf, buildOptions)
			err = websocket.JSON.Send(ws, Res{
				Content: buf.String(),
			})
			if err != nil {
				logger.Println(err)
				break
			}
		}

	}
}
