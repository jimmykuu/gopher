package apis

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pborman/uuid"

	"github.com/jimmykuu/gopher/conf"
)

// UploadImage 上传文件
type UploadImage struct {
	Base
}

// Post /upload/image
func (a *UploadImage) Post() interface{} {
	return uploadImage(a.Req(), []string{"upload", "image"}, -1)
}

type Sizer interface {
	Size() int64
}

// uploadImage 上传图片，保存图片到指定位置，并返回图片 URL 地址
// maxSize: byte 如果是 -1，不检查图片大小
func uploadImage(r *http.Request, folders []string, maxSize int64) interface{} {
	file, header, err := r.FormFile("image")
	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": fmt.Sprintf("图片上传失败 error(%s)", err.Error()),
		}
	}

	fileSize := file.(Sizer).Size()

	if maxSize > 0 {
		if fileSize > maxSize {
			return map[string]interface{}{
				"status":  0,
				"message": fmt.Sprintf("图片尺寸大于 %dK，请选择 %dK 以内的图片上传", maxSize/1024, maxSize/1024),
			}
		}
	}

	defer file.Close()

	// 检查是否是 jpg/png/gif 格式
	uploadFileType := header.Header["Content-Type"][0]

	filenameExtension := ""
	if uploadFileType == "image/jpeg" {
		filenameExtension = ".jpg"
	} else if uploadFileType == "image/png" {
		filenameExtension = ".png"
	} else if uploadFileType == "image/gif" {
		filenameExtension = ".gif"
	}

	if filenameExtension == "" {
		return map[string]interface{}{
			"status":  0,
			"message": "不支持的文件格式，请上传 jpg/png/gif 图片文件",
		}
	}

	imageFolder := filepath.Join(folders...)
	// 文件名：32位uuid+后缀组成
	filename := strings.Replace(uuid.NewUUID().String(), "-", "", -1) + filenameExtension
	toFile, err := os.Create(filepath.Join(conf.Config.ImagePath, imageFolder, filename))
	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": fmt.Sprintf("图片上传失败 error(%s)", err.Error()),
		}
	}

	io.Copy(toFile, file)

	return map[string]interface{}{
		"status":    1,
		"image_url": fmt.Sprintf("https://is.golangtc.com/upload/%s/%s", strings.Join(folders, "/"), filename),
	}
}
