package gopher

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jimmykuu/wtforms"
	"gopkg.in/mgo.v2/bson"
)

// URL: /admin/ads
// 广告列表
func adminListAdsHandler(handler Handler) {
	var ads []AD
	c := handler.DB.C(ADS)
	c.Find(nil).All(&ads)

	renderTemplate(handler, "admin/ads.html", ADMIN, map[string]interface{}{
		"ads": ads,
	})
}

// URL: /admin/ad/new
// 添加广告
func adminNewAdHandler(handler Handler) {
	choices := []wtforms.Choice{
		wtforms.Choice{"frongpage", "首页"},
		wtforms.Choice{"2cols", "2列宽度"},
		wtforms.Choice{"3cols", "3列宽度"},
		wtforms.Choice{"4cols", "4列宽度"},
	}
	form := wtforms.NewForm(
		wtforms.NewSelectField("position", "位置", choices, "", wtforms.Required{}),
		wtforms.NewTextField("name", "名称", "", wtforms.Required{}),
		wtforms.NewTextArea("code", "代码", "", wtforms.Required{}),
	)

	if handler.Request.Method == "POST" {
		if !form.Validate(handler.Request) {
			renderTemplate(handler, "ad/form.html", ADMIN, map[string]interface{}{
				"form":  form,
				"isNew": true,
			})
			return
		}

		c := handler.DB.C(ADS)
		err := c.Insert(&AD{
			Id_:      bson.NewObjectId(),
			Position: form.Value("position"),
			Name:     form.Value("name"),
			Code:     form.Value("code"),
		})

		if err != nil {
			panic(err)
		}

		http.Redirect(handler.ResponseWriter, handler.Request, "/admin/ads", http.StatusFound)
		return
	}

	renderTemplate(handler, "ad/form.html", ADMIN, map[string]interface{}{
		"form":  form,
		"isNew": true,
	})
}

// URL: /admin/ad/{id}/delete
// 删除广告
func adminDeleteAdHandler(handler Handler) {
	id := mux.Vars(handler.Request)["id"]

	c := handler.DB.C(ADS)
	c.RemoveId(bson.ObjectIdHex(id))

	handler.ResponseWriter.Write([]byte("true"))
}

// URL: /admin/ad/{id}/edit
// 编辑广告
func adminEditAdHandler(handler Handler) {
	id := mux.Vars(handler.Request)["id"]

	c := handler.DB.C(ADS)
	var ad AD
	c.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&ad)

	choices := []wtforms.Choice{
		wtforms.Choice{"frongpage", "首页"},
		wtforms.Choice{"3cols", "3列宽度"},
		wtforms.Choice{"4cols", "4列宽度"},
	}
	form := wtforms.NewForm(
		wtforms.NewSelectField("position", "位置", choices, ad.Position, wtforms.Required{}),
		wtforms.NewTextField("name", "名称", ad.Name, wtforms.Required{}),
		wtforms.NewTextArea("code", "代码", ad.Code, wtforms.Required{}),
	)

	if handler.Request.Method == "POST" {
		if !form.Validate(handler.Request) {
			renderTemplate(handler, "ad/form.html", ADMIN, map[string]interface{}{
				"form":  form,
				"isNew": false,
			})
			return
		}

		err := c.Update(bson.M{"_id": ad.Id_}, bson.M{"$set": bson.M{
			"position": form.Value("position"),
			"name":     form.Value("name"),
			"code":     form.Value("code"),
		}})

		if err != nil {
			panic(err)
		}

		http.Redirect(handler.ResponseWriter, handler.Request, "/admin/ads", http.StatusFound)
		return
	}

	renderTemplate(handler, "ad/form.html", ADMIN, map[string]interface{}{
		"form":  form,
		"isNew": false,
	})
}
