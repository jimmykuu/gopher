package apis

import (
	"encoding/json"

	"github.com/lunny/tango"
)

type Base struct {
	tango.Json
	tango.Ctx
}

func (b *Base) ReadJSON(v interface{}) error {
	decoder := json.NewDecoder(b.Req().Body)
	return decoder.Decode(&v)
}
