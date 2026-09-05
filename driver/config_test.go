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

// TestRun_deprecatedObjectVariants covers preserve / extend / replace / drop through Config.Run.
func TestRun_deprecatedObjectVariants(t *testing.T) {
	type step struct {
		migration string
	}
	tests := []struct {
		name    string
		initial map[string]interface{}
		steps   []step
		check   func(t *testing.T, got map[string]interface{})
	}{
		{
			name: "preserve_only",
			initial: map[string]interface{}{
				"vcs": map[string]interface{}{
					"webrtc": map[string]interface{}{
						"rtc": map[string]interface{}{
							"a":    1,
							"user": "x",
						},
					},
				},
			},
			steps: []step{
				{`{"vcs":{"webrtc":{"rtc_deprecated":"vcs.webrtc.rtc"}}}`},
				{`{"vcs":{"webrtc":{"rtc_deprecated":"vcs.webrtc.rtc"}}}`},
			},
			check: func(t *testing.T, got map[string]interface{}) {
				rtc := got["vcs"].(map[string]interface{})["webrtc"].(map[string]interface{})["rtc"].(map[string]interface{})
				if rtc["a"].(float64) != 1 || rtc["user"] != "x" {
					t.Fatalf("preserve failed: %v", rtc)
				}
			},
		},
		{
			name: "extend_add_key",
			initial: map[string]interface{}{
				"vcs": map[string]interface{}{
					"webrtc": map[string]interface{}{
						"rtc": map[string]interface{}{"user": "x", "n": map[string]interface{}{"k": 1}},
					},
				},
			},
			steps: []step{
				{`{"vcs":{"webrtc":{"rtc_deprecated":"vcs.webrtc.rtc","rtc":{"new":true,"n":{"m":2}}}}}`},
			},
			check: func(t *testing.T, got map[string]interface{}) {
				rtc := got["vcs"].(map[string]interface{})["webrtc"].(map[string]interface{})["rtc"].(map[string]interface{})
				if rtc["user"] != "x" || rtc["new"] != true {
					t.Fatalf("extend top: %v", rtc)
				}
				n := rtc["n"].(map[string]interface{})
				if n["k"].(float64) != 1 || n["m"].(float64) != 2 {
					t.Fatalf("extend nested: %v", n)
				}
			},
		},
		{
			name: "extend_does_not_overwrite",
			initial: map[string]interface{}{
				"vcs": map[string]interface{}{
					"webrtc": map[string]interface{}{
						"rtc": map[string]interface{}{"port": 51000},
					},
				},
			},
			steps: []step{
				{`{"vcs":{"webrtc":{"rtc_deprecated":"vcs.webrtc.rtc","rtc":{"port":1,"extra":0}}}}`},
			},
			check: func(t *testing.T, got map[string]interface{}) {
				rtc := got["vcs"].(map[string]interface{})["webrtc"].(map[string]interface{})["rtc"].(map[string]interface{})
				if rtc["port"].(float64) != 51000 {
					t.Fatalf("overwrote user port: %v", rtc["port"])
				}
				if rtc["extra"].(float64) != 0 {
					t.Fatalf("extra not added: %v", rtc)
				}
			},
		},
		{
			name: "replace_inside_extend",
			initial: map[string]interface{}{
				"vcs": map[string]interface{}{
					"webrtc": map[string]interface{}{
						"rtc": map[string]interface{}{"port": 51000, "user": "x"},
					},
				},
			},
			steps: []step{
				{`{"vcs":{"webrtc":{"rtc_deprecated":"vcs.webrtc.rtc","rtc":{"port_deprecated_replace":"","port":50000,"extra":1}}}}`},
			},
			check: func(t *testing.T, got map[string]interface{}) {
				rtc := got["vcs"].(map[string]interface{})["webrtc"].(map[string]interface{})["rtc"].(map[string]interface{})
				if rtc["port"].(float64) != 50000 {
					t.Fatalf("replace failed: %v", rtc["port"])
				}
				if rtc["user"] != "x" || rtc["extra"].(float64) != 1 {
					t.Fatalf("side effects: %v", rtc)
				}
			},
		},
		{
			name: "omit_without_deprecated_drops",
			initial: map[string]interface{}{
				"vcs": map[string]interface{}{
					"webrtc": map[string]interface{}{
						"rtc": map[string]interface{}{"user": "x"},
					},
				},
			},
			steps: []step{
				{`{"vcs":{"webrtc":{}}}`},
			},
			check: func(t *testing.T, got map[string]interface{}) {
				webrtc := got["vcs"].(map[string]interface{})["webrtc"].(map[string]interface{})
				if _, ok := webrtc["rtc"]; ok {
					t.Fatalf("rtc should be dropped without deprecated, got %v", webrtc)
				}
			},
		},
		{
			name: "path_missing_with_template_uses_defaults",
			initial: map[string]interface{}{
				"vcs": map[string]interface{}{"tcp": map[string]interface{}{"port": 1}},
			},
			steps: []step{
				{`{"vcs":{"tcp":{"port":1},"webrtc":{"rtc_deprecated":"vcs.webrtc.rtc","rtc":{"port_range_start":50000}}}}`},
			},
			check: func(t *testing.T, got map[string]interface{}) {
				rtc := got["vcs"].(map[string]interface{})["webrtc"].(map[string]interface{})["rtc"].(map[string]interface{})
				if rtc["port_range_start"].(float64) != 50000 {
					t.Fatalf("expected template default, got %v", rtc)
				}
			},
		},
		{
			name: "full_template_still_keeps_unknown_with_deprecated",
			initial: map[string]interface{}{
				"vcs": map[string]interface{}{
					"webrtc": map[string]interface{}{
						"rtc": map[string]interface{}{
							"port_range_start": 51000,
							"user_only":        "keep",
						},
					},
				},
			},
			steps: []step{
				{`{"vcs":{"webrtc":{"rtc_deprecated":"vcs.webrtc.rtc","rtc":{"port_range_start":50000,"port_range_end":60000}}}}`},
			},
			check: func(t *testing.T, got map[string]interface{}) {
				rtc := got["vcs"].(map[string]interface{})["webrtc"].(map[string]interface{})["rtc"].(map[string]interface{})
				if rtc["user_only"] != "keep" {
					t.Fatalf("user_only lost: %v", rtc)
				}
				if rtc["port_range_start"].(float64) != 51000 {
					t.Fatalf("user port overwritten: %v", rtc["port_range_start"])
				}
				if rtc["port_range_end"].(float64) != 60000 {
					t.Fatalf("new key not added: %v", rtc)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			path := filepath.Join(tmp, "config.json")
			writeJSON(t, path, tt.initial)
			c := cfg.New(&jsonDriver.Json{}, cfg.Settings{Path: path})
			d, _ := c.Open("json://" + path)
			if err := d.Lock(); err != nil {
				t.Fatal(err)
			}
			defer d.Unlock()
			for i, s := range tt.steps {
				if err := d.Run(bytes.NewBufferString(s.migration)); err != nil {
					t.Fatalf("step %d: %v", i, err)
				}
			}
			tt.check(t, readJSON(t, path))
		})
	}
}

// TestRun_deprecatedPreserveSubtree keeps webrtc.rtc as-is (including user-only keys)
// when a later migration lists only rtc_deprecated and omits the rtc template.
func TestRun_deprecatedPreserveSubtree(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.json")
	initial := map[string]interface{}{
		"vcs": map[string]interface{}{
			"tcp": map[string]interface{}{"port": 8059},
			"webrtc": map[string]interface{}{
				"rtc": map[string]interface{}{
					"port_range_start": 51000,
					"port_range_end":   60000,
					"user_ice_servers": []interface{}{"stun:custom.example"},
					"vendor":           map[string]interface{}{"flag": "on"},
				},
			},
		},
	}
	writeJSON(t, path, initial)
	c := cfg.New(&jsonDriver.Json{}, cfg.Settings{Path: path})
	d, _ := c.Open("json://" + path)
	if err := d.Lock(); err != nil {
		t.Fatal(err)
	}
	defer d.Unlock()

	// Several later migrations: force-update tcp, add/update recording, never redefine full rtc.
	migrations := []string{
		`{"vcs":{"tcp":{"port_deprecated_replace":"","port":9001},"webrtc":{"rtc_deprecated":"vcs.webrtc.rtc"},"recording":{"filepath_prefix":"records"}}}`,
		`{"vcs":{"tcp":{"port_deprecated_replace":"","port":9002},"webrtc":{"rtc_deprecated":"vcs.webrtc.rtc"},"recording":{"filepath_prefix_deprecated_replace":"","filepath_prefix":"records_v2"}}}`,
		// Add a new key into rtc; user-only keys must remain.
		`{"vcs":{"tcp":{"port_deprecated_replace":"","port":9003},"webrtc":{"rtc_deprecated":"vcs.webrtc.rtc","rtc":{"new_setting":false}},"recording":{"filepath_prefix_deprecated_replace":"","filepath_prefix":"records_v3"}}}`,
	}
	for _, m := range migrations {
		if err := d.Run(bytes.NewBufferString(m)); err != nil {
			t.Fatal(err)
		}
	}

	got := readJSON(t, path)
	vcs := got["vcs"].(map[string]interface{})
	if vcs["tcp"].(map[string]interface{})["port"].(float64) != 9003 {
		t.Errorf("tcp.port: want 9003, got %v", vcs["tcp"])
	}
	if vcs["recording"].(map[string]interface{})["filepath_prefix"] != "records_v3" {
		t.Errorf("recording should update via replace, got %v", vcs["recording"])
	}
	rtc := vcs["webrtc"].(map[string]interface{})["rtc"].(map[string]interface{})
	if rtc["port_range_start"].(float64) != 51000 {
		t.Errorf("rtc.port_range_start should stay user value, got %v", rtc["port_range_start"])
	}
	servers, ok := rtc["user_ice_servers"].([]interface{})
	if !ok || len(servers) != 1 || servers[0] != "stun:custom.example" {
		t.Errorf("user_ice_servers lost: %v", rtc["user_ice_servers"])
	}
	vendor, ok := rtc["vendor"].(map[string]interface{})
	if !ok || vendor["flag"] != "on" {
		t.Errorf("user vendor map lost: %v", rtc["vendor"])
	}
	if rtc["new_setting"] != false {
		t.Errorf("new_setting should be added, got %v", rtc["new_setting"])
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
