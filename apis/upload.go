package apis

import (
	"fmt"
	"io"
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
	file, header, err := a.Req().FormFile("image")
	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": fmt.Sprintf("图片上传失败 error(%s)", err.Error()),
		}
	}

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

	// 文件名：32位uuid+后缀组成
	filename := strings.Replace(uuid.NewUUID().String(), "-", "", -1) + filenameExtension
	toFile, err := os.Create(filepath.Join(conf.Config.ImagePath, "upload", "image", filename))
	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": fmt.Sprintf("图片上传失败 error(%s)", err.Error()),
		}
	}

	io.Copy(toFile, file)

	return map[string]interface{}{
		"status":    1,
		"image_url": "https://is.golangtc.com/upload/image/" + filename,
	}
}
