package actions

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mcuadros/go-version"
	"github.com/tango-contrib/renders"

	"github.com/jimmykuu/gopher/conf"
)

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

// getVersions 获取版本信息
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

// DownloadGo 下载 Go
type DownloadGo struct {
	RenderBase
}

// Get /download
func (a *DownloadGo) Get() error {
	return a.Render("download/go.html", renders.T{
		"title":    "下载 Go",
		"versions": getVersions(conf.Config.GoDownloadPath),
	})
}

// DownloadLiteIDE 下载 LiteIDE
type DownloadLiteIDE struct {
	RenderBase
}

// Get /download/liteide
func (a *DownloadLiteIDE) Get() error {
	return a.Render("download/liteide.html", renders.T{
		"title":    "下载 LiteIDE",
		"versions": getVersions(conf.Config.LiteIDEDownloadPath),
	})
}
