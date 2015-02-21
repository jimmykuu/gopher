package gopher

import (
	"bytes"
	"golang.org/x/net/websocket"
	"io"
	"strings"

	Lib "github.com/go-on/gopherjslib"
)

type CMD string

const (
	RUN   CMD = "run"
	CLOSE CMD = "close"
)

type Command struct {
	Command CMD
	Content string // 内容
}

type Res struct {
	Content string // 相应内容
	Err     string // 错误
}

// Playground 的静态请求
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

func buildGoCode(input string, buf *bytes.Buffer) error {
	err := Lib.Build(strings.NewReader(input), buf, buildOptions)
	output := buf.Bytes()
	output = output[13 : len(output)-2] // len( "use strict"\n ) == 11 and final ";\n"
	buf.Reset()
	io.Copy(buf, bytes.NewReader(output))
	return err
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
			err := buildGoCode(cmd.Content, &buf)
			if err != nil {
				err = websocket.JSON.Send(ws, Res{
					Err: err.Error(),
				})
			} else {
				err = websocket.JSON.Send(ws, Res{
					Content: buf.String(),
				})
			}
			if err != nil {
				logger.Println(err)
				break
			}
		}
	}
}
