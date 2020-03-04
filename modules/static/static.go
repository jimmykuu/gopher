// +build bindata

package static

import (
	"net/http"

	"gitea.com/lunny/tango"
)

func PublicFileSystem(publicDir string) http.FileSystem {
	return Assets
}

// Static implements the tango static handler for serving assets.
func Static(publicDir string) tango.Handler {
	return tango.Static(tango.StaticOptions{
		Prefix:     "static",
		FileSystem: PublicFileSystem(publicDir),
	})
}
