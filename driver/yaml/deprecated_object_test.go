package yaml_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	cfg "github.com/c2pc/config-migrate/driver"
	yamlDriver "github.com/c2pc/config-migrate/driver/yaml"
	"gopkg.in/yaml.v3"
)

func writeYAML(t *testing.T, path string, m map[string]interface{}) {
	t.Helper()
	data, err := yaml.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func readYAML(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := yaml.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

// TestRun_yamlDeprecatedObjectPreserveExtend mirrors production YAML migrations for webrtc.rtc.
func TestRun_yamlDeprecatedObjectPreserveExtend(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.yaml")
	writeYAML(t, path, map[string]interface{}{
		"vcs": map[string]interface{}{
			"tcp": map[string]interface{}{"port": 8059},
			"webrtc": map[string]interface{}{
				"rtc": map[string]interface{}{
					"port_range_start": 51000,
					"port_range_end":   60000,
					"user_ice_servers": []interface{}{"stun:custom.example"},
					"vendor": map[string]interface{}{
						"flag":  "on",
						"extra": 42,
					},
				},
			},
		},
	})

	c := yamlDriver.New(cfg.Settings{Path: path})
	d, err := c.Open("yaml://" + path)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Lock(); err != nil {
		t.Fatal(err)
	}
	defer d.Unlock()

	migrations := []string{
		// preserve only
		`
vcs:
  tcp:
    port_deprecated_replace: ""
    port: 9001
  webrtc:
    rtc_deprecated: vcs.webrtc.rtc
  recording:
    filepath_prefix: records
`,
		// preserve again
		`
vcs:
  tcp:
    port_deprecated_replace: ""
    port: 9002
  webrtc:
    rtc_deprecated: vcs.webrtc.rtc
  recording:
    filepath_prefix_deprecated_replace: ""
    filepath_prefix: records_v2
`,
		// extend: add key + nested key, replace one field
		`
vcs:
  tcp:
    port_deprecated_replace: ""
    port: 9003
  webrtc:
    rtc_deprecated: vcs.webrtc.rtc
    rtc:
      new_setting: false
      port_range_end_deprecated_replace: ""
      port_range_end: 70000
      vendor:
        added: true
  recording:
    filepath_prefix_deprecated_replace: ""
    filepath_prefix: records_v3
`,
	}

	for i, m := range migrations {
		if err := d.Run(bytes.NewBufferString(m)); err != nil {
			t.Fatalf("migration %d: %v", i, err)
		}
	}

	got := readYAML(t, path)
	vcs := got["vcs"].(map[string]interface{})
	if vcs["tcp"].(map[string]interface{})["port"] != 9003 {
		t.Fatalf("tcp.port: %v", vcs["tcp"])
	}
	if vcs["recording"].(map[string]interface{})["filepath_prefix"] != "records_v3" {
		t.Fatalf("recording: %v", vcs["recording"])
	}
	rtc := vcs["webrtc"].(map[string]interface{})["rtc"].(map[string]interface{})
	if rtc["port_range_start"] != 51000 {
		t.Fatalf("user port_range_start lost: %v", rtc["port_range_start"])
	}
	if rtc["port_range_end"] != 70000 {
		t.Fatalf("replace port_range_end failed: %v", rtc["port_range_end"])
	}
	if rtc["new_setting"] != false {
		t.Fatalf("new_setting: %v", rtc["new_setting"])
	}
	servers, ok := rtc["user_ice_servers"].([]interface{})
	if !ok || len(servers) != 1 || servers[0] != "stun:custom.example" {
		t.Fatalf("user_ice_servers: %v", rtc["user_ice_servers"])
	}
	vendor := rtc["vendor"].(map[string]interface{})
	if vendor["flag"] != "on" || vendor["extra"] != 42 || vendor["added"] != true {
		t.Fatalf("vendor: %v", vendor)
	}
}
