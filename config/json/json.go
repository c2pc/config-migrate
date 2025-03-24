package json

import (
	"encoding/json"

	"github.com/c2pc/config-migrate/config"
	"github.com/golang-migrate/migrate/v4/database"
)

type Json struct {
}

func init() {
	config.Register("json", &Json{}, config.Settings{})
}

func New(cfg config.Settings) database.Driver {
	return config.New(&Json{}, cfg)
}

func (m Json) Unmarshal(bytes []byte, i interface{}) error {
	if bytes == nil || len(bytes) == 0 {
		bytes = []byte(`{}`)
	}
	return json.Unmarshal(bytes, i)
}

func (m Json) Marshal(i interface{}) ([]byte, error) {
	return json.Marshal(i)
}

type version struct {
	Version int  `json:"version"`
	Force   bool `json:"force"`
}

func (m Json) Version(bytes []byte) (int, bool, error) {
	v := new(version)
	if err := m.Unmarshal(bytes, v); err != nil {
		return 0, false, err
	}

	return v.Version, v.Force, nil
}

func (m Json) EmptyData() []byte {
	return []byte("{}")
}
