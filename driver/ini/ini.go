package ini

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/c2pc/config-migrate/driver"
	"github.com/golang-migrate/migrate/v4/database"
	"gopkg.in/ini.v1"
)

// Ini implements driver.Driver for INI config files using gopkg.in/ini.v1.
// Structure: sections [name] become top-level keys; key=value in section become nested map.
// Keys before any [section] go to the default section ""; "version" and "force" are also exposed at top level for the migrator.
type Ini struct{}

func init() {
	config.Register("ini", &Ini{}, config.Settings{})
}

// New returns a database.Driver that uses the INI driver with the given settings.
func New(cfg config.Settings) database.Driver {
	return config.New(&Ini{}, cfg)
}

// Unmarshal parses INI bytes into a map using gopkg.in/ini.v1.
// Result: root["section"] = map[string]interface{} for each [section],
// root[""] = default section (DEFAULT in ini.v1); root["version"] and root["force"] at top level.
func (Ini) Unmarshal(data []byte, out interface{}) error {
	ptr, ok := out.(*map[string]interface{})
	if !ok {
		return fmt.Errorf("ini: out must be *map[string]interface{}")
	}
	if data == nil {
		data = []byte{}
	}
	if len(data) == 0 {
		*ptr = map[string]interface{}{}
		return nil
	}

	f, err := ini.LoadSources(ini.LoadOptions{}, data)
	if err != nil {
		return err
	}

	*ptr = fileToMap(f)
	return nil
}

// fileToMap converts ini.File to map[string]interface{} (section name -> map of key->value).
// Default section (DEFAULT) is also exposed as "" and its keys at top level.
func fileToMap(f *ini.File) map[string]interface{} {
	out := make(map[string]interface{})
	for _, sec := range f.Sections() {
		name := sec.Name()
		if name == ini.DefaultSection {
			name = ""
		}
		sm := make(map[string]interface{})
		for _, k := range sec.Keys() {
			sm[k.Name()] = k.String()
		}
		if name == "" {
			for k, v := range sm {
				out[k] = v
			}
		}
		out[name] = sm
	}
	return out
}

// Marshal serializes the map back to INI using gopkg.in/ini.v1.
func (Ini) Marshal(i interface{}, _ bool) ([]byte, error) {
	m, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("ini: expected map[string]interface{}")
	}

	f := ini.Empty()
	// Default section: version, force, and any other non-map top-level
	defaultSec, err := f.NewSection(ini.DefaultSection)
	if err != nil {
		return nil, err
	}
	for k, v := range m {
		if k == "" {
			continue
		}
		_, isMap := v.(map[string]interface{})
		if isMap {
			continue
		}
		if _, err := defaultSec.NewKey(k, fmt.Sprint(v)); err != nil {
			return nil, err
		}
	}
	// Named sections (excluding "" which we already wrote as default)
	for _, name := range sectionOrder(m) {
		if name == "" {
			continue
		}
		v := m[name]
		sm, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		sec, err := f.NewSection(name)
		if err != nil {
			return nil, err
		}
		for k, val := range sm {
			if _, err := sec.NewKey(k, fmt.Sprint(val)); err != nil {
				return nil, err
			}
		}
	}

	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func sectionOrder(m map[string]interface{}) []string {
	order := make([]string, 0, len(m))
	for k := range m {
		if _, isMap := m[k].(map[string]interface{}); isMap {
			order = append(order, k)
		}
	}
	return order
}

// Version reads version and force from INI data (all values are stored as strings).
func (m Ini) Version(data []byte) (int, bool, error) {
	var out map[string]interface{}
	if err := m.Unmarshal(data, &out); err != nil {
		return 0, false, err
	}
	version := 0
	if v, ok := out["version"]; ok {
		if s, ok := v.(string); ok {
			version, _ = strconv.Atoi(s)
		}
	}
	force := false
	if v, ok := out["force"]; ok {
		if s, ok := v.(string); ok {
			force = strings.EqualFold(s, "true") || strings.EqualFold(s, "yes") || s == "1"
		}
	}
	return version, force, nil
}

// EmptyData returns minimal INI with version=0 and force=false.
func (Ini) EmptyData() []byte {
	return []byte("version=0\nforce=false\n")
}
