package config

import (
	"io/fs"

	"github.com/c2pc/config-migrate/internal/migrator"
	"github.com/golang-migrate/migrate/v4/database"
)

type Configure interface {
	Unmarshal([]byte, interface{}) error
	Marshal(interface{}) ([]byte, error)
	Version([]byte) (int, bool, error)
	EmptyData() []byte
}

type Settings struct {
	Path string
	Perm fs.FileMode
}

func New(m Configure, cfg Settings) database.Driver {
	return migrator.New(m, migrator.Config(cfg))
}
