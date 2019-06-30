// +build !bindata

package static

import (
	"net/http"

	"gitea.com/lunny/tango"
)

func PublicFileSystem(publicDir string) http.FileSystem {
	return http.Dir(publicDir)
}

// Static implements the tango static handler for serving assets.
func Static(publicDir string) tango.Handler {
	return tango.Static(tango.StaticOptions{
		RootPath: publicDir,
		Prefix:   "static",
	})
}
