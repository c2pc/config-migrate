package config_migrate

import (
	"github.com/c2pc/config-migrate/internal/replacer"
)

func RegisterReplacer(name string, fn func() string) {
	replacer.Register(name, fn)
}
