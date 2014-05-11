/*
下载
*/
package gopher

import (
	"encoding/json"
	"net/http"
	"os"
)

type File struct {
	Filename string `json:"filename"`
	Summary  string `json:"summary"`
	Size     string `json:"size"`
	SHA1     string `json:"sha1"`
	URL      string `json:"url"`
}

type Version struct {
	Version string `json:"version"`
	Files   []File `json:"files`
	Date    string `json:"date"`
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("etc/download.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	dec := json.NewDecoder(file)

	var versions []Version

	err = dec.Decode(&versions)

	if err != nil {
		panic(err)
	}
	renderTemplate(w, r, "download.html", BASE, map[string]interface{}{"versions": versions, "active": "download"})
}
