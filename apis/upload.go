package apis

import (
	"errors"
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
			"message": fmt.Sprintf("图片上传失败（%s）", err.Error()),
		}
	}
	defer file.Close()

	fileType := header.Header["Content-Type"][0]

	filename, err := saveImage(file, fileType, []string{"upload", "image"}, -1)
	if err != nil {
		return map[string]interface{}{
			"status":  0,
			"message": fmt.Sprintf("图片上传失败（%s）", err.Error()),
		}
	}

	return map[string]interface{}{
		"status":    1,
		"image_url": fmt.Sprintf("https://is.golangtc.com/upload/image/%s", filename),
	}
}

type Sizer interface {
	Size() int64
}

// uploadImage 上传图片，保存图片到指定位置，并返回图片 URL 地址
// maxSize: byte 如果是 -1，不检查图片大小
// 返回：文件名
func saveImage(source io.Reader, fileType string, folders []string, maxSize int64) (string, error) {
	if maxSize > 0 {
		fileSize := source.(Sizer).Size()

		if fileSize > maxSize {
			return "", errors.New(fmt.Sprintf("图片尺寸大于 %dK，请选择 %dK 以内的图片上传", maxSize/1024, maxSize/1024))
		}
	}

	// 检查是否是 jpg/png/gif 格式
	filenameExtension := ""
	if fileType == "image/jpeg" {
		filenameExtension = ".jpg"
	} else if fileType == "image/png" {
		filenameExtension = ".png"
	} else if fileType == "image/gif" {
		filenameExtension = ".gif"
	}

	if filenameExtension == "" {
		return "", errors.New("不支持的文件格式，请上传 jpg/png/gif 格式文件")
	}

	imageFolder := filepath.Join(folders...)
	// 文件名：32位uuid+后缀组成
	filename := strings.Replace(uuid.NewUUID().String(), "-", "", -1) + filenameExtension
	toFile, err := os.Create(filepath.Join(conf.Config.ImagePath, imageFolder, filename))
	if err != nil {
		return "", err
	}

	io.Copy(toFile, source)

	return filename, nil
}
