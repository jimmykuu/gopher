/*
下载
*/
package gopher

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mcuadros/go-version"
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

type FileInfo struct {
	Filename string
	Size     int64 // bytes
}

func (info FileInfo) HumanSize() string {
	if info.Size < 1024 {
		return fmt.Sprintf("%d B", info.Size)
	} else if info.Size < 1024*1024 {
		return fmt.Sprintf("%d K", info.Size/1024)
	} else {
		return fmt.Sprintf("%d M", info.Size/1024/1024)
	}
}

type VersionInfo struct {
	Name  string
	Files []FileInfo
}

type ByVersion []VersionInfo

func (a ByVersion) Len() int           { return len(a) }
func (a ByVersion) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByVersion) Less(i, j int) bool { return version.Compare(a[i].Name, a[j].Name, "<") }

// 获取版本信息
// downloadPaht: 下载路径
// categoryLength: 分类路径
func getVersions(downloadPath string) []VersionInfo {
	downloadPath, _ = filepath.Abs(downloadPath)
	categoryLength := len(strings.Split(downloadPath, "/")) + 1
	fileLength := categoryLength + 1
	versions := []VersionInfo{}
	var version VersionInfo

	first := true
	filepath.Walk(downloadPath, func(path string, info os.FileInfo, err error) error {
		if path == downloadPath {
			return nil
		}

		temp := strings.Split(path, "/")
		if len(temp) == categoryLength {
			// 版本文件夹
			if !first {
				versions = append(versions, version)
			} else {
				first = false
			}

			version = VersionInfo{
				Name:  info.Name(),
				Files: []FileInfo{},
			}
		} else if len(temp) == fileLength {
			// 文件
			version.Files = append(version.Files, FileInfo{
				Filename: info.Name(),
				Size:     info.Size(),
			})
		}

		return nil
	})

	versions = append(versions, version)
	// 倒序排列
	sort.Sort(sort.Reverse(ByVersion(versions)))

	return versions
}

func downloadGoHandler(handler *Handler) {
	handler.renderTemplate("download.html", BASE, map[string]interface{}{
		"versions": getVersions(Config.GoDownloadPath),
		"active":   "download",
	})
}

func downloadLiteIDEHandler(handler *Handler) {
	handler.renderTemplate("download/liteide.html", BASE, map[string]interface{}{
		"versions": getVersions(Config.GoDownloadPath),
		"active":   "download",
	})
}
