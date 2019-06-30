// +build !bindata

package templates

import (
	"net/http"
	"path"
)

// FileSystem implements the macaron handler for serving the templates.
func FileSystem(templatesDir string) http.FileSystem {
	return http.Dir(path.Join("./", templatesDir))
}
