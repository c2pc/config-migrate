package yaml

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/c2pc/config-migrate/driver"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gopkg.in/yaml.v3"
)

const configPath = "./examples/config.yaml"
const migrationsPath = "./examples/migrations"

func getConfig() database.Driver {
	return New(config.Settings{
		Path:                    configPath,
		Perm:                    0777,
		UnableToReplaceComments: true,
	})
}

func getSourceURL() string {
	return fmt.Sprintf("file://%s", migrationsPath)
}

func readConfigFile(path string) (map[string]interface{}, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, config.DefaultPerm)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fileData, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	if len(fileData) == 0 {
		fileData = []byte("{}")
	}

	fileMap := map[string]interface{}{}
	if err = yaml.Unmarshal(fileData, &fileMap); err != nil {
		return nil, err
	}

	return fileMap, nil
}

func convertMapToJsonString(m map[string]interface{}) (string, error) {
	res, err := yaml.Marshal(m)
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
	var _ database.Driver
	_ = New(config.Settings{})
}

func TestOpen(t *testing.T) {
	y := New(config.Settings{})

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

	y := New(config.Settings{})

	_, err := y.Open("yaml://" + configPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := y.Lock(); err != nil {
		t.Fatal(err)
	}

	if err := y.Unlock(); err != nil {
		t.Fatal(err)
	}

	_, err = readConfigFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLock_Close(t *testing.T) {
	defer os.Remove(configPath)

	y := New(config.Settings{Path: configPath})

	if err := y.Lock(); err != nil {
		t.Fatal(err)
	}

	if err := y.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestUp1(t *testing.T) {
	defer os.Remove(configPath)

	m, err := migrate.NewWithDatabaseInstance(getSourceURL(), "yaml", getConfig())
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

	v, f, err := m.Version()
	if err != nil {
		t.Error(err)
		return
	}

	if v != 1 {
		t.Errorf("Expected: %d, got: %d", 1, v)
	}

	if f != false {
		t.Errorf("Expected: %t, got: %t", false, f)
	}
}

func TestUp2(t *testing.T) {
	defer os.Remove(configPath)

	m, err := migrate.NewWithDatabaseInstance(getSourceURL(), "yaml", getConfig())
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

	v, f, err := m.Version()
	if err != nil {
		t.Error(err)
		return
	}

	if v != 2 {
		t.Errorf("Expected: %d, got: %d", 2, v)
	}

	if f != false {
		t.Errorf("Expected: %t, got: %t", false, f)
	}
}

func TestUp3(t *testing.T) {
	defer os.Remove(configPath)

	m, err := migrate.NewWithDatabaseInstance(getSourceURL(), "yaml", getConfig())
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
		t.Errorf("Expected:\n %s, got:\n %s", expected, result)
	}

	v, f, err := m.Version()
	if err != nil {
		t.Error(err)
		return
	}

	if v != 3 {
		t.Errorf("Expected: %d, got: %d", 3, v)
	}

	if f != false {
		t.Errorf("Expected: %t, got: %t", false, f)
	}
}

func TestUp3_Comments(t *testing.T) {
	defer os.Remove(configPath)

	m, err := migrate.NewWithDatabaseInstance(getSourceURL(), "yaml", getConfig())
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
		t.Errorf("Expected:\n %s, got:\n %s", expected, result)
	}

	v, f, err := m.Version()
	if err != nil {
		t.Error(err)
		return
	}

	if v != 3 {
		t.Errorf("Expected: %d, got: %d", 3, v)
	}

	if f != false {
		t.Errorf("Expected: %t, got: %t", false, f)
	}
}

func TestUp3_Invalid_Config_File(t *testing.T) {
	defer os.Remove(configPath)

	m, err := migrate.NewWithDatabaseInstance(getSourceURL(), "yaml", getConfig())
	if err != nil {
		t.Error(err)
		return
	}

	if err := m.Steps(2); err != nil {
		t.Error(err)
		return
	}

	err = func() error {
		f, err := os.OpenFile(configPath, os.O_WRONLY, config.DefaultPerm)
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

	m, err := migrate.NewWithDatabaseInstance(getSourceURL(), "yaml", getConfig())
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

	v, f, err := m.Version()
	if err != nil {
		t.Error(err)
		return
	}

	if v != 4 {
		t.Errorf("Expected: %d, got: %d", 4, v)
	}

	if f != true {
		t.Errorf("Expected: %t, got: %t", true, f)
	}
}

func TestDrop(t *testing.T) {
	defer os.Remove(configPath)

	m, err := migrate.NewWithDatabaseInstance(getSourceURL(), "yaml", getConfig())
	if err != nil {
		t.Error(err)
		return
	}

	if err := m.Steps(2); err != nil {
		t.Error(err)
		return
	}

	if err := m.Drop(); err != nil {
		t.Error(err)
		return
	}

	result, err := readConfigFileAndConvert(configPath)
	if err != nil {
		t.Error(err)
		return
	}

	expected, err := convertMapToJsonString(map[string]interface{}{})
	if err != nil {
		t.Error(err)
		return
	}

	if result != expected {
		t.Errorf("Expected: %s, got: %s", expected, result)
	}
}
