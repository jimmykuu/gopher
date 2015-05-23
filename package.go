/*
第三方包
*/

package gopher

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"code.google.com/p/go.net/websocket"
	"github.com/gorilla/mux"
	"github.com/jimmykuu/wtforms"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// URL: /packages
// 列出最新的一些第三方包
func packagesHandler(handler *Handler) {
	var categories []PackageCategory

	c := handler.DB.C(PACKAGE_CATEGORIES)
	c.Find(nil).All(&categories)

	var latestPackages []Package
	c = handler.DB.C(CONTENTS)
	c.Find(bson.M{"content.type": TypePackage}).Sort("-content.createdat").Limit(10).All(&latestPackages)

	handler.renderTemplate("package/index.html", BASE, map[string]interface{}{
		"categories":     categories,
		"latestPackages": latestPackages,
		"active":         "package",
	})
}

// URL: /package/new
// 新建第三方包
func newPackageHandler(handler *Handler) {
	user, _ := currentUser(handler)

	var categories []PackageCategory

	c := handler.DB.C(PACKAGE_CATEGORIES)
	c.Find(nil).All(&categories)

	var choices []wtforms.Choice

	for _, category := range categories {
		choices = append(choices, wtforms.Choice{Value: category.Id_.Hex(), Label: category.Name})
	}

	form := wtforms.NewForm(
		wtforms.NewHiddenField("html", ""),
		wtforms.NewTextField("name", "名称", "", wtforms.Required{}),
		wtforms.NewSelectField("category_id", "分类", choices, ""),
		wtforms.NewTextField("url", "网址", "", wtforms.Required{}, wtforms.URL{}),
		wtforms.NewTextArea("description", "描述", "", wtforms.Required{}),
	)

	if handler.Request.Method == "POST" && form.Validate(handler.Request) {
		c = handler.DB.C(CONTENTS)
		id := bson.NewObjectId()
		categoryId := bson.ObjectIdHex(form.Value("category_id"))
		html := form.Value("html")
		html = strings.Replace(html, "<pre>", `<pre class="prettyprint linenums">`, -1)
		c.Insert(&Package{
			Content: Content{
				Id_:       id,
				Type:      TypePackage,
				Title:     form.Value("name"),
				Markdown:  form.Value("description"),
				Html:      template.HTML(html),
				CreatedBy: user.Id_,
				CreatedAt: time.Now(),
			},
			Id_:        id,
			CategoryId: categoryId,
			Url:        form.Value("url"),
		})

		c = handler.DB.C(PACKAGE_CATEGORIES)
		// 增加数量
		c.Update(bson.M{"_id": categoryId}, bson.M{"$inc": bson.M{"packagecount": 1}})

		http.Redirect(handler.ResponseWriter, handler.Request, "/p/"+id.Hex(), http.StatusFound)
		return
	}
	handler.renderTemplate("package/form.html", BASE, map[string]interface{}{
		"form":   form,
		"title":  "提交第三方包",
		"action": "/package/new",
		"active": "package",
	})
}

// URL: /package/{packageId}/edit
// 编辑第三方包
func editPackageHandler(handler *Handler) {
	user, _ := currentUser(handler)

	vars := mux.Vars(handler.Request)
	packageId := vars["packageId"]

	if !bson.IsObjectIdHex(packageId) {
		http.NotFound(handler.ResponseWriter, handler.Request)
		return
	}

	package_ := Package{}
	c := handler.DB.C(CONTENTS)
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(packageId), "content.type": TypePackage}).One(&package_)

	if err != nil {
		message(handler, "没有该包", "没有该包", "error")
		return
	}

	if !package_.CanEdit(user.Username, handler.DB) {
		message(handler, "没有权限", "你没有权限编辑该包", "error")
		return
	}

	var categories []PackageCategory

	c = handler.DB.C(PACKAGE_CATEGORIES)
	c.Find(nil).All(&categories)

	var choices []wtforms.Choice

	for _, category := range categories {
		choices = append(choices, wtforms.Choice{Value: category.Id_.Hex(), Label: category.Name})
	}

	form := wtforms.NewForm(
		wtforms.NewHiddenField("html", ""),
		wtforms.NewTextField("name", "名称", package_.Title, wtforms.Required{}),
		wtforms.NewSelectField("category_id", "分类", choices, package_.CategoryId.Hex()),
		wtforms.NewTextField("url", "网址", package_.Url, wtforms.Required{}, wtforms.URL{}),
		wtforms.NewTextArea("description", "描述", package_.Markdown, wtforms.Required{}),
	)

	if handler.Request.Method == "POST" && form.Validate(handler.Request) {
		c = handler.DB.C(CONTENTS)
		categoryId := bson.ObjectIdHex(form.Value("category_id"))
		html := form.Value("html")
		html = strings.Replace(html, "<pre>", `<pre class="prettyprint linenums">`, -1)
		c.Update(bson.M{"_id": package_.Id_}, bson.M{"$set": bson.M{
			"categoryid":        categoryId,
			"url":               form.Value("url"),
			"content.title":     form.Value("name"),
			"content.markdown":  form.Value("description"),
			"content.html":      template.HTML(html),
			"content.updateDBy": user.Id_.Hex(),
			"content.updatedat": time.Now(),
		}})

		c = handler.DB.C(PACKAGE_CATEGORIES)
		if categoryId != package_.CategoryId {
			// 减少原来类别的包数量
			c.Update(bson.M{"_id": package_.CategoryId}, bson.M{"$inc": bson.M{"packagecount": -1}})
			// 增加新类别的包数量
			c.Update(bson.M{"_id": categoryId}, bson.M{"$inc": bson.M{"packagecount": 1}})
		}

		http.Redirect(handler.ResponseWriter, handler.Request, "/p/"+package_.Id_.Hex(), http.StatusFound)
		return
	}

	form.SetValue("html", "")
	handler.renderTemplate("package/form.html", BASE, map[string]interface{}{
		"form":   form,
		"title":  "编辑第三方包",
		"action": "/p/" + packageId + "/edit",
		"active": "package",
	})
}

// URL: /packages/{categoryId}
// 根据类别列出包
func listPackagesHandler(handler *Handler) {
	vars := mux.Vars(handler.Request)
	categoryId := vars["categoryId"]
	c := handler.DB.C(PACKAGE_CATEGORIES)

	category := PackageCategory{}
	err := c.Find(bson.M{"id": categoryId}).One(&category)

	if err != nil {
		message(handler, "没有该类别", "没有该类别", "error")
		return
	}

	var packages []Package

	c = handler.DB.C(CONTENTS)
	c.Find(bson.M{"categoryid": category.Id_, "content.type": TypePackage}).Sort("name").All(&packages)

	var categories []PackageCategory

	c = handler.DB.C(PACKAGE_CATEGORIES)
	c.Find(nil).All(&categories)

	handler.renderTemplate("package/list.html", BASE, map[string]interface{}{
		"categories": categories,
		"packages":   packages,
		"category":   category,
		"active":     "package",
	})
}

// URL: /p/{packageId}
// 显示第三方包详情
func showPackageHandler(handler *Handler) {
	vars := mux.Vars(handler.Request)

	packageId := vars["packageId"]

	if !bson.IsObjectIdHex(packageId) {
		http.NotFound(handler.ResponseWriter, handler.Request)
		return
	}

	c := handler.DB.C(CONTENTS)

	package_ := Package{}
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(packageId), "content.type": TypePackage}).One(&package_)

	if err != nil {
		message(handler, "没找到该包", "请检查链接是否正确", "error")
		fmt.Println("showPackageHandler:", err.Error())
		return
	}

	var categories []PackageCategory

	c = handler.DB.C(PACKAGE_CATEGORIES)
	c.Find(nil).All(&categories)

	handler.renderTemplate("package/show.html", BASE, map[string]interface{}{
		"package":    package_,
		"categories": categories,
		"active":     "package",
	})
}

// URL: /p/{packageId}/delete
// 删除第三方包
func deletePackageHandler(handler *Handler) {
	vars := mux.Vars(handler.Request)
	packageId := bson.ObjectIdHex(vars["packageId"])

	c := handler.DB.C(CONTENTS)

	package_ := Package{}
	err := c.Find(bson.M{"_id": packageId, "content.type": TypePackage}).One(&package_)

	if err != nil {
		return
	}

	c.Remove(bson.M{"_id": packageId})

	// 修改分类下的数量
	c = handler.DB.C(PACKAGE_CATEGORIES)
	c.Update(bson.M{"_id": package_.CategoryId}, bson.M{"$inc": bson.M{"packagecount": -1}})

	http.Redirect(handler.ResponseWriter, handler.Request, "/packages", http.StatusFound)
}

// URL: /download/package
// 下载第三方包
func downloadPackagesHandler(handler *Handler) {
	var packages []DownloadedPackage
	c := handler.DB.C(DOWNLOADED_PACKAGES)
	c.Find(nil).Sort("-count").Limit(20).All(&packages)
	handler.renderTemplate("package/download.html", BASE, map[string]interface{}{
		"packages": packages,
		"active":   "package-download",
	})
}

// 用于接收Command的输出
type ConsoleWriter struct {
	ws       *websocket.Conn
	packages []string
}

type Message struct {
	Type string `json:"type"`
	Msg  string `json:"msg"`
}

func NewConsoleWriter(ws *websocket.Conn) *ConsoleWriter {
	return &ConsoleWriter{ws: ws}
}

type DownloadedPackage struct {
	Name  string `bson:"name"`
	Count int    `bson:"count"`
}

func (cw *ConsoleWriter) Write(p []byte) (n int, err error) {
	line := strings.Trim(string(p), "\n")

	if strings.LastIndex(line, " (download)") == len(line)-len(" (download)") {
		packageName := line[:len(line)-len(" (download)")]

		cw.packages = append(cw.packages, packageName)
		message := Message{
			Type: "output",
			Msg:  line,
		}

		err = websocket.JSON.Send(cw.ws, message)
	} else if strings.Index(line, "# cd") == 0 {
		// 出错
		err = websocket.JSON.Send(cw.ws, Message{
			Type: "error",
			Msg:  line,
		})
	} else {
		// 一些输出提示
		message := Message{
			Type: "output",
			Msg:  line,
		}

		err = websocket.JSON.Send(cw.ws, message)
	}

	n = len(p)

	return
}

// URL: ws://.../get/package
// 和页面WebSocket通信
func getPackageHandler(ws *websocket.Conn) {
	defer dps.Persist()
	defer ws.Close()

	var err error

	for {
		var name string

		if err = websocket.Message.Receive(ws, &name); err != nil {
			fmt.Println("can't receive")
			break
		}

		fmt.Println("received back from client:", name)

		cmd := exec.Command("go", "get", "-u", "-v", name)
		cmd.Env = os.Environ()[:]
		cmd.Env = append(cmd.Env, "GOPATH="+Config.GoGetPath)

		writer := NewConsoleWriter(ws)

		cmd.Stdout = writer
		cmd.Stderr = writer
		err := cmd.Start()
		if err != nil {
			fmt.Println(err)
			break
		}

		err = cmd.Wait()
		if err != nil {
			fmt.Println(err)
			break
		}

		// 压缩
		for _, packageName := range writer.packages {
			tarFilename := strings.Replace(packageName, "/", ".", -1) + ".tar.gz"
			message := Message{
				Type: "command",
				Msg:  fmt.Sprintf("tar %s %s", tarFilename, packageName),
			}
			websocket.JSON.Send(ws, message)

			cmd := exec.Command("tar", "czf", filepath.Join(Config.PackagesDownloadPath, tarFilename), "--exclude-vcs", packageName)

			cmd.Dir = filepath.Join(Config.GoGetPath, "src")
			err = cmd.Run()
			if err != nil {
				panic(err)
			}

			// 发送可以可以下载的package
			websocket.JSON.Send(ws, Message{
				Type: "download",
				Msg:  packageName,
			})
		}

		websocket.JSON.Send(ws, Message{
			Type: "completed",
			Msg:  "------Done------",
		})

		session, err := mgo.Dial(Config.DB)
		if err != nil {
			panic(err)
		}

		session.SetMode(mgo.Monotonic, true)

		defer session.Close()

		DB := session.DB("gopher")

		c := DB.C(DOWNLOADED_PACKAGES)

		err = c.Find(bson.M{"name": name}).One(nil)
		if err == nil {
			c.Update(bson.M{"name": name}, bson.M{"$inc": bson.M{"count": 1}})
		} else {
			c.Insert(&DownloadedPackage{
				Name:  name,
				Count: 1,
			})
		}
		break
	}
}

// URL: /package?name=github.com/jimmykuu/webhelpers
// 返回第三方包压缩包的下载地址
func getPackageUrlHandler(handler *Handler) {
	packageName := handler.Request.FormValue("name")

	c := handler.DB.C(DOWNLOADED_PACKAGES)
	err := c.Find(bson.M{"name": packageName}).One(nil)

	if err == nil {
		// 找到压缩包并返回下载链接
		tarFilename := strings.Replace(packageName, "/", ".", -1) + ".tar.gz"

		// 检查是否存在
		_, err = os.Stat(filepath.Join("static", "download", "packages", tarFilename))
		if err != nil {
			handler.notFound()
			return
		}

		handler.renderText(Config.Host + "/static/download/packages/" + tarFilename)
	} else {
		c.Insert(&DownloadedPackage{
			Name:  packageName,
			Count: 0,
		})

		handler.notFound()
	}

}
