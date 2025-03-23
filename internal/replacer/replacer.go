package replacer

import (
	"strings"
	"sync"
)

var replacersMu sync.RWMutex
var replacers = make(map[string]func() string)

// Register globally registers a replacer.
func Register(name string, replacer func() string) {
	replacersMu.Lock()
	defer replacersMu.Unlock()
	if replacer == nil {
		panic("Register replacer is nil")
	}
	if _, dup := replacers[name]; dup {
		panic("Register called twice for replacer " + name)
	}
	replacers[name] = replacer
}

func Replace(value string) string {
	for name, replacer := range replacers {
		index := strings.Index(value, name)
		if index != -1 {
			value = strings.Replace(value, name, replacer(), -1)
		}
	}

	return value
}
