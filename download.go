/*
下载
*/
package gopher

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	Filename    string `json:"filename"`
	Summary     string `json:"summary"`
	Size        string `json:"size"`
	SHA1        string `json:"sha1"`
	URL         string `json:"url"`
	Recommended bool   `json:"recommended,omitempty"`
}

type Version struct {
	Version string `json:"version"`
	Files   []File `json:"files`
	Date    string `json:"date"`
}

func downloadHandler(handler *Handler) {
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
	handler.renderTemplate("download.html", BASE, map[string]interface{}{"versions": versions, "active": "download"})
}

type LiteIDEFileInfo struct {
	Filename string
	Size     int64 // bytes
}

func (info LiteIDEFileInfo) HumanSize() string {
	if info.Size < 1024 {
		return fmt.Sprintf("%d B", info.Size)
	} else if info.Size < 1024*1024 {
		return fmt.Sprintf("%d K", info.Size/1024)
	} else {
		return fmt.Sprintf("%d M", info.Size/1024/1024)
	}
}

type LiteIDEVersionInfo struct {
	Name  string
	Files []LiteIDEFileInfo
}

func downloadLiteIDEHandler(handler *Handler) {
	versions := []LiteIDEVersionInfo{}

	var version LiteIDEVersionInfo

	first := true
	filepath.Walk("./static/liteide", func(path string, info os.FileInfo, err error) error {
		if path == "./static/liteide" {
			return nil
		}

		temp := strings.Split(path, "/")
		if len(temp) == 3 {
			// 版本文件夹
			if !first {
				versions = append(versions, version)
			} else {
				first = false
			}

			version = LiteIDEVersionInfo{
				Name:  info.Name(),
				Files: []LiteIDEFileInfo{},
			}
		} else if len(temp) == 4 {
			// 文件
			version.Files = append(version.Files, LiteIDEFileInfo{
				Filename: info.Name(),
				Size:     info.Size(),
			})
		}

		fmt.Println(path)
		return nil
	})

	versions = append(versions, version)

	// 倒序排列
	count := len(versions)
	for i := 0; i < count/2; i++ {
		versions[i], versions[count-i-1] = versions[count-i-1], versions[i]
	}

	handler.renderTemplate("download/liteide.html", BASE, map[string]interface{}{"versions": versions, "active": "download"})
}
