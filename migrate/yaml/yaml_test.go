package yaml

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/c2pc/config-migrate/internal/url"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gopkg.in/yaml.v3"
)

const configPath = "./examples/config.yaml"
const migrationsPath = "./examples/migrations"

func getConfigURL() string {
	return fmt.Sprintf("yaml://%s", configPath)
}

func getSourceURL() string {
	return fmt.Sprintf("file://%s", migrationsPath)
}

func readConfigFile(path string) (map[string]interface{}, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, DefaultPerm)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fileData, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	fileMap := map[string]interface{}{}
	if err = yaml.Unmarshal(fileData, &fileMap); err != nil {
		return nil, err
	}

	return fileMap, nil
}

func convertMapToJsonString(m map[string]interface{}) (string, error) {
	res, err := json.Marshal(m)
	if err != nil {
		return "", err
	}

	return string(res), nil
}

func readConfigFileAndConvert(path string) (string, error) {
	result, err := readConfigFile(path)
	if err != nil {
		return "", err
	}

	return convertMapToJsonString(result)
}

func TestNew(t *testing.T) {
	_, err := New(nil)
	if err == nil {
		t.Fatal("expected an error when calling New with empty config")
	}

	_, err = New(&Config{Path: ""})
	if err == nil {
		t.Fatal("expected an error when calling New with empty path")
	}

	_, err = New(&Config{Path: "1http://foo.com"})
	if err == nil {
		t.Fatal("expected an error when calling New with invalid path")
	}

	y, err := New(&Config{Path: getSourceURL(), Perm: 0777})
	if err != nil {
		t.Fatal(err)
	}

	path, err := url.ParseURL(getSourceURL())
	if err != nil {
		t.Fatal(err)
	}

	if y.config.Path != path {
		t.Fatal("path does not match")
	}

	if y.config.Perm != 0777 {
		t.Fatal("perm does not match")
	}

	y, err = New(&Config{Path: getSourceURL()})
	if err != nil {
		t.Fatal(err)
	}

	if y.config.Path != path {
		t.Fatal("path does not match")
	}

	if y.config.Perm != DefaultPerm {
		t.Fatal("perm does not match")
	}
}

func TestOpen(t *testing.T) {
	y := Yaml{}

	_, err := y.Open("1http://foo.com")
	if err == nil {
		t.Fatal("expected an error when calling New with invalid path")
	}

	_, err = y.Open(getSourceURL())
	if err != nil {
		t.Fatal(err)
	}
}

func TestLock_Unlock(t *testing.T) {
	defer os.Remove(configPath)

	y, err := New(&Config{Path: getConfigURL()})
	if err != nil {
		t.Fatal(err)
	}

	y.config.Path = "1http://foo.com"
	if err := y.Lock(); err == nil {
		t.Fatal("expected an error when calling Lock with invalid path")
	}

	y, err = New(&Config{Path: getConfigURL()})
	if err != nil {
		t.Fatal(err)
	}

	if err := y.Lock(); err != nil {
		t.Fatal(err)
	}

	if y.lockedFile == nil {
		t.Fatal("expected a locked file")
	}

	if err := y.Unlock(); err != nil {
		t.Fatal(err)
	}

	_, err = readConfigFile(y.config.Path)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLock_Close(t *testing.T) {
	defer os.Remove(configPath)

	y, err := New(&Config{Path: getConfigURL()})
	if err != nil {
		t.Fatal(err)
	}

	if err := y.Close(); err != nil {
		t.Fatal(err)
	}

	if err := y.Lock(); err != nil {
		t.Fatal(err)
	}
	defer y.mu.Unlock()

	if y.lockedFile == nil {
		t.Fatal("expected a locked file")
	}

	if err := y.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestUp1(t *testing.T) {
	defer os.Remove(configPath)

	m, err := migrate.New(getSourceURL(), getConfigURL())
	if err != nil {
		t.Error(err)
		return
	}

	if err := m.Steps(1); err != nil {
		t.Error(err)
		return
	}

	result, err := readConfigFileAndConvert(configPath)
	if err != nil {
		t.Error(err)
		return
	}

	expected, err := convertMapToJsonString(map[string]interface{}{
		"version": 1,
		"force":   false,
		"str":     "str",
		"number":  1,
		"boolean": true,
	})
	if err != nil {
		t.Error(err)
		return
	}

	if result != expected {
		t.Errorf("Expected: %s, got: %s", expected, result)
	}
}

func TestUp2(t *testing.T) {
	defer os.Remove(configPath)

	m, err := migrate.New(getSourceURL(), getConfigURL())
	if err != nil {
		t.Error(err)
		return
	}

	if err := m.Steps(2); err != nil {
		t.Error(err)
		return
	}

	result, err := readConfigFileAndConvert(configPath)
	if err != nil {
		t.Error(err)
		return
	}

	expected, err := convertMapToJsonString(map[string]interface{}{
		"version": 2,
		"force":   false,
		"str":     "str",
		"number":  1,
		"boolean": true,
		"map": map[string]interface{}{
			"map_str":     "map_str",
			"map_number":  2,
			"map_boolean": false,
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	if result != expected {
		t.Errorf("Expected: %s, got: %s", expected, result)
	}
}

func TestUp3(t *testing.T) {
	defer os.Remove(configPath)

	m, err := migrate.New(getSourceURL(), getConfigURL())
	if err != nil {
		t.Error(err)
		return
	}

	if err := m.Steps(3); err != nil {
		t.Error(err)
		return
	}

	result, err := readConfigFileAndConvert(configPath)
	if err != nil {
		t.Error(err)
		return
	}

	expected, err := convertMapToJsonString(map[string]interface{}{
		"version": 3,
		"force":   false,
		"array": []map[string]interface{}{
			{
				"map_array_boolean": []bool{true, false, true},
				"map_array_number":  []int{1, 2, 3},
				"map_array_str":     []string{"str1", "str2", "str3"},
				"map_boolean":       false,
				"map_number":        2,
				"map_str":           "map_str",
			},
			{
				"map_array_boolean": []bool{false, true, false},
				"map_array_number":  []int{4, 5, 6},
				"map_array_str":     []string{"str4", "str5", "str6"},
				"map_boolean":       false,
				"map_number":        2,
				"map_str":           "map_str2",
			},
		},
		"array2": []int{1, 2, 3},
		"array3": []string{"str1", "str2", "str3"},
		"array4": []bool{true, false, true},
	})
	if err != nil {
		t.Error(err)
		return
	}

	if result != expected {
		t.Errorf("Expected: %s, got: %s", expected, result)
	}
}

func TestUp3_Invalid_Config_File(t *testing.T) {
	defer os.Remove(configPath)

	m, err := migrate.New(getSourceURL(), getConfigURL())
	if err != nil {
		t.Error(err)
		return
	}

	if err := m.Steps(2); err != nil {
		t.Error(err)
		return
	}

	err = func() error {
		f, err := os.OpenFile(configPath, os.O_WRONLY, DefaultPerm)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := f.Write([]byte("\ninvalid string\n")); err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		t.Error(err)
		return
	}

	if err := m.Steps(1); err == nil {
		t.Error("expected error")
		return
	}
}

func TestUp4_Invalid_Migration_File(t *testing.T) {
	defer os.Remove(configPath)

	m, err := migrate.New(getSourceURL(), getConfigURL())
	if err != nil {
		t.Error(err)
		return
	}

	if err := m.Steps(4); err == nil {
		t.Error("expected error")
		return
	}

	result, err := readConfigFileAndConvert(configPath)
	if err != nil {
		t.Error(err)
		return
	}

	expected, err := convertMapToJsonString(map[string]interface{}{
		"version": 4,
		"force":   true,
		"array": []map[string]interface{}{
			{
				"map_array_boolean": []bool{true, false, true},
				"map_array_number":  []int{1, 2, 3},
				"map_array_str":     []string{"str1", "str2", "str3"},
				"map_boolean":       false,
				"map_number":        2,
				"map_str":           "map_str",
			},
			{
				"map_array_boolean": []bool{false, true, false},
				"map_array_number":  []int{4, 5, 6},
				"map_array_str":     []string{"str4", "str5", "str6"},
				"map_boolean":       false,
				"map_number":        2,
				"map_str":           "map_str2",
			},
		},
		"array2": []int{1, 2, 3},
		"array3": []string{"str1", "str2", "str3"},
		"array4": []bool{true, false, true},
	})
	if err != nil {
		t.Error(err)
		return
	}

	if result != expected {
		t.Errorf("Expected: %s, got: %s", expected, result)
	}
}
