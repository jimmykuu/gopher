// +build bindata

package etc

import (
	"net/http"

	"github.com/elazarl/go-bindata-assetfs"
)

// FileSystem implements the macaron handler for serving the templates.
func FileSystem(templatesDir string) http.FileSystem {
	return &assetfs.AssetFS{
		Asset:     Asset,
		AssetDir:  AssetDir,
		AssetInfo: AssetInfo,
		Prefix:    "../../etc",
	}
}
