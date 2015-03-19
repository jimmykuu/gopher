package gopher

import (
	"bytes"
	"testing"
)

func TestBuildCode(t *testing.T) {
	var buf bytes.Buffer
	err := buildGoCode(`package main fun`, &buf)
	if err == nil {
		t.Fail()
	}
}
