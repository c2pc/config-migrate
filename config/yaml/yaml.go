package yaml

import (
	"github.com/c2pc/config-migrate/config"
	"github.com/golang-migrate/migrate/v4/database"
	"gopkg.in/yaml.v3"
)

type Yaml struct {
}

func init() {
	config.Register("yaml", &Yaml{}, config.Settings{})
}

func New(cfg config.Settings) database.Driver {
	return config.New(&Yaml{}, cfg)
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
	if err := m.Unmarshal(bytes, v); err != nil {
		return 0, false, err
	}

	return v.Version, v.Force, nil
}

func (m Yaml) EmptyData() []byte {
	return []byte{}
}
