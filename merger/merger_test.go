package merger

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/c2pc/config-migrate/replacer"
)

// TestMergeMaps covers the core merge behaviour: empty inputs, nil handling,
// and that when a key exists in both configs we prefer the old value (or merge
// maps/arrays as defined).
func TestMergeMaps(t *testing.T) {
	tests := []struct {
		name     string
		oldMap   map[string]interface{}
		newMap   map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:     "Nil maps",
			oldMap:   nil,
			newMap:   nil,
			expected: map[string]interface{}{},
		},
		{
			name:     "Empty maps",
			oldMap:   map[string]interface{}{},
			newMap:   map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name:     "New is nil, old is empty",
			oldMap:   map[string]interface{}{},
			newMap:   nil,
			expected: map[string]interface{}{},
		},
		{
			name:     "New is empty, old is nil",
			oldMap:   nil,
			newMap:   map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name:   "Old is nil",
			oldMap: nil,
			newMap: map[string]interface{}{
				"foo": "bar",
			},
			expected: map[string]interface{}{
				"foo": "bar",
			},
		},
		{
			name:   "Old is empty",
			oldMap: map[string]interface{}{},
			newMap: map[string]interface{}{
				"foo": "bar",
			},
			expected: map[string]interface{}{
				"foo": "bar",
			},
		},
		{
			name: "New is empty",
			oldMap: map[string]interface{}{
				"foo": "bar",
			},
			newMap:   map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name: "New is nil",
			oldMap: map[string]interface{}{
				"foo": "bar",
			},
			newMap:   nil,
			expected: map[string]interface{}{},
		},
		{
			name: "New string",
			oldMap: map[string]interface{}{
				"foo": "bar",
			},
			newMap: map[string]interface{}{
				"foo2": "bar",
			},
			expected: map[string]interface{}{
				"foo2": "bar",
			},
		},
		{
			name: "New string and old string",
			oldMap: map[string]interface{}{
				"foo": "bar",
			},
			newMap: map[string]interface{}{
				"foo":  "bar",
				"foo2": "bar",
			},
			expected: map[string]interface{}{
				"foo":  "bar",
				"foo2": "bar",
			},
		},
		{
			name: "New number",
			oldMap: map[string]interface{}{
				"foo": 1,
			},
			newMap: map[string]interface{}{
				"foo2": 2,
			},
			expected: map[string]interface{}{
				"foo2": 2,
			},
		},
		{
			name: "New number and old number",
			oldMap: map[string]interface{}{
				"foo": 1,
			},
			newMap: map[string]interface{}{
				"foo":  1,
				"foo2": 2,
			},
			expected: map[string]interface{}{
				"foo":  1,
				"foo2": 2,
			},
		},
		{
			name: "New bool",
			oldMap: map[string]interface{}{
				"foo": true,
			},
			newMap: map[string]interface{}{
				"foo2": false,
			},
			expected: map[string]interface{}{
				"foo2": false,
			},
		},
		{
			name: "New bool and old bool",
			oldMap: map[string]interface{}{
				"foo": true,
			},
			newMap: map[string]interface{}{
				"foo":  true,
				"foo2": false,
			},
			expected: map[string]interface{}{
				"foo":  true,
				"foo2": false,
			},
		},
		{
			name: "Old value is empty(bool)",
			oldMap: map[string]interface{}{
				"foo": nil,
			},
			newMap: map[string]interface{}{
				"foo": true,
			},
			expected: map[string]interface{}{
				"foo": nil,
			},
		},
		{
			name: "Old value is empty(string)",
			oldMap: map[string]interface{}{
				"foo": nil,
			},
			newMap: map[string]interface{}{
				"foo": "bar",
			},
			expected: map[string]interface{}{
				"foo": nil,
			},
		},
		{
			name: "Old value is empty(number)",
			oldMap: map[string]interface{}{
				"foo": nil,
			},
			newMap: map[string]interface{}{
				"foo": 1,
			},
			expected: map[string]interface{}{
				"foo": nil,
			},
		},
		{
			name: "New map",
			oldMap: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "bar",
				},
			},
			newMap: map[string]interface{}{
				"foo2": map[string]interface{}{
					"bar": "bar",
				},
			},
			expected: map[string]interface{}{
				"foo2": map[string]interface{}{
					"bar": "bar",
				},
			},
		},
		{
			name: "New map and old map",
			oldMap: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "bar",
				},
			},
			newMap: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "bar",
				},
				"foo2": map[string]interface{}{
					"bar": "bar",
				},
			},
			expected: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "bar",
				},
				"foo2": map[string]interface{}{
					"bar": "bar",
				},
			},
		},
		{
			name: "Old value is empty(map)",
			oldMap: map[string]interface{}{
				"foo": nil,
			},
			newMap: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "bar",
				},
			},
			expected: map[string]interface{}{
				"foo": nil,
			},
		},
		{
			name: "New array",
			oldMap: map[string]interface{}{
				"foo": []interface{}{
					"bar",
				},
			},
			newMap: map[string]interface{}{
				"foo2": []interface{}{
					"bar",
				},
			},
			expected: map[string]interface{}{
				"foo2": []interface{}{
					"bar",
				},
			},
		},
		{
			name: "New array and old array",
			oldMap: map[string]interface{}{
				"foo": []interface{}{
					"bar",
				},
			},
			newMap: map[string]interface{}{
				"foo": []interface{}{
					"bar",
				},
				"foo2": []interface{}{
					"bar",
				},
			},
			expected: map[string]interface{}{
				"foo": []interface{}{
					"bar",
				},
				"foo2": []interface{}{
					"bar",
				},
			},
		},
		{
			name: "Old value is empty(array)",
			oldMap: map[string]interface{}{
				"foo": nil,
			},
			newMap: map[string]interface{}{
				"foo": []interface{}{
					"bar",
				},
			},
			expected: map[string]interface{}{
				"foo": nil,
			},
		},
		{
			name: "New array of numbers",
			oldMap: map[string]interface{}{
				"foo": []int{
					1, 2, 3,
				},
			},
			newMap: map[string]interface{}{
				"foo": []int{
					1, 2, 3,
				},
				"foo2": []int{
					1, 2, 3, 4,
				},
			},
			expected: map[string]interface{}{
				"foo": []int{
					1, 2, 3,
				},
				"foo2": []int{
					1, 2, 3, 4,
				},
			},
		},
		{
			name: "New array of maps",
			oldMap: map[string]interface{}{
				"foo": []interface{}{
					map[string]interface{}{"foo2": "bar"},
				},
			},
			newMap: map[string]interface{}{
				"foo": []interface{}{
					map[string]interface{}{"foo2": "bar"},
				},
				"foo2": []interface{}{
					map[string]interface{}{"foo3": "bar"},
				},
			},
			expected: map[string]interface{}{
				"foo": []map[string]interface{}{
					{"foo2": "bar"},
				},
				"foo2": []map[string]interface{}{
					{"foo3": "bar"},
				},
			},
		},
		{
			name: "All types",
			oldMap: map[string]interface{}{
				"foo": []interface{}{
					map[string]interface{}{
						"foo_1": "bar",
						"foo_2": 1,
						"foo_3": true,
					},
				},
				"foo2": map[string]interface{}{
					"foo2_1": "bar",
					"foo2_2": 1,
					"foo2_3": true,
				},
				"foo3": 1,
				"foo4": "bar",
				"foo5": true,
				"foo6": []interface{}{
					1, 2, 3,
				},
				"foo7": []interface{}{
					"bar", "bar2", "bar3",
				},
				"foo8": []interface{}{
					true, false, true,
				},
				"foo14": []interface{}{
					1, 2, 3,
				},
			},
			newMap: map[string]interface{}{
				"foo": []interface{}{
					map[string]interface{}{
						"foo_1": "bar1",
						"foo_2": 2,
						"foo_3": false,
						"foo_4": "bar",
						"foo_5": 1,
						"foo_6": true,
					},
				},
				"foo2": map[string]interface{}{
					"foo2_1": "bar",
					"foo2_2": 2,
					"foo2_3": true,
					"foo2_4": "bar",
					"foo2_5": 1,
					"foo2_6": true,
				},
				"foo3": 1,
				"foo4": "bar2",
				"foo5": false,
				"foo6": []interface{}{
					map[string]interface{}{
						"foo6_1": "bar",
						"foo6_2": 1,
						"foo6_3": false,
					},
				},
				"foo7": map[string]interface{}{
					"foo7_1": "bar",
					"foo7_2": 1,
					"foo7_3": true,
				},
				"foo8":  1,
				"foo9":  "bar",
				"foo10": true,
				"foo11": []interface{}{
					1, 2, 3, 4, 5,
				},
				"foo12": []interface{}{
					"bar", "bar2", "bar3",
				},
				"foo13": []interface{}{
					true, false, true,
				},
				"foo14": []interface{}{
					1, 2, 3, 4, 5,
				},
			},
			expected: map[string]interface{}{
				"foo": []interface{}{
					map[string]interface{}{
						"foo_1": "bar",
						"foo_2": 1,
						"foo_3": true,
						"foo_4": "bar",
						"foo_5": 1,
						"foo_6": true,
					},
				},
				"foo2": map[string]interface{}{
					"foo2_1": "bar",
					"foo2_2": 1,
					"foo2_3": true,
					"foo2_4": "bar",
					"foo2_5": 1,
					"foo2_6": true,
				},
				"foo3": 1,
				"foo4": "bar",
				"foo5": true,
				"foo6": []interface{}{
					map[string]interface{}{
						"foo6_1": "bar",
						"foo6_2": 1,
						"foo6_3": false,
					},
				},
				"foo7": map[string]interface{}{
					"foo7_1": "bar",
					"foo7_2": 1,
					"foo7_3": true,
				},
				"foo8":  1,
				"foo9":  "bar",
				"foo10": true,
				"foo11": []interface{}{
					1, 2, 3, 4, 5,
				},
				"foo12": []interface{}{
					"bar", "bar2", "bar3",
				},
				"foo13": []interface{}{
					true, false, true,
				},
				"foo14": []interface{}{
					1, 2, 3,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Merge(tt.newMap, tt.oldMap)
			// Compare as JSON so slice/map order and type differences don't matter.
			res, err := json.Marshal(result)
			if err != nil {
				t.Error(err)
			}
			exp, err := json.Marshal(tt.expected)
			if err != nil {
				t.Error(err)
			}
			if string(res) != string(exp) {
				t.Errorf("Expected\n %s, got\n %s", string(exp), string(res))
			}
		})
	}
}

// TestDeprecatedExpand checks that key_deprecated_expand: "path->field" expands
// an array of scalars from old into an array of objects, using the new config as template.
func TestDeprecatedExpand(t *testing.T) {
	tests := []struct {
		name     string
		oldMap   map[string]interface{}
		newMap   map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "fps.urls array -> addresses with url field and certificate from template",
			oldMap: map[string]interface{}{
				"fps": map[string]interface{}{
					"urls": []interface{}{"url1", "url2", "url3"},
				},
			},
			newMap: map[string]interface{}{
				"fps": map[string]interface{}{
					"addresses_deprecated_expand": "fps.urls->url",
					"addresses": []interface{}{
						map[string]interface{}{"url": "url1", "certificate": "crt2.cert"},
						map[string]interface{}{"url": "url2", "certificate": "crt2.cert"},
						map[string]interface{}{"url": "url3", "certificate": "crt3.cert"},
					},
				},
			},
			expected: map[string]interface{}{
				"fps": map[string]interface{}{
					"addresses": []interface{}{
						map[string]interface{}{"url": "url1", "certificate": "crt2.cert"},
						map[string]interface{}{"url": "url2", "certificate": "crt2.cert"},
						map[string]interface{}{"url": "url3", "certificate": "crt3.cert"},
					},
				},
			},
		},
		{
			name: "expand with more source elements than template - reuse last template",
			oldMap: map[string]interface{}{
				"list": []interface{}{"a", "b", "c", "d"},
			},
			newMap: map[string]interface{}{
				"items_deprecated_expand": "list->value",
				"items": []interface{}{
					map[string]interface{}{"value": "x", "default": true},
					map[string]interface{}{"value": "y", "default": false},
				},
			},
			expected: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"value": "a", "default": true},
					map[string]interface{}{"value": "b", "default": false},
					map[string]interface{}{"value": "c", "default": false},
					map[string]interface{}{"value": "d", "default": false},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Merge(tt.newMap, tt.oldMap)
			res, _ := json.Marshal(result)
			exp, _ := json.Marshal(tt.expected)
			if string(res) != string(exp) {
				t.Errorf("Expected\n %s\n got\n %s", string(exp), string(res))
			}
		})
	}
}

// TestDeprecatedCollapse checks that key_deprecated_collapse: "arrayPath.field->targetPath"
// extracts field from each object in the array at arrayPath and writes array of scalars at targetPath.
func TestDeprecatedCollapse(t *testing.T) {
	tests := []struct {
		name     string
		oldMap   map[string]interface{}
		newMap   map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "fps.urls array of objects -> fps.urls array of url strings",
			oldMap: map[string]interface{}{
				"fps": map[string]interface{}{
					"urls": []interface{}{
						map[string]interface{}{"url": "url1", "certificate": "crt2.cert"},
						map[string]interface{}{"url": "url2", "certificate": "crt2.cert"},
						map[string]interface{}{"url": "url3", "certificate": "crt3.cert"},
					},
				},
			},
			newMap: map[string]interface{}{
				"fps": map[string]interface{}{
					"urls_deprecated_collapse": "fps.urls.url->fps.urls",
					"urls":                     []interface{}{},
				},
			},
			expected: map[string]interface{}{
				"fps": map[string]interface{}{
					"urls": []interface{}{"url1", "url2", "url3"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Merge(tt.newMap, tt.oldMap)
			res, _ := json.Marshal(result)
			exp, _ := json.Marshal(tt.expected)
			if string(res) != string(exp) {
				t.Errorf("Expected\n %s\n got\n %s", string(exp), string(res))
			}
		})
	}
}

// TestMergeDeprecated checks that keys like hosts_deprecated: "url" pull the
// value from the old config and merge it into the target (including scalar→array,
// array→scalar, and nested object migration).
func TestMergeDeprecated(t *testing.T) {
	tests := []struct {
		name     string
		oldMap   map[string]interface{}
		newMap   map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "hosts_deprecated: url, old has url only",
			oldMap: map[string]interface{}{
				"url": "1.1.1.1",
			},
			newMap: map[string]interface{}{
				"hosts_deprecated": "url",
				"hosts":            []interface{}{"2.2.2.2"},
			},
			expected: map[string]interface{}{
				"hosts": []interface{}{"1.1.1.1"},
			},
		},
		{
			name: "hosts_deprecated: url, old has url and hosts",
			oldMap: map[string]interface{}{
				"url":   "1.1.1.1",
				"hosts": []interface{}{"3.3.3.3"},
			},
			newMap: map[string]interface{}{
				"hosts_deprecated": "url",
				"hosts":            []interface{}{"2.2.2.2"},
			},
			expected: map[string]interface{}{
				"hosts": []interface{}{"3.3.3.3", "1.1.1.1"},
			},
		},
		{
			name: "hosts_deprecated: url, no url in old - use new default",
			oldMap: map[string]interface{}{
				"other": "x",
			},
			newMap: map[string]interface{}{
				"hosts_deprecated": "url",
				"hosts":            []interface{}{"2.2.2.2"},
			},
			expected: map[string]interface{}{
				"hosts": []interface{}{"2.2.2.2"},
			},
		},
		{
			name: "no _deprecated key - normal merge",
			oldMap: map[string]interface{}{
				"url": "1.1.1.1",
			},
			newMap: map[string]interface{}{
				"hosts": []interface{}{"2.2.2.2"},
			},
			expected: map[string]interface{}{
				"hosts": []interface{}{"2.2.2.2"},
			},
		},
		{
			name: "array to scalar: dsn_deprecated takes first element",
			oldMap: map[string]interface{}{
				"dsn_list": []interface{}{"first-dsn", "second-dsn"},
			},
			newMap: map[string]interface{}{
				"dsn_deprecated": "dsn_list",
				"dsn":            "default-dsn",
			},
			expected: map[string]interface{}{
				"dsn": "first-dsn",
			},
		},
		{
			name: "nested: database_deprecated and dsn_deprecated",
			oldMap: map[string]interface{}{
				"sql": map[string]interface{}{
					"file": map[string]interface{}{
						"name": "1.sql",
					},
					"db": map[string]interface{}{
						"url": "rapid1212312:1232342@tcp(10.77.13.157:3306)/rapidcall",
					},
				},
			},
			newMap: map[string]interface{}{
				"database_deprecated": "sql.db",
				"database": map[string]interface{}{
					"dsn_deprecated": "sql.db.url",
					"dsn":            []interface{}{"rapid:1232342@tcp(localhost:3306)/rapidcall"},
					"max_idle_conn":  10,
					"max_open_conn":  100,
					"use_postgres":   false,
				},
			},
			expected: map[string]interface{}{
				"database": map[string]interface{}{
					"dsn":           []interface{}{"rapid1212312:1232342@tcp(10.77.13.157:3306)/rapidcall"},
					"max_idle_conn": 10,
					"max_open_conn": 100,
					"use_postgres":  false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Merge(tt.newMap, tt.oldMap)

			res, err := json.Marshal(result)
			if err != nil {
				t.Error(err)
			}
			exp, err := json.Marshal(tt.expected)
			if err != nil {
				t.Error(err)
			}

			if string(res) != string(exp) {
				t.Errorf("Expected\n %s, got\n %s", string(exp), string(res))
			}
		})
	}
}

// TestMergeReplaceKeys checks that keys ending with _deprecated_replace in the new config
// override the merged value: result uses the new value instead of the old one.
func TestMergeReplaceKeys(t *testing.T) {
	tests := []struct {
		name     string
		oldMap   map[string]interface{}
		newMap   map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "hosts_deprecated_replace overrides old hosts",
			oldMap: map[string]interface{}{
				"hosts": []interface{}{"1.1.1.1", "2.2.2.2"},
			},
			newMap: map[string]interface{}{
				"hosts_deprecated_replace": "",
				"hosts":                    []interface{}{"9.9.9.9"},
			},
			expected: map[string]interface{}{
				"hosts": []interface{}{"9.9.9.9"},
			},
		},
		{
			name: "nested: database.dsn_deprecated_replace overrides",
			oldMap: map[string]interface{}{
				"database": map[string]interface{}{
					"dsn":      "old-dsn",
					"max_conn": 5,
				},
			},
			newMap: map[string]interface{}{
				"database": map[string]interface{}{
					"dsn_deprecated_replace": "",
					"dsn":                    "new-dsn",
					"max_conn":               10,
				},
			},
			expected: map[string]interface{}{
				"database": map[string]interface{}{
					"dsn":      "new-dsn",
					"max_conn": 5,
				},
			},
		},
		{
			name: "nested: database.dsn_deprecated_replace not exists",
			oldMap: map[string]interface{}{
				"database": map[string]interface{}{
					"max_conn": 5,
				},
			},
			newMap: map[string]interface{}{
				"database": map[string]interface{}{
					"dsn_deprecated_replace": "",
					"dsn":                    "new-dsn",
					"max_conn":               10,
				},
			},
			expected: map[string]interface{}{
				"database": map[string]interface{}{
					"dsn":      "new-dsn",
					"max_conn": 5,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Merge(tt.newMap, tt.oldMap)
			res, _ := json.Marshal(result)
			exp, _ := json.Marshal(tt.expected)
			if string(res) != string(exp) {
				t.Errorf("Expected\n %s, got\n %s", string(exp), string(res))
			}
		})
	}
}

// TestMergeConcatKeys checks key_deprecated_concat: paths are taken from the merged map (after
// _deprecated/_deprecated_expand/_deprecated_collapse), then template {0}, {1} is filled and written to target.
func TestMergeConcatKeys(t *testing.T) {
	tests := []struct {
		name     string
		oldMap   map[string]interface{}
		newMap   map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "concat two top-level fields from merged map",
			oldMap: map[string]interface{}{
				"host": "localhost",
				"port": float64(8080),
			},
			newMap: map[string]interface{}{
				"host":                       "localhost",
				"port":                       float64(8080),
				"endpoint_deprecated_concat": "host,port->{0}:{1}",
			},
			expected: map[string]interface{}{
				"host":     "localhost",
				"port":     float64(8080),
				"endpoint": "localhost:8080",
			},
		},
		{
			name: "concat after deprecated: deprecated fills server.host/port, then concat glues them",
			oldMap: map[string]interface{}{
				"old_host": "db.example.com",
				"old_port": float64(5432),
			},
			newMap: map[string]interface{}{
				"server": map[string]interface{}{
					"host_deprecated":       "old_host",
					"port_deprecated":       "old_port",
					"host":                  "",
					"port":                  float64(0),
					"url_deprecated_concat": "server.host,server.port->{0}:{1}",
				},
			},
			expected: map[string]interface{}{
				"server": map[string]interface{}{
					"host": "db.example.com",
					"port": float64(5432),
					"url":  "db.example.com:5432",
				},
			},
		},
		{
			name: "nested concat",
			oldMap: map[string]interface{}{
				"a": "x",
				"b": "y",
			},
			newMap: map[string]interface{}{
				"a": "x",
				"b": "y",
				"nested": map[string]interface{}{
					"p_deprecated_concat": "a,b->{0}-{1}",
				},
			},
			expected: map[string]interface{}{
				"a": "x",
				"b": "y",
				"nested": map[string]interface{}{
					"p": "x-y",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Merge(tt.newMap, tt.oldMap)
			res, _ := json.Marshal(result)
			exp, _ := json.Marshal(tt.expected)
			if string(res) != string(exp) {
				t.Errorf("Expected\n %s, got\n %s", string(exp), string(res))
			}
		})
	}
}

// assertMerged asserts that Merge(newMap, oldMap) produces expected (compared as JSON).
func assertMerged(t *testing.T, newMap, oldMap, expected map[string]interface{}) {
	t.Helper()
	result := Merge(newMap, oldMap)
	res, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	exp, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}
	if string(res) != string(exp) {
		t.Errorf("Merge result:\n got    %s\n expect %s", string(res), string(exp))
	}
}

// TestMergeAllScenarios covers merge, _deprecated, _deprecated_replace, and their combinations
// across types and nesting so that all code paths and semantics are exercised.
func TestMergeAllScenarios(t *testing.T) {
	t.Run("merge_only", func(t *testing.T) {
		// Same-type scalars: old wins when key in both; keys only in old are not in result (merge only touches keys in new)
		assertMerged(t,
			map[string]interface{}{"a": "new", "b": 2},
			map[string]interface{}{"a": "old", "c": 3},
			map[string]interface{}{"a": "old", "b": 2},
		)
		assertMerged(t,
			map[string]interface{}{"x": true, "y": 1.5},
			map[string]interface{}{"x": false, "z": 2.5},
			map[string]interface{}{"x": false, "y": 1.5},
		)
		// Nested map: recursive merge, old wins for same keys; keys only in old's nested map are not added
		assertMerged(t,
			map[string]interface{}{
				"db": map[string]interface{}{"host": "newhost", "port": 5432},
			},
			map[string]interface{}{
				"db": map[string]interface{}{"host": "oldhost", "name": "mydb"},
			},
			map[string]interface{}{
				"db": map[string]interface{}{"host": "oldhost", "port": 5432},
			},
		)
		// Arrays of maps: merge element-wise; each old element is merged with template (new[0]), new fields added
		assertMerged(t,
			map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"id": 0, "name": "default"},
				},
			},
			map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"id": 1, "name": "one"},
					map[string]interface{}{"id": 2, "name": "two"},
				},
			},
			map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"id": 1, "name": "one"},
					map[string]interface{}{"id": 2, "name": "two"},
				},
			},
		)
		// nil: old nil → nil; new nil → old value
		assertMerged(t,
			map[string]interface{}{"a": "x", "b": nil},
			map[string]interface{}{"a": nil, "b": "y"},
			map[string]interface{}{"a": nil, "b": "y"},
		)
		// Primitive types: bool, int, float64 same-type → old wins
		assertMerged(t,
			map[string]interface{}{"on": false, "n": 0, "ratio": 0.5},
			map[string]interface{}{"on": true, "n": 42, "ratio": 0.99},
			map[string]interface{}{"on": true, "n": 42, "ratio": 0.99},
		)
		// Add new field to each element of array of maps: template has use_tls, old has url/certificate
		assertMerged(t,
			map[string]interface{}{
				"fps": map[string]interface{}{
					"urls": []interface{}{
						map[string]interface{}{"url": "url_example", "certificate": "cert_example.cert", "use_tls": false},
					},
				},
			},
			map[string]interface{}{
				"fps": map[string]interface{}{
					"urls": []interface{}{
						map[string]interface{}{"url": "url1", "certificate": "crt2.cert"},
						map[string]interface{}{"url": "url2", "certificate": "crt2.cert"},
						map[string]interface{}{"url": "url3", "certificate": "crt3.cert"},
					},
				},
			},
			map[string]interface{}{
				"fps": map[string]interface{}{
					"urls": []interface{}{
						map[string]interface{}{"url": "url1", "certificate": "crt2.cert", "use_tls": false},
						map[string]interface{}{"url": "url2", "certificate": "crt2.cert", "use_tls": false},
						map[string]interface{}{"url": "url3", "certificate": "crt3.cert", "use_tls": false},
					},
				},
			},
		)
	})

	t.Run("deprecated_only", func(t *testing.T) {
		// Scalar → array: deprecated value wrapped in array
		assertMerged(t,
			map[string]interface{}{
				"hosts_deprecated": "url",
				"hosts":            []interface{}{"default"},
			},
			map[string]interface{}{"url": "1.2.3.4"},
			map[string]interface{}{"hosts": []interface{}{"1.2.3.4"}},
		)
		// Old has both target and deprecated path: concatenate
		assertMerged(t,
			map[string]interface{}{
				"hosts_deprecated": "url",
				"hosts":            []interface{}{"new"},
			},
			map[string]interface{}{"url": "old-url", "hosts": []interface{}{"old-host"}},
			map[string]interface{}{"hosts": []interface{}{"old-host", "old-url"}},
		)
		// Deprecated path missing in old: keep new default
		assertMerged(t,
			map[string]interface{}{
				"hosts_deprecated": "missing",
				"hosts":            []interface{}{"default"},
			},
			map[string]interface{}{"other": 1},
			map[string]interface{}{"hosts": []interface{}{"default"}},
		)
		// Array → scalar: first element
		assertMerged(t,
			map[string]interface{}{
				"name_deprecated": "names",
				"name":            "default-name",
			},
			map[string]interface{}{"names": []interface{}{"alice", "bob"}},
			map[string]interface{}{"name": "alice"},
		)
		// Nested deprecated path (full path from root old)
		assertMerged(t,
			map[string]interface{}{
				"server": map[string]interface{}{
					"addr_deprecated": "legacy.host",
					"addr":            "0.0.0.0",
				},
			},
			map[string]interface{}{
				"legacy": map[string]interface{}{"host": "10.0.0.1"},
			},
			map[string]interface{}{
				"server": map[string]interface{}{"addr": "10.0.0.1"},
			},
		)
		// Deprecated path is nested (a.b.c)
		assertMerged(t,
			map[string]interface{}{
				"value_deprecated": "level1.level2.key",
				"value":            "default",
			},
			map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{"key": "from-deep"},
				},
			},
			map[string]interface{}{"value": "from-deep"},
		)
	})

	t.Run("replace_only", func(t *testing.T) {
		// Top-level: new value overwrites merged (old) value
		assertMerged(t,
			map[string]interface{}{
				"port_deprecated_replace": "",
				"port":                    8080,
			},
			map[string]interface{}{"port": 3000},
			map[string]interface{}{"port": 8080},
		)
		// Nested replace
		assertMerged(t,
			map[string]interface{}{
				"db": map[string]interface{}{
					"driver_deprecated_replace": "",
					"driver":                    "postgres",
					"pool":                      20,
				},
			},
			map[string]interface{}{
				"db": map[string]interface{}{"driver": "mysql", "pool": 10},
			},
			map[string]interface{}{
				"db": map[string]interface{}{"driver": "postgres", "pool": 10},
			},
		)
		// _deprecated_replace present but no sibling target in new: we don't set (target not in new); merge didn't add key from old
		assertMerged(t,
			map[string]interface{}{
				"missing_target_deprecated_replace": "", // no "missing_target" in new
			},
			map[string]interface{}{"missing_target": "from-old"},
			map[string]interface{}{},
		)
	})

	t.Run("replace_and_deprecated_same_key", func(t *testing.T) {
		// Replace runs after deprecated → final value is from new (replace wins)
		assertMerged(t,
			map[string]interface{}{
				"hosts_deprecated":         "url",
				"hosts_deprecated_replace": "",
				"hosts":                    []interface{}{"replacement"},
			},
			map[string]interface{}{"url": "1.2.3.4", "hosts": []interface{}{"old"}},
			map[string]interface{}{"hosts": []interface{}{"replacement"}},
		)
	})

	t.Run("replace_and_deprecated_different_keys", func(t *testing.T) {
		assertMerged(t,
			map[string]interface{}{
				"hosts_deprecated":        "url",
				"hosts":                   []interface{}{"new-default"},
				"port_deprecated_replace": "",
				"port":                    443,
			},
			map[string]interface{}{"url": "1.2.3.4", "port": 80},
			map[string]interface{}{
				"hosts": []interface{}{"1.2.3.4"},
				"port":  443,
			},
		)
	})

	t.Run("nested_deprecated_replace_and_deprecated", func(t *testing.T) {
		assertMerged(t,
			map[string]interface{}{
				"database": map[string]interface{}{
					"dsn_deprecated":         "legacy.connection",
					"dsn_deprecated_replace": "",
					"dsn":                    "new-dsn",
					"max_conn":               100,
				},
			},
			map[string]interface{}{
				"legacy":   map[string]interface{}{"connection": "old-dsn"},
				"database": map[string]interface{}{"max_conn": 5},
			},
			map[string]interface{}{
				"database": map[string]interface{}{"dsn": "new-dsn", "max_conn": 5},
			},
		)
	})

	t.Run("types_with_deprecated_replace_and_deprecated", func(t *testing.T) {
		// bool replace
		assertMerged(t,
			map[string]interface{}{"debug_deprecated_replace": "", "debug": true},
			map[string]interface{}{"debug": false},
			map[string]interface{}{"debug": true},
		)
		// int deprecated (scalar)
		assertMerged(t,
			map[string]interface{}{
				"timeout_deprecated": "old_timeout",
				"timeout":            30,
			},
			map[string]interface{}{"old_timeout": 60},
			map[string]interface{}{"timeout": 60},
		)
		// string deprecated → array (new wants array)
		assertMerged(t,
			map[string]interface{}{
				"peers_deprecated": "master",
				"peers":            []interface{}{"b"},
			},
			map[string]interface{}{"master": "a"},
			map[string]interface{}{"peers": []interface{}{"a"}},
		)
	})

	t.Run("edge_empty_and_nil", func(t *testing.T) {
		assertMerged(t, nil, nil, map[string]interface{}{})
		assertMerged(t, map[string]interface{}{}, map[string]interface{}{}, map[string]interface{}{})
		assertMerged(t, map[string]interface{}{"a": 1}, nil, map[string]interface{}{"a": 1})
		assertMerged(t, nil, map[string]interface{}{"a": 1}, map[string]interface{}{})
		assertMerged(t, map[string]interface{}{}, map[string]interface{}{"a": 1}, map[string]interface{}{})
	})

	t.Run("edge_deprecated_empty_array", func(t *testing.T) {
		// Deprecated path points to empty array → mergeDeprecatedIntoTarget returns oldTarget (nil), so list becomes nil
		assertMerged(t,
			map[string]interface{}{
				"list_deprecated": "empty_list",
				"list":            []interface{}{"x"},
			},
			map[string]interface{}{"empty_list": []interface{}{}},
			map[string]interface{}{"list": nil},
		)
	})

	t.Run("edge_multiple_deprecated_same_map", func(t *testing.T) {
		assertMerged(t,
			map[string]interface{}{
				"a_deprecated": "x",
				"a":            "default-a",
				"b_deprecated": "y",
				"b":            "default-b",
			},
			map[string]interface{}{"x": "val-x", "y": "val-y"},
			map[string]interface{}{"a": "val-x", "b": "val-y"},
		)
	})

	t.Run("edge_multiple_deprecated_replace_same_map", func(t *testing.T) {
		assertMerged(t,
			map[string]interface{}{
				"a_deprecated_replace": "", "a": 1,
				"b_deprecated_replace": "", "b": 2,
			},
			map[string]interface{}{"a": 10, "b": 20},
			map[string]interface{}{"a": 1, "b": 2},
		)
	})

	t.Run("edge_deprecated_replace_different_types", func(t *testing.T) {
		// Replace can set any type from new (here: array overwrites scalar)
		assertMerged(t,
			map[string]interface{}{
				"tags_deprecated_replace": "",
				"tags":                    []interface{}{"a", "b"},
			},
			map[string]interface{}{"tags": "single"},
			map[string]interface{}{"tags": []interface{}{"a", "b"}},
		)
	})

	t.Run("edge_key_only_in_new", func(t *testing.T) {
		// Key only in new stays as in new (nothing to merge from old)
		assertMerged(t,
			map[string]interface{}{"new_only": 1},
			map[string]interface{}{"old_only": 2},
			map[string]interface{}{"new_only": 1},
		)
	})

	t.Run("edge_array_of_primitives", func(t *testing.T) {
		// Both arrays but elements not maps → old wins
		assertMerged(t,
			map[string]interface{}{"ports": []interface{}{80, 443}},
			map[string]interface{}{"ports": []interface{}{8080, 8443}},
			map[string]interface{}{"ports": []interface{}{8080, 8443}},
		)
	})
}

func TestMergeReplace(t *testing.T) {
	replacer.Register("__test_replacer__", func() string {
		return "-TEST_STRING-"
	})

	tests := []struct {
		name     string
		oldMap   map[string]interface{}
		newMap   map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:   "New map",
			oldMap: map[string]interface{}{},
			newMap: map[string]interface{}{
				"foo":  "some_string__test_replacer__some_string",
				"foo2": "some_string__test_replacer__",
				"foo3": "__test_replacer__",
				"foo4": "__test_replacer__some_string",
			},
			expected: map[string]interface{}{
				"foo":  "some_string-TEST_STRING-some_string",
				"foo2": "some_string-TEST_STRING-",
				"foo3": "-TEST_STRING-",
				"foo4": "-TEST_STRING-some_string",
			},
		},
		{
			name:   "New array",
			oldMap: map[string]interface{}{},
			newMap: map[string]interface{}{
				"foo": []interface{}{"some_string__test_replacer__some_string"},
			},
			expected: map[string]interface{}{
				"foo": []interface{}{"some_string-TEST_STRING-some_string"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Merge(tt.newMap, tt.oldMap)

			res, err := json.Marshal(result)
			if err != nil {
				t.Error(err)
			}
			exp, err := json.Marshal(tt.expected)
			if err != nil {
				t.Error(err)
			}

			if string(res) != string(exp) {
				t.Errorf("Expected\n %s, got\n %s", string(exp), string(res))
			}
		})
	}
}

// Benchmarks for merge only; no deprecated or replacer logic.
func BenchmarkMergeMapsFlat(b *testing.B) {
	newMap := map[string]interface{}{
		"string": "value",
		"int":    1,
		"bool":   true,
	}

	oldMap := map[string]interface{}{
		"string": "old_value",
		"int":    99,
		"bool":   false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mergeMaps(newMap, oldMap)
	}
}

func BenchmarkMergeMapsNested(b *testing.B) {
	newMap := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Alice",
			"age":  30,
		},
	}

	oldMap := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Bob",
			"city": "New York",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mergeMaps(newMap, oldMap)
	}
}

func BenchmarkMergeMapsDeepNestedArray(b *testing.B) {
	newMap := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{
				"id":   0,
				"name": "default",
			},
		},
	}

	oldMap := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{
				"id":   1,
				"name": "item1",
			},
			map[string]interface{}{
				"id":   2,
				"name": "item2",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mergeMaps(newMap, oldMap)
	}
}

func generateDeepNestedMap() map[string]interface{} {
	return map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": map[string]interface{}{
					"string": "value",
					"number": 123,
					"bool":   true,
					"array": []interface{}{
						map[string]interface{}{
							"id":   1,
							"name": "item1",
							"meta": map[string]interface{}{
								"enabled": true,
								"tags":    []interface{}{"a", "b", "c"},
							},
						},
						map[string]interface{}{
							"id":   2,
							"name": "item2",
							"meta": map[string]interface{}{
								"enabled": false,
								"tags":    []interface{}{},
							},
						},
					},
				},
			},
		},
	}
}

func BenchmarkMergeMaps_DeepNestedLarge(b *testing.B) {
	newMap := generateDeepNestedMap()
	oldMap := generateDeepNestedMap()

	// Изменим значения в oldMap для "реального" слияния
	oldMap["level1"].(map[string]interface{})["level2"].(map[string]interface{})["level3"].(map[string]interface{})["string"] = "oldValue"
	oldMap["level1"].(map[string]interface{})["level2"].(map[string]interface{})["level3"].(map[string]interface{})["new_key"] = "added"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mergeMaps(newMap, oldMap)
	}
}

func generateMassiveMap(from, to int) map[string]interface{} {
	m := make(map[string]interface{}, from-to)
	for i := from; i < to; i++ {
		key := "key_" + fmt.Sprint(i)
		m[key] = map[string]interface{}{
			"nested": map[string]interface{}{
				"value": i,
				"bool":  i%2 == 1,
				"array": []interface{}{
					map[string]interface{}{
						"id":   i,
						"name": fmt.Sprintf("item_%d", i),
					},
					map[string]interface{}{
						"id":   i,
						"name": fmt.Sprintf("item_%d", i),
					},
					map[string]interface{}{
						"id":   i,
						"name": fmt.Sprintf("item_%d", i),
					},
				},
			},
		}
	}
	return m
}

func BenchmarkMergeMaps_Massive(b *testing.B) {
	newMap := generateMassiveMap(0, 1000)
	oldMap := generateMassiveMap(200, 1200)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mergeMaps(newMap, oldMap)
	}
}
