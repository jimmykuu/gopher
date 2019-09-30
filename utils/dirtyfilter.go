package utils

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	filter "github.com/antlinker/go-dirtyfilter"
	"github.com/antlinker/go-dirtyfilter/store"

	"github.com/jimmykuu/gopher/conf"
)

// NewDirtyManager 新建敏感词管理器
func NewDirtyManager(reader io.Reader) *filter.DirtyManager {
	buf := bufio.NewReader(reader)
	var words = []string{}
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)

		if line != "" {
			words = append(words, line)
		}

		if err != nil {
			if err == io.EOF {
				fmt.Println("File read ok!")
				break
			}
		}
	}

	memStore, err := store.NewMemoryStore(store.MemoryConfig{
		DataSource: words,
	})
	if err != nil {
		panic(err)
	}

	return filter.NewDirtyManager(memStore)
}

// HasSensitiveWords 是否敏感词
func HasSensitiveWords(text string) bool {
	result, err := conf.DirtyManager.Filter().Filter(text, '*', '@')
	if err != nil {
		fmt.Println(err)

		return true
	}

	fmt.Println(result)

	return len(result) > 0
}
