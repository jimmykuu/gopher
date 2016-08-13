package gopher

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jimmykuu/wtforms"
	"gopkg.in/mgo.v2/bson"
)

// URL: /admin/link_exchanges
// 友情链接列表
func adminListLinkExchangesHandler(handler *Handler) {
	c := handler.DB.C(LINK_EXCHANGES)
	var linkExchanges []LinkExchange
	c.Find(nil).All(&linkExchanges)

	handler.renderTemplate("admin/link_exchanges.html", ADMIN, map[string]interface{}{
		"linkExchanges": linkExchanges,
	})
}

// ULR: /admin/link_exchange/new
// 增加友链
func adminNewLinkExchangeHandler(handler *Handler) {
	form := wtforms.NewForm(
		wtforms.NewTextField("name", "名称", "", wtforms.Required{}),
		wtforms.NewTextField("url", "URL", "", wtforms.Required{}, wtforms.URL{}),
		wtforms.NewTextField("description", "描述", "", wtforms.Required{}),
		wtforms.NewTextField("logo", "Logo", ""),
	)

	if handler.Request.Method == "POST" {
		if !form.Validate(handler.Request) {
			handler.renderTemplate("link_exchange/form.html", ADMIN, map[string]interface{}{
				"form":  form,
				"isNew": true,
			})
			return
		}

		c := handler.DB.C(LINK_EXCHANGES)
		var linkExchange LinkExchange
		err := c.Find(bson.M{"url": form.Value("url")}).One(&linkExchange)

		if err == nil {
			form.AddError("url", "该URL已经有了")
			handler.renderTemplate("link_exchange/form.html", ADMIN, map[string]interface{}{
				"form":  form,
				"isNew": true,
			})
			return
		}

		err = c.Insert(&LinkExchange{
			Id_:         bson.NewObjectId(),
			Name:        form.Value("name"),
			URL:         form.Value("url"),
			Description: form.Value("description"),
			Logo:        form.Value("logo"),
			IsOnHome:    handler.Request.FormValue("is_on_home") == "on",
			IsOnBottom:  handler.Request.FormValue("is_on_bottom") == "on",
		})

		if err != nil {
			panic(err)
		}

		http.Redirect(handler.ResponseWriter, handler.Request, "/admin/link_exchanges", http.StatusFound)
		return
	}

	handler.renderTemplate("link_exchange/form.html", ADMIN, map[string]interface{}{
		"form":  form,
		"isNew": true,
	})
}

// URL: /admin/link_exchange/{linkExchangeId}/edit
// 编辑友情链接
func adminEditLinkExchangeHandler(handler *Handler) {
	linkExchangeId := mux.Vars(handler.Request)["linkExchangeId"]

	c := handler.DB.C(LINK_EXCHANGES)
	var linkExchange LinkExchange
	c.Find(bson.M{"_id": bson.ObjectIdHex(linkExchangeId)}).One(&linkExchange)

	form := wtforms.NewForm(
		wtforms.NewTextField("name", "名称", linkExchange.Name, wtforms.Required{}),
		wtforms.NewTextField("url", "URL", linkExchange.URL, wtforms.Required{}, wtforms.URL{}),
		wtforms.NewTextField("description", "描述", linkExchange.Description, wtforms.Required{}),
		wtforms.NewTextField("logo", "Logo", linkExchange.Logo),
	)

	if handler.Request.Method == "POST" {
		if !form.Validate(handler.Request) {
			handler.renderTemplate("link_exchange/form.html", ADMIN, map[string]interface{}{
				"linkExchange": linkExchange,
				"form":         form,
				"isNew":        false,
			})
			return
		}

		err := c.Update(bson.M{"_id": linkExchange.Id_}, bson.M{"$set": bson.M{
			"name":         form.Value("name"),
			"url":          form.Value("url"),
			"description":  form.Value("description"),
			"logo":         form.Value("logo"),
			"is_on_home":   handler.Request.FormValue("is_on_home") == "on",
			"is_on_bottom": handler.Request.FormValue("is_on_bottom") == "on",
		}})

		if err != nil {
			panic(err)
		}

		http.Redirect(handler.ResponseWriter, handler.Request, "/admin/link_exchanges", http.StatusFound)
		return
	}

	handler.renderTemplate("link_exchange/form.html", ADMIN, map[string]interface{}{
		"linkExchange": linkExchange,
		"form":         form,
		"isNew":        false,
	})
}

// URL: /admin/link_exchange/{linkExchangeId}/delete
// 删除友情链接
func adminDeleteLinkExchangeHandler(handler *Handler) {
	linkExchangeId := mux.Vars(handler.Request)["linkExchangeId"]

	c := handler.DB.C(LINK_EXCHANGES)
	c.RemoveId(bson.ObjectIdHex(linkExchangeId))

	handler.ResponseWriter.Write([]byte("true"))
}

// URL: /link
// 友情链接
func linksHandler(handler *Handler) {
	var links []LinkExchange
	c := handler.DB.C(LINK_EXCHANGES)
	c.Find(nil).All(&links)
	handler.renderTemplate("link_exchange/all.html", BASE, map[string]interface{}{
		"links": links,
	})
}
