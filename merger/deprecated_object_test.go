package merger

import (
	"encoding/json"
	"testing"
)

// assertJSONEqual compares via JSON so map key order does not matter.
func assertJSONEqual(t *testing.T, expected, got map[string]interface{}) {
	t.Helper()
	exp, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}
	res, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if string(exp) != string(res) {
		t.Errorf("expected\n%s\ngot\n%s", exp, res)
	}
}

func sampleRTC() map[string]interface{} {
	return map[string]interface{}{
		"port_range_start":            50000,
		"port_range_end":              60000,
		"allow_tcp_fallback":          true,
		"tcp_fallback_rtt_threshold":  150,
		"allow_udp_unstable_fallback": true,
	}
}

func userRTC() map[string]interface{} {
	return map[string]interface{}{
		"port_range_start":            51000,
		"port_range_end":              60000,
		"allow_tcp_fallback":          true,
		"tcp_fallback_rtt_threshold":  150,
		"allow_udp_unstable_fallback": true,
		"user_ice_servers":            []interface{}{"stun:custom.example"},
		"vendor": map[string]interface{}{
			"flag":  "on",
			"extra": 42,
		},
	}
}

func vcsWithRTC(rtc map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"vcs": map[string]interface{}{
			"tcp": map[string]interface{}{"port": 8059},
			"webrtc": map[string]interface{}{
				"rtc": rtc,
			},
		},
	}
}

// TestDeprecatedObject_allVariants is the safety net for webrtc.rtc-style object
// preserve/extend across migrations. Every production-relevant variant must stay green.
func TestDeprecatedObject_allVariants(t *testing.T) {
	tests := []struct {
		name     string
		oldMap   map[string]interface{}
		newMap   map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:     "01_first_migration_introduces_rtc_defaults",
			oldMap:   map[string]interface{}{},
			newMap:   vcsWithRTC(sampleRTC()),
			expected: vcsWithRTC(sampleRTC()),
		},
		{
			name:   "02_preserve_only_marker_keeps_user_keys",
			oldMap: vcsWithRTC(userRTC()),
			newMap: map[string]interface{}{
				"vcs": map[string]interface{}{
					"tcp": map[string]interface{}{"port": 8059},
					"webrtc": map[string]interface{}{
						"rtc_deprecated": "vcs.webrtc.rtc",
					},
				},
			},
			expected: vcsWithRTC(userRTC()),
		},
		{
			name:   "03_empty_rtc_template_equals_full_preserve",
			oldMap: vcsWithRTC(userRTC()),
			newMap: map[string]interface{}{
				"vcs": map[string]interface{}{
					"tcp": map[string]interface{}{"port": 8059},
					"webrtc": map[string]interface{}{
						"rtc_deprecated": "vcs.webrtc.rtc",
						"rtc":            map[string]interface{}{},
					},
				},
			},
			expected: vcsWithRTC(userRTC()),
		},
		{
			name:   "04_omit_rtc_without_deprecated_DROPS_subtree",
			oldMap: vcsWithRTC(userRTC()),
			newMap: map[string]interface{}{
				"vcs": map[string]interface{}{
					"tcp":    map[string]interface{}{"port": 8059},
					"webrtc": map[string]interface{}{},
				},
			},
			expected: map[string]interface{}{
				"vcs": map[string]interface{}{
					"tcp":    map[string]interface{}{"port": 8059},
					"webrtc": map[string]interface{}{},
				},
			},
		},
		{
			name:   "05_add_top_level_key_keeps_user_data",
			oldMap: vcsWithRTC(userRTC()),
			newMap: map[string]interface{}{
				"vcs": map[string]interface{}{
					"tcp": map[string]interface{}{"port": 8059},
					"webrtc": map[string]interface{}{
						"rtc_deprecated": "vcs.webrtc.rtc",
						"rtc": map[string]interface{}{
							"new_setting": false,
						},
					},
				},
			},
			expected: vcsWithRTC(map[string]interface{}{
				"port_range_start":            51000,
				"port_range_end":              60000,
				"allow_tcp_fallback":          true,
				"tcp_fallback_rtt_threshold":  150,
				"allow_udp_unstable_fallback": true,
				"user_ice_servers":            []interface{}{"stun:custom.example"},
				"vendor": map[string]interface{}{
					"flag":  "on",
					"extra": 42,
				},
				"new_setting": false,
			}),
		},
		{
			name:   "06_add_nested_key_keeps_sibling_user_keys",
			oldMap: vcsWithRTC(userRTC()),
			newMap: map[string]interface{}{
				"vcs": map[string]interface{}{
					"tcp": map[string]interface{}{"port": 8059},
					"webrtc": map[string]interface{}{
						"rtc_deprecated": "vcs.webrtc.rtc",
						"rtc": map[string]interface{}{
							"vendor": map[string]interface{}{
								"new_vendor_flag": true,
							},
						},
					},
				},
			},
			expected: vcsWithRTC(map[string]interface{}{
				"port_range_start":            51000,
				"port_range_end":              60000,
				"allow_tcp_fallback":          true,
				"tcp_fallback_rtt_threshold":  150,
				"allow_udp_unstable_fallback": true,
				"user_ice_servers":            []interface{}{"stun:custom.example"},
				"vendor": map[string]interface{}{
					"flag":            "on",
					"extra":           42,
					"new_vendor_flag": true,
				},
			}),
		},
		{
			name:   "07_template_defaults_do_not_overwrite_existing",
			oldMap: vcsWithRTC(userRTC()),
			newMap: map[string]interface{}{
				"vcs": map[string]interface{}{
					"tcp": map[string]interface{}{"port": 8059},
					"webrtc": map[string]interface{}{
						"rtc_deprecated": "vcs.webrtc.rtc",
						"rtc": map[string]interface{}{
							"port_range_start": 1,
							"port_range_end":   2,
							"new_setting":      true,
						},
					},
				},
			},
			expected: vcsWithRTC(map[string]interface{}{
				"port_range_start":            51000, // user value kept
				"port_range_end":              60000, // user value kept
				"allow_tcp_fallback":          true,
				"tcp_fallback_rtt_threshold":  150,
				"allow_udp_unstable_fallback": true,
				"user_ice_servers":            []interface{}{"stun:custom.example"},
				"vendor": map[string]interface{}{
					"flag":  "on",
					"extra": 42,
				},
				"new_setting": true,
			}),
		},
		{
			name:   "08_replace_one_field_forces_new_keeps_rest",
			oldMap: vcsWithRTC(userRTC()),
			newMap: map[string]interface{}{
				"vcs": map[string]interface{}{
					"tcp": map[string]interface{}{"port": 8059},
					"webrtc": map[string]interface{}{
						"rtc_deprecated": "vcs.webrtc.rtc",
						"rtc": map[string]interface{}{
							"port_range_start_deprecated_replace": "",
							"port_range_start":                    50000,
							"brand_new":                           "x",
						},
					},
				},
			},
			expected: vcsWithRTC(map[string]interface{}{
				"port_range_start":            50000, // replaced
				"port_range_end":              60000,
				"allow_tcp_fallback":          true,
				"tcp_fallback_rtt_threshold":  150,
				"allow_udp_unstable_fallback": true,
				"user_ice_servers":            []interface{}{"stun:custom.example"},
				"vendor": map[string]interface{}{
					"flag":  "on",
					"extra": 42,
				},
				"brand_new": "x",
			}),
		},
		{
			name:   "09_deprecated_path_missing_keeps_template",
			oldMap: map[string]interface{}{"vcs": map[string]interface{}{"tcp": map[string]interface{}{"port": 1}}},
			newMap: map[string]interface{}{
				"vcs": map[string]interface{}{
					"tcp": map[string]interface{}{"port": 1},
					"webrtc": map[string]interface{}{
						"rtc_deprecated": "vcs.webrtc.rtc",
						"rtc":            sampleRTC(),
					},
				},
			},
			expected: map[string]interface{}{
				"vcs": map[string]interface{}{
					"tcp": map[string]interface{}{"port": 1},
					"webrtc": map[string]interface{}{
						"rtc": sampleRTC(),
					},
				},
			},
		},
		{
			name:   "10_marker_only_path_missing_no_rtc",
			oldMap: map[string]interface{}{"vcs": map[string]interface{}{"tcp": map[string]interface{}{"port": 1}}},
			newMap: map[string]interface{}{
				"vcs": map[string]interface{}{
					"tcp": map[string]interface{}{"port": 1},
					"webrtc": map[string]interface{}{
						"rtc_deprecated": "vcs.webrtc.rtc",
					},
				},
			},
			expected: map[string]interface{}{
				"vcs": map[string]interface{}{
					"tcp":    map[string]interface{}{"port": 1},
					"webrtc": map[string]interface{}{},
				},
			},
		},
		{
			name:   "11_sibling_keys_still_merge_normally",
			oldMap: vcsWithRTC(userRTC()),
			newMap: map[string]interface{}{
				"vcs": map[string]interface{}{
					"tcp": map[string]interface{}{
						"port_deprecated_replace": "",
						"port":                    9000,
						"host":                    "localhost",
					},
					"webrtc": map[string]interface{}{
						"rtc_deprecated": "vcs.webrtc.rtc",
					},
					"recording": map[string]interface{}{
						"filepath_prefix": "records",
					},
				},
			},
			expected: map[string]interface{}{
				"vcs": map[string]interface{}{
					"tcp": map[string]interface{}{
						"port": 9000,
						"host": "localhost",
					},
					"webrtc": map[string]interface{}{
						"rtc": userRTC(),
					},
					"recording": map[string]interface{}{
						"filepath_prefix": "records",
					},
				},
			},
		},
		{
			name:   "12_full_schema_template_plus_deprecated_still_keeps_unknown",
			oldMap: vcsWithRTC(userRTC()),
			newMap: map[string]interface{}{
				"vcs": map[string]interface{}{
					"tcp": map[string]interface{}{"port": 8059},
					"webrtc": map[string]interface{}{
						"rtc_deprecated": "vcs.webrtc.rtc",
						"rtc":            sampleRTC(), // full known schema, no user keys
					},
				},
			},
			expected: vcsWithRTC(userRTC()), // user keys survive even if not in template
		},
		{
			name: "13_array_user_value_preserved_when_template_has_default_array",
			oldMap: vcsWithRTC(map[string]interface{}{
				"codecs": []interface{}{"opus", "custom"},
				"other":  "x",
			}),
			newMap: map[string]interface{}{
				"vcs": map[string]interface{}{
					"tcp": map[string]interface{}{"port": 8059},
					"webrtc": map[string]interface{}{
						"rtc_deprecated": "vcs.webrtc.rtc",
						"rtc": map[string]interface{}{
							"codecs": []interface{}{"default"},
							"mode":   "safe",
						},
					},
				},
			},
			expected: vcsWithRTC(map[string]interface{}{
				"codecs": []interface{}{"opus", "custom"}, // old array kept
				"other":  "x",
				"mode":   "safe",
			}),
		},
		{
			name: "14_bool_int_string_types_preserved_and_added",
			oldMap: vcsWithRTC(map[string]interface{}{
				"enabled": true,
				"count":   3,
				"name":    "user",
			}),
			newMap: map[string]interface{}{
				"vcs": map[string]interface{}{
					"tcp": map[string]interface{}{"port": 8059},
					"webrtc": map[string]interface{}{
						"rtc_deprecated": "vcs.webrtc.rtc",
						"rtc": map[string]interface{}{
							"enabled": false,
							"count":   99,
							"name":    "default",
							"ratio":   1.5,
							"ok":      true,
						},
					},
				},
			},
			expected: vcsWithRTC(map[string]interface{}{
				"enabled": true,
				"count":   3,
				"name":    "user",
				"ratio":   1.5,
				"ok":      true,
			}),
		},
		{
			name:   "15_marker_stripped_from_result",
			oldMap: vcsWithRTC(sampleRTC()),
			newMap: map[string]interface{}{
				"vcs": map[string]interface{}{
					"tcp": map[string]interface{}{"port": 8059},
					"webrtc": map[string]interface{}{
						"rtc_deprecated": "vcs.webrtc.rtc",
						"rtc": map[string]interface{}{
							"x": 1,
						},
					},
				},
			},
			expected: vcsWithRTC(map[string]interface{}{
				"port_range_start":            50000,
				"port_range_end":              60000,
				"allow_tcp_fallback":          true,
				"tcp_fallback_rtt_threshold":  150,
				"allow_udp_unstable_fallback": true,
				"x":                           1,
			}),
		},
		{
			name: "16_deep_nested_extend_three_levels",
			oldMap: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": map[string]interface{}{
							"keep": "yes",
							"n": map[string]interface{}{
								"u": 1,
							},
						},
					},
				},
			},
			newMap: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c_deprecated": "a.b.c",
						"c": map[string]interface{}{
							"added": "new",
							"n": map[string]interface{}{
								"v": 2,
							},
						},
					},
				},
			},
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": map[string]interface{}{
							"keep":  "yes",
							"added": "new",
							"n": map[string]interface{}{
								"u": 1,
								"v": 2,
							},
						},
					},
				},
			},
		},
		{
			name: "17_map_deprecated_from_other_path_copies_then_extends",
			oldMap: map[string]interface{}{
				"legacy": map[string]interface{}{
					"rtc": map[string]interface{}{
						"old_only": "keep",
						"shared":   "from-old",
					},
				},
			},
			newMap: map[string]interface{}{
				"webrtc": map[string]interface{}{
					"rtc_deprecated": "legacy.rtc",
					"rtc": map[string]interface{}{
						"shared":   "from-new",
						"new_only": true,
					},
				},
			},
			expected: map[string]interface{}{
				"webrtc": map[string]interface{}{
					"rtc": map[string]interface{}{
						"old_only": "keep",
						"shared":   "from-old",
						"new_only": true,
					},
				},
			},
		},
		{
			name: "18_classic_scalar_deprecated_unchanged",
			oldMap: map[string]interface{}{
				"url": "1.2.3.4",
			},
			newMap: map[string]interface{}{
				"hosts_deprecated": "url",
				"hosts":            []interface{}{"default"},
			},
			expected: map[string]interface{}{
				"hosts": []interface{}{"1.2.3.4"},
			},
		},
		{
			name: "19_classic_array_to_scalar_unchanged",
			oldMap: map[string]interface{}{
				"dsn_list": []interface{}{"first", "second"},
			},
			newMap: map[string]interface{}{
				"dsn_deprecated": "dsn_list",
				"dsn":            "default",
			},
			expected: map[string]interface{}{
				"dsn": "first",
			},
		},
		{
			name:   "20_without_deprecated_normal_merge_drops_unknown_nested_keys",
			oldMap: vcsWithRTC(userRTC()),
			newMap: map[string]interface{}{
				"vcs": map[string]interface{}{
					"tcp": map[string]interface{}{"port": 8059},
					"webrtc": map[string]interface{}{
						"rtc": sampleRTC(),
					},
				},
			},
			expected: vcsWithRTC(map[string]interface{}{
				"port_range_start":            51000,
				"port_range_end":              60000,
				"allow_tcp_fallback":          true,
				"tcp_fallback_rtt_threshold":  150,
				"allow_udp_unstable_fallback": true,
				// user_ice_servers and vendor DROPPED — no rtc_deprecated
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Merge(deepCopyMap(tt.newMap), deepCopyMap(tt.oldMap))
			assertJSONEqual(t, tt.expected, got)
		})
	}
}

// TestDeprecatedObject_migrationChain applies many sequential migrations the way
// production does: v1 introduces rtc, user edits, then N migrations preserve/extend.
func TestDeprecatedObject_migrationChain(t *testing.T) {
	cfg := Merge(vcsWithRTC(sampleRTC()), map[string]interface{}{})

	// User edits after v1.
	rtc := cfg["vcs"].(map[string]interface{})["webrtc"].(map[string]interface{})["rtc"].(map[string]interface{})
	rtc["port_range_start"] = 51000
	rtc["user_ice_servers"] = []interface{}{"stun:a", "stun:b"}
	rtc["vendor"] = map[string]interface{}{"flag": "on", "extra": 7}

	preserveMigration := map[string]interface{}{
		"vcs": map[string]interface{}{
			"tcp": map[string]interface{}{
				"port_deprecated_replace": "",
				"port":                    8059,
			},
			"webrtc": map[string]interface{}{
				"rtc_deprecated": "vcs.webrtc.rtc",
			},
			"recording": map[string]interface{}{
				"filepath_prefix": "records",
			},
		},
	}

	for i := 0; i < 25; i++ {
		cfg = Merge(deepCopyMap(preserveMigration), cfg)
	}

	gotRTC := cfg["vcs"].(map[string]interface{})["webrtc"].(map[string]interface{})["rtc"].(map[string]interface{})
	if gotRTC["port_range_start"] != 51000 {
		t.Fatalf("after 25 preserve migrations port_range_start=%v", gotRTC["port_range_start"])
	}
	if gotRTC["user_ice_servers"] == nil {
		t.Fatal("user_ice_servers lost after preserve chain")
	}
	vendor := gotRTC["vendor"].(map[string]interface{})
	if vendor["flag"] != "on" || vendor["extra"] != 7 {
		t.Fatalf("vendor corrupted: %v", vendor)
	}

	// One migration adds keys.
	cfg = Merge(map[string]interface{}{
		"vcs": map[string]interface{}{
			"tcp": map[string]interface{}{
				"port_deprecated_replace": "",
				"port":                    8059,
			},
			"webrtc": map[string]interface{}{
				"rtc_deprecated": "vcs.webrtc.rtc",
				"rtc": map[string]interface{}{
					"new_a": 1,
					"vendor": map[string]interface{}{
						"new_b": "bb",
					},
				},
			},
			"recording": map[string]interface{}{
				"filepath_prefix": "records",
			},
		},
	}, cfg)

	gotRTC = cfg["vcs"].(map[string]interface{})["webrtc"].(map[string]interface{})["rtc"].(map[string]interface{})
	if gotRTC["new_a"] != 1 {
		t.Fatalf("new_a not added: %v", gotRTC["new_a"])
	}
	vendor = gotRTC["vendor"].(map[string]interface{})
	if vendor["flag"] != "on" || vendor["extra"] != 7 || vendor["new_b"] != "bb" {
		t.Fatalf("vendor after extend: %v", vendor)
	}
	if gotRTC["port_range_start"] != 51000 {
		t.Fatalf("user value lost on extend: %v", gotRTC["port_range_start"])
	}

	// More preserve migrations after extend.
	for i := 0; i < 10; i++ {
		cfg = Merge(deepCopyMap(preserveMigration), cfg)
	}
	gotRTC = cfg["vcs"].(map[string]interface{})["webrtc"].(map[string]interface{})["rtc"].(map[string]interface{})
	if gotRTC["new_a"] != 1 || gotRTC["port_range_start"] != 51000 {
		t.Fatalf("lost data after post-extend preserve: %#v", gotRTC)
	}
	vendor = gotRTC["vendor"].(map[string]interface{})
	if vendor["new_b"] != "bb" || vendor["extra"] != 7 {
		t.Fatalf("vendor lost after post-extend preserve: %v", vendor)
	}

	// Force replace one field late in the chain.
	cfg = Merge(map[string]interface{}{
		"vcs": map[string]interface{}{
			"tcp": map[string]interface{}{"port": 8059},
			"webrtc": map[string]interface{}{
				"rtc_deprecated": "vcs.webrtc.rtc",
				"rtc": map[string]interface{}{
					"port_range_end_deprecated_replace": "",
					"port_range_end":                    70000,
				},
			},
			"recording": map[string]interface{}{"filepath_prefix": "records"},
		},
	}, cfg)
	gotRTC = cfg["vcs"].(map[string]interface{})["webrtc"].(map[string]interface{})["rtc"].(map[string]interface{})
	if gotRTC["port_range_end"] != 70000 {
		t.Fatalf("replace failed: %v", gotRTC["port_range_end"])
	}
	if gotRTC["port_range_start"] != 51000 || gotRTC["new_a"] != 1 {
		t.Fatalf("replace leaked into other fields: %#v", gotRTC)
	}
}

// TestDeprecatedObject_iterationOrderStress runs the same merge many times.
// Map iteration order is random; result must be stable.
func TestDeprecatedObject_iterationOrderStress(t *testing.T) {
	oldMap := vcsWithRTC(userRTC())
	newMap := map[string]interface{}{
		"vcs": map[string]interface{}{
			"tcp": map[string]interface{}{"port": 8059},
			"webrtc": map[string]interface{}{
				"rtc_deprecated": "vcs.webrtc.rtc",
				"rtc": map[string]interface{}{
					"new_setting":                         false,
					"port_range_start_deprecated_replace": "",
					"port_range_start":                    1,
					"vendor": map[string]interface{}{
						"added": true,
					},
				},
			},
		},
	}
	expected := vcsWithRTC(map[string]interface{}{
		"port_range_start":            1, // replaced
		"port_range_end":              60000,
		"allow_tcp_fallback":          true,
		"tcp_fallback_rtt_threshold":  150,
		"allow_udp_unstable_fallback": true,
		"user_ice_servers":            []interface{}{"stun:custom.example"},
		"vendor": map[string]interface{}{
			"flag":  "on",
			"extra": 42,
			"added": true,
		},
		"new_setting": false,
	})

	expJSON, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 200; i++ {
		got := Merge(deepCopyMap(newMap), deepCopyMap(oldMap))
		gotJSON, err := json.Marshal(got)
		if err != nil {
			t.Fatal(err)
		}
		if string(expJSON) != string(gotJSON) {
			t.Fatalf("iteration %d unstable result\nwant %s\ngot  %s", i, expJSON, gotJSON)
		}
	}
}

// TestDeprecatedObject_doesNotBreakExpandCollapseConcat ensures object-extend
// changes did not regress other deprecated operators.
func TestDeprecatedObject_doesNotBreakExpandCollapseConcat(t *testing.T) {
	t.Run("expand", func(t *testing.T) {
		got := Merge(
			map[string]interface{}{
				"items_deprecated_expand": "list->value",
				"items": []interface{}{
					map[string]interface{}{"label": "a"},
				},
			},
			map[string]interface{}{
				"list": []interface{}{"one", "two"},
			},
		)
		assertJSONEqual(t, map[string]interface{}{
			"items": []interface{}{
				map[string]interface{}{"label": "a", "value": "one"},
				map[string]interface{}{"label": "a", "value": "two"},
			},
		}, got)
	})

	t.Run("collapse", func(t *testing.T) {
		got := Merge(
			map[string]interface{}{
				"urls_deprecated_collapse": "fps.urls.url->fps.urls",
				"fps": map[string]interface{}{
					"urls": []interface{}{
						map[string]interface{}{"url": "x", "certificate": "c"},
					},
				},
			},
			map[string]interface{}{
				"fps": map[string]interface{}{
					"urls": []interface{}{
						map[string]interface{}{"url": "u1", "certificate": "c1"},
						map[string]interface{}{"url": "u2", "certificate": "c2"},
					},
				},
			},
		)
		assertJSONEqual(t, map[string]interface{}{
			"fps": map[string]interface{}{
				"urls": []interface{}{"u1", "u2"},
			},
		}, got)
	})

	t.Run("concat", func(t *testing.T) {
		got := Merge(
			map[string]interface{}{
				"endpoint_deprecated_concat": "host,port->{0}:{1}",
				"host":                       "localhost",
				"port":                       8080,
				"endpoint":                   "",
			},
			map[string]interface{}{
				"host": "localhost",
				"port": 8080,
			},
		)
		assertJSONEqual(t, map[string]interface{}{
			"host":     "localhost",
			"port":     8080,
			"endpoint": "localhost:8080",
		}, got)
	})

	t.Run("replace", func(t *testing.T) {
		got := Merge(
			map[string]interface{}{
				"port_deprecated_replace": "",
				"port":                    443,
			},
			map[string]interface{}{"port": 80},
		)
		assertJSONEqual(t, map[string]interface{}{"port": 443}, got)
	})
}
