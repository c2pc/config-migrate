package replacer

import (
	"strings"
	"sync"
)

var replacersMu sync.RWMutex
var replacers = make(map[string]func() string)

func init() {
	replacersMu.Lock()
	defer replacersMu.Unlock()
	for name, replacer := range defaultReplacers {
		replacers[name] = replacer
	}
}

// RegisterReplacer globally registers a replacer.
func RegisterReplacer(name string, replacer func() string) {
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
	newValue := value
	for name, replacer := range replacers {
		index := strings.Index(value, name)
		if index != -1 {
			value = strings.Replace(value, name, replacer(), -1)
		}
	}

	return newValue
}
