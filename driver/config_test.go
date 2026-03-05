package config_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	cfg "github.com/c2pc/config-migrate/driver"
	jsonDriver "github.com/c2pc/config-migrate/driver/json"
	"github.com/golang-migrate/migrate/v4/database"
)

// TestNew ensures New returns a database.Driver and accepts different settings.
func TestNew(t *testing.T) {
	var _ database.Driver = cfg.New(&jsonDriver.Json{}, cfg.Settings{})
	_ = cfg.New(&jsonDriver.Json{}, cfg.Settings{Path: "/tmp/cfg.json", Perm: 0600})
	_ = cfg.New(&jsonDriver.Json{}, cfg.Settings{OnlyOneVersion: true})
}

// TestOpen invalid URL returns error; valid URL returns same driver.
func TestOpen(t *testing.T) {
	c := cfg.New(&jsonDriver.Json{}, cfg.Settings{})
	_, err := c.Open("1http://foo.com")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.json")
	_, err = c.Open("json://" + path)
	if err != nil {
		t.Fatal(err)
	}
}

// TestLockUnlockClose runs Lock, Unlock, Close and ensures file is created.
func TestLockUnlockClose(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.json")
	c := cfg.New(&jsonDriver.Json{}, cfg.Settings{Path: path})
	d, err := c.Open("json://" + path)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Lock(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("config file not created after Lock:", err)
	}
	if err := d.Unlock(); err != nil {
		t.Fatal(err)
	}
	// Unlock already closes the file; no need to call Close
}

// TestRun_merge applies a migration and checks merge semantics (old wins for same key).
func TestRun_merge(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.json")
	initial := map[string]interface{}{"a": "old", "b": 2}
	writeJSON(t, path, initial)
	c := cfg.New(&jsonDriver.Json{}, cfg.Settings{Path: path})
	d, _ := c.Open("json://" + path)
	if err := d.Lock(); err != nil {
		t.Fatal(err)
	}
	defer d.Unlock()
	migration := `{"a": "new", "b": 99, "c": 3}`
	if err := d.Run(bytes.NewBufferString(migration)); err != nil {
		t.Fatal(err)
	}
	got := readJSON(t, path)
	if got["a"] != "old" {
		t.Errorf("merge: expected a=old, got %v", got["a"])
	}
	if got["b"].(float64) != 2 {
		t.Errorf("merge: expected b=2, got %v", got["b"])
	}
	if got["c"].(float64) != 3 {
		t.Errorf("merge: expected c=3, got %v", got["c"])
	}
}

// TestRun_deprecated applies migration with hosts_deprecated and checks value from old url.
func TestRun_deprecated(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.json")
	initial := map[string]interface{}{"url": "1.2.3.4"}
	writeJSON(t, path, initial)
	c := cfg.New(&jsonDriver.Json{}, cfg.Settings{Path: path})
	d, _ := c.Open("json://" + path)
	if err := d.Lock(); err != nil {
		t.Fatal(err)
	}
	defer d.Unlock()
	migration := `{"hosts_deprecated": "url", "hosts": ["default"]}`
	if err := d.Run(bytes.NewBufferString(migration)); err != nil {
		t.Fatal(err)
	}
	got := readJSON(t, path)
	hosts, ok := got["hosts"].([]interface{})
	if !ok || len(hosts) != 1 || hosts[0] != "1.2.3.4" {
		t.Errorf("deprecated: expected hosts=[1.2.3.4], got %v", got["hosts"])
	}
}

// TestRun_replace applies migration with port_replace and checks new value wins.
func TestRun_replace(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.json")
	initial := map[string]interface{}{"port": 3000}
	writeJSON(t, path, initial)
	c := cfg.New(&jsonDriver.Json{}, cfg.Settings{Path: path})
	d, _ := c.Open("json://" + path)
	if err := d.Lock(); err != nil {
		t.Fatal(err)
	}
	defer d.Unlock()
	migration := `{"port_deprecated_replace": "", "port": 8080}`
	if err := d.Run(bytes.NewBufferString(migration)); err != nil {
		t.Fatal(err)
	}
	got := readJSON(t, path)
	if got["port"].(float64) != 8080 {
		t.Errorf("replace: expected port=8080, got %v", got["port"])
	}
}

// TestRun_deprecatedAndReplace applies both; replace wins for its key.
func TestRun_deprecatedAndReplace(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.json")
	initial := map[string]interface{}{"url": "1.2.3.4", "port": 80}
	writeJSON(t, path, initial)
	c := cfg.New(&jsonDriver.Json{}, cfg.Settings{Path: path})
	d, _ := c.Open("json://" + path)
	if err := d.Lock(); err != nil {
		t.Fatal(err)
	}
	defer d.Unlock()
	migration := `{"hosts_deprecated": "url", "hosts": ["x"], "port_deprecated_replace": "", "port": 443}`
	if err := d.Run(bytes.NewBufferString(migration)); err != nil {
		t.Fatal(err)
	}
	got := readJSON(t, path)
	hosts, _ := got["hosts"].([]interface{})
	if len(hosts) != 1 || hosts[0] != "1.2.3.4" {
		t.Errorf("expected hosts=[1.2.3.4], got %v", got["hosts"])
	}
	if got["port"].(float64) != 443 {
		t.Errorf("expected port=443, got %v", got["port"])
	}
}

// TestVersion_emptyFile returns NilVersion for empty file.
func TestVersion_emptyFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.json")
	c := cfg.New(&jsonDriver.Json{}, cfg.Settings{Path: path})
	d, _ := c.Open("json://" + path)
	if err := d.Lock(); err != nil {
		t.Fatal(err)
	}
	defer d.Unlock()
	v, dirty, err := d.Version()
	if err != nil {
		t.Fatal(err)
	}
	if v != database.NilVersion {
		t.Errorf("expected NilVersion for empty file, got %d", v)
	}
	if dirty {
		t.Error("expected dirty=false")
	}
}

// TestVersion_and_SetVersion writes version/force and reads them back.
func TestVersion_and_SetVersion(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.json")
	writeJSON(t, path, map[string]interface{}{"version": 1, "force": false})
	c := cfg.New(&jsonDriver.Json{}, cfg.Settings{Path: path})
	d, _ := c.Open("json://" + path)
	if err := d.Lock(); err != nil {
		t.Fatal(err)
	}
	defer d.Unlock()
	v, dirty, err := d.Version()
	if err != nil {
		t.Fatal(err)
	}
	if v != 1 || dirty {
		t.Errorf("expected version=1 dirty=false, got %d %t", v, dirty)
	}
	if err := d.SetVersion(5, true); err != nil {
		t.Fatal(err)
	}
	v, dirty, err = d.Version()
	if err != nil {
		t.Fatal(err)
	}
	if v != 5 || !dirty {
		t.Errorf("expected version=5 dirty=true, got %d %t", v, dirty)
	}
}

// TestDrop truncates file and writes empty data.
func TestDrop(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.json")
	writeJSON(t, path, map[string]interface{}{"version": 1, "data": "x"})
	c := cfg.New(&jsonDriver.Json{}, cfg.Settings{Path: path})
	d, _ := c.Open("json://" + path)
	if err := d.Lock(); err != nil {
		t.Fatal(err)
	}
	defer d.Unlock()
	if err := d.Drop(); err != nil {
		t.Fatal(err)
	}
	got := readJSON(t, path)
	if len(got) != 0 {
		t.Errorf("expected empty map after Drop, got %v", got)
	}
}

// TestOnlyOneVersion Version returns 0,false and SetVersion is no-op.
func TestOnlyOneVersion(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.json")
	writeJSON(t, path, map[string]interface{}{"version": 3, "force": true})
	c := cfg.New(&jsonDriver.Json{}, cfg.Settings{Path: path, OnlyOneVersion: true})
	d, _ := c.Open("json://" + path)
	if err := d.Lock(); err != nil {
		t.Fatal(err)
	}
	defer d.Unlock()
	v, dirty, err := d.Version()
	if err != nil {
		t.Fatal(err)
	}
	if v != 0 || dirty {
		t.Errorf("OnlyOneVersion: expected 0, false, got %d %t", v, dirty)
	}
	if err := d.SetVersion(10, true); err != nil {
		t.Fatal(err)
	}
	got := readJSON(t, path)
	if got["version"].(float64) != 3 {
		t.Errorf("SetVersion with OnlyOneVersion should not change file, got version %v", got["version"])
	}
}

func writeJSON(t *testing.T, path string, m map[string]interface{}) {
	t.Helper()
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func readJSON(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		return map[string]interface{}{}
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

// TestList returns registered driver names (e.g. json, yaml are registered in init).
func TestList(t *testing.T) {
	names := cfg.List()
	if len(names) == 0 {
		t.Error("expected at least one registered driver")
	}
	found := false
	for _, n := range names {
		if n == "json" || n == "yaml" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected json or yaml in list, got %v", names)
	}
}
