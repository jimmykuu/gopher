package gopher

import (
	"bytes"
	"golang.org/x/net/websocket"
	"io"
	"strings"

	Lib "github.com/go-on/gopherjslib"
	"gopkg.in/mgo.v2/bson"
)

type CMD string

const (
	RUN    CMD = "run"
	CLOSE  CMD = "close"
	SHARE  CMD = "share"
	UPDATE CMD = "update"
)

type Command struct {
	Command CMD
	Id      string
	Content string // 内容
}

type Res struct {
	Command CMD    //
	Content string // 相应内容
	Err     string // 错误
}

// Playground 的静态请求
func playGroundHandler(handler *Handler) {
	if handler.Method != "POST" {
		id := handler.param("id")

		var (
			code    *Code
			err     error
			content string = "这是Go语言游乐场,可以用来分享代码"
		)
		if id != "" {
			code, err = GetCodeById(id, handler.DB)
			content = code.Content
			if err != nil {
				logger.Println(err)
			}
		}
		handler.render("templates/playground.html", map[string]interface{}{
			"Code":   content,
			"CodeId": id,
			"Host":   Config.Host,
		})
		return
	}
}

// share code
func shareCodeHandler(handler *Handler) {

}

// 封装websocket
func playSocketHandler(handler *Handler) {
	websocket.Handler(playWebSocket(handler)).
		ServeHTTP(handler.ResponseWriter, handler.Request)
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
	if err != nil {
		return err
	}
	output := buf.Bytes()
	output = output[13 : len(output)-2] // len( "use strict"\n ) == 11 and final ";\n"
	buf.Reset()
	io.Copy(buf, bytes.NewReader(output))
	return nil
}

// socket 处理主体
func playWebSocket(handler *Handler) func(ws *websocket.Conn) {
	return func(ws *websocket.Conn) {
		cmd := new(Command)
		for {

			err := websocket.JSON.Receive(ws, cmd)
			if err != nil {
				logger.Println(err)
				break
			}

			if cmd.Command == CLOSE {
				break
			}
			if cmd.Command == UPDATE {
				println("update")
				var id = bson.ObjectIdHex(cmd.Id)

				code := &Code{
					Id_: id,
				}
				err := code.Update(handler.DB, bson.M{
					"content": cmd.Content,
				})
				if err != nil {
					logger.Println(err)
					websocket.JSON.Send(ws, Res{
						Command: SHARE,
						Err:     err.Error(),
					})
					break
				}
				websocket.JSON.Send(ws, Res{
					Command: SHARE,
					Content: id.Hex(),
				})
			}

			if cmd.Command == SHARE {
				id := bson.NewObjectId()
				code := &Code{
					Id_:     id,
					Content: cmd.Content,
				}
				err := code.Save(handler.DB)
				if err != nil {
					logger.Println(err)
					websocket.JSON.Send(ws, Res{
						Command: SHARE,
						Err:     err.Error(),
					})
					break
				}
				websocket.JSON.Send(ws, Res{
					Command: SHARE,
					Content: id.Hex(),
				})
			}

			if cmd.Command == RUN {
				var buf bytes.Buffer
				err := buildGoCode(cmd.Content, &buf)
				if err != nil {
					err = websocket.JSON.Send(ws, Res{
						Command: RUN,
						Err:     err.Error(),
					})
				} else {
					err = websocket.JSON.Send(ws, Res{
						Command: RUN,
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
}
