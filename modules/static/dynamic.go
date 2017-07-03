// +build !bindata

package static

import (
	"github.com/lunny/tango"
)

// Static implements the macaron static handler for serving assets.
func Static() tango.Handler {
	return tango.Static(tango.StaticOptions{
		RootPath: "./static",
		Prefix:   "static",
	})
}
