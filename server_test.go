package gopher

import (
	"bytes"
	"log"
)

func init() {
	logger = log.New(new(bytes.Buffer), "", log.LstdFlags)
}
