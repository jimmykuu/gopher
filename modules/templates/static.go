// +build bindata

package templates

import (
	"net/http"
)

// FileSystem implements the handler for serving the templates.
func FileSystem(templatesDir string) http.FileSystem {
	return Assets
}
