package migrate

import (
	"github.com/c2pc/golang-file-migrate/internal/replacer"
)

// RegisterReplacer globally registers a replacer.
func RegisterReplacer(name string, fn func() string) {
	replacer.RegisterReplacer(name, fn)
}
