// +build bindata

package static

import (
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/lunny/tango"
)

// Static implements the macaron static handler for serving assets.
func Static() tango.Handler {
	return tango.Static(tango.StaticOptions{
		Prefix: "static",
		FileSystem: &assetfs.AssetFS{
			Asset:     Asset,
			AssetDir:  AssetDir,
			AssetInfo: AssetInfo,
			Prefix:    "../../static",
		},
	})
}
