package migrate

import (
	"github.com/c2pc/golang-file-migrate/internal/replacer"
)

func RegisterReplacer(name string, fn func() string) {
	replacer.Register(name, fn)
}
