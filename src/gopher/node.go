/*
节点
*/

package gopher

import (
	"net/http"

	"github.com/jimmykuu/wtforms"
	"labix.org/v2/mgo/bson"
)

// URL: /nodes
// 列出所有节点及其详细信息
func nodesHandler(handler Handler) {
	var nodes []Node

	c := DB.C(NODES)
	c.Find(nil).Sort("-topiccount").All(&nodes)

	renderTemplate(handler, "node/list.html", BASE, map[string]interface{}{"nodes": nodes})
}

// URL: /admin/node/new
// 新建节点
func adminNewNodeHandler(handler Handler) {
	form := wtforms.NewForm(
		wtforms.NewTextField("id", "ID", "", &wtforms.Required{}),
		wtforms.NewTextField("name", "名称", "", &wtforms.Required{}),
		wtforms.NewTextArea("description", "描述", "", &wtforms.Required{}),
	)

	if handler.Request.Method == "POST" {
		if form.Validate(handler.Request) {
			c := DB.C(NODES)
			node := Node{}

			err := c.Find(bson.M{"id": form.Value("id")}).One(&node)

			if err == nil {
				form.AddError("id", "该ID已经存在")

				renderTemplate(handler, "node/new.html", ADMIN, map[string]interface{}{"form": form})
				return
			}

			err = c.Find(bson.M{"name": form.Value("name")}).One(&node)

			if err == nil {
				form.AddError("name", "该名称已经存在")

				renderTemplate(handler, "node/new.html", ADMIN, map[string]interface{}{"form": form})
				return
			}

			Id_ := bson.NewObjectId()
			err = c.Insert(&Node{
				Id_:         Id_,
				Id:          form.Value("id"),
				Name:        form.Value("name"),
				Description: form.Value("description")})

			if err != nil {
				panic(err)
			}

			http.Redirect(handler.ResponseWriter, handler.Request, "/admin/node/new", http.StatusFound)
		}
	}

	renderTemplate(handler, "node/new.html", ADMIN, map[string]interface{}{"form": form})
}

// URL: /admin/nodes
// 列出所有的节点
func adminListNodesHandler(handler Handler) {
	var nodes []Node
	c := DB.C(NODES)
	c.Find(nil).All(&nodes)
	renderTemplate(handler, "admin/nodes.html", ADMIN, map[string]interface{}{"nodes": nodes})
}
