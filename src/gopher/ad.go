package gopher

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jimmykuu/wtforms"
	"labix.org/v2/mgo/bson"
)

// URL: /admin/ads
// 广告列表
func adminListAdsHandler(w http.ResponseWriter, r *http.Request) {
	var ads []AD
	c := DB.C(ADS)
	c.Find(nil).All(&ads)

	renderTemplate(w, r, "admin/ads.html", ADMIN, map[string]interface{}{
		"ads": ads,
	})
}

// URL: /admin/ad/new
// 添加广告
func adminNewAdHandler(w http.ResponseWriter, r *http.Request) {
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

	if r.Method == "POST" {
		if !form.Validate(r) {
			renderTemplate(w, r, "ad/form.html", ADMIN, map[string]interface{}{
				"form":  form,
				"isNew": true,
			})
			return
		}

		c := DB.C(ADS)
		err := c.Insert(&AD{
			Id_:      bson.NewObjectId(),
			Position: form.Value("position"),
			Name:     form.Value("name"),
			Code:     form.Value("code"),
		})

		if err != nil {
			panic(err)
		}

		http.Redirect(w, r, "/admin/ads", http.StatusFound)
		return
	}

	renderTemplate(w, r, "ad/form.html", ADMIN, map[string]interface{}{
		"form":  form,
		"isNew": true,
	})
}

// URL: /admin/ad/{id}/delete
// 删除广告
func adminDeleteAdHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	c := DB.C(ADS)
	c.RemoveId(bson.ObjectIdHex(id))

	w.Write([]byte("true"))
}

// URL: /admin/ad/{id}/edit
// 编辑广告
func adminEditAdHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	c := DB.C(ADS)
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

	if r.Method == "POST" {
		if !form.Validate(r) {
			renderTemplate(w, r, "ad/form.html", ADMIN, map[string]interface{}{
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

		http.Redirect(w, r, "/admin/ads", http.StatusFound)
		return
	}

	renderTemplate(w, r, "ad/form.html", ADMIN, map[string]interface{}{
		"form":  form,
		"isNew": false,
	})
}
