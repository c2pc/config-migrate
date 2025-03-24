package yaml

import (
	"github.com/c2pc/config-migrate/config"
	"github.com/c2pc/config-migrate/internal/migrator"
	"github.com/golang-migrate/migrate/v4/database"
	"gopkg.in/yaml.v3"
)

type Yaml struct {
}

func init() {
	j := migrator.New(&Yaml{}, migrator.Config{})
	database.Register("yaml", j)
}

func New(cfg config.Settings) database.Driver {
	return migrator.New(&Yaml{}, migrator.Config(cfg))
}

func (m Yaml) Unmarshal(bytes []byte, i interface{}) error {
	return yaml.Unmarshal(bytes, i)
}

func (m Yaml) Marshal(i interface{}) ([]byte, error) {
	return yaml.Marshal(i)
}

type version struct {
	Version int  `yaml:"version"`
	Force   bool `yaml:"force"`
}

func (m Yaml) Version(bytes []byte) (int, bool, error) {
	v := new(version)
	if err := yaml.Unmarshal(bytes, v); err != nil {
		return 0, false, err
	}

	return v.Version, v.Force, nil
}

func (m Yaml) EmptyData() []byte {
	return []byte{}
}
