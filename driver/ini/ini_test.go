package ini

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	config "github.com/c2pc/config-migrate/driver"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const configPath = "./examples/config.ini"
const migrationsPath = "./examples/migrations"

func getConfig() database.Driver {
	return New(config.Settings{
		Path:                    configPath,
		Perm:                    0777,
		UnableToReplaceComments: false,
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
		fileData = (Ini{}).EmptyData()
	}

	fileMap := map[string]interface{}{}
	if err := (Ini{}).Unmarshal(fileData, &fileMap); err != nil {
		return nil, err
	}

	return fileMap, nil
}

func convertMapToIniString(m map[string]interface{}) (string, error) {
	b, err := (Ini{}).Marshal(m, false)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func readConfigFileAndConvert(path string) (string, error) {
	result, err := readConfigFile(path)
	if err != nil {
		return "", err
	}
	return convertMapToIniString(result)
}

func TestNew(t *testing.T) {
	var _ database.Driver
	_ = New(config.Settings{})
}

func TestOpen(t *testing.T) {
	iniDriver := New(config.Settings{})

	_, err := iniDriver.Open("1http://foo.com")
	if err == nil {
		t.Fatal("expected an error when calling Open with invalid path")
	}

	_, err = iniDriver.Open(getSourceURL())
	if err != nil {
		t.Fatal(err)
	}
}

func TestLock_Unlock(t *testing.T) {
	defer os.Remove(configPath)

	iniDriver := New(config.Settings{})

	_, err := iniDriver.Open("ini://" + configPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := iniDriver.Lock(); err != nil {
		t.Fatal(err)
	}

	if err := iniDriver.Unlock(); err != nil {
		t.Fatal(err)
	}

	_, err = readConfigFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLock_Close(t *testing.T) {
	defer os.Remove(configPath)

	iniDriver := New(config.Settings{Path: configPath})

	if err := iniDriver.Lock(); err != nil {
		t.Fatal(err)
	}

	if err := iniDriver.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestUp1(t *testing.T) {
	defer os.Remove(configPath)

	m, err := migrate.NewWithDatabaseInstance(getSourceURL(), "ini", getConfig())
	if err != nil {
		t.Error(err)
		return
	}

	if err := m.Steps(1); err != nil {
		t.Error(err)
		return
	}

	result, err := readConfigFile(configPath)
	if err != nil {
		t.Error(err)
		return
	}

	expected := map[string]interface{}{
		"": map[string]interface{}{
			"version": "1", "force": "false", "str": "str", "number": "1", "boolean": "true",
		},
		"version": "1", "force": "false", "str": "str", "number": "1", "boolean": "true",
		"host": map[string]interface{}{
			"url": "url", "host": "host10",
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected: %v, got: %v", expected, result)
	}

	v, f, err := m.Version()
	if err != nil {
		t.Error(err)
		return
	}

	if v != 1 {
		t.Errorf("Expected version %d, got: %d", 1, v)
	}

	if f != false {
		t.Errorf("Expected force %t, got: %t", false, f)
	}
}

func TestUp2(t *testing.T) {
	defer os.Remove(configPath)

	m, err := migrate.NewWithDatabaseInstance(getSourceURL(), "ini", getConfig())
	if err != nil {
		t.Error(err)
		return
	}

	if err := m.Steps(2); err != nil {
		t.Error(err)
		return
	}

	result, err := readConfigFile(configPath)
	if err != nil {
		t.Error(err)
		return
	}

	expected := map[string]interface{}{
		"": map[string]interface{}{
			"version": "2", "force": "false", "str": "str", "number": "1", "boolean": "true",
		},
		"version": "2", "force": "false", "str": "str", "number": "1", "boolean": "true",
		"map": map[string]interface{}{
			"map_str": "map_str", "map_number": "2", "map_boolean": "false",
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected: %v, got: %v", expected, result)
	}

	v, f, err := m.Version()
	if err != nil {
		t.Error(err)
		return
	}

	if v != 2 {
		t.Errorf("Expected version %d, got: %d", 2, v)
	}

	if f != false {
		t.Errorf("Expected force %t, got: %t", false, f)
	}
}

func TestUp3(t *testing.T) {
	defer os.Remove(configPath)

	m, err := migrate.NewWithDatabaseInstance(getSourceURL(), "ini", getConfig())
	if err != nil {
		t.Error(err)
		return
	}

	if err := m.Steps(3); err != nil {
		t.Error(err)
		return
	}

	result, err := readConfigFile(configPath)
	if err != nil {
		t.Error(err)
		return
	}

	expected := map[string]interface{}{
		"": map[string]interface{}{
			"version": "3", "force": "false", "str": "str", "number": "1", "boolean": "true",
		},
		"version": "3", "force": "false", "str": "str", "number": "1", "boolean": "true",
		"map": map[string]interface{}{
			"map_str": "map_str", "map_number": "2", "map_boolean": "false",
		},
		"app": map[string]interface{}{
			"name": "config", "debug": "true",
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected: %v, got: %v", expected, result)
	}

	v, f, err := m.Version()
	if err != nil {
		t.Error(err)
		return
	}

	if v != 3 {
		t.Errorf("Expected version %d, got: %d", 3, v)
	}

	if f != false {
		t.Errorf("Expected force %t, got: %t", false, f)
	}
}

func TestUp3_Invalid_Config_File(t *testing.T) {
	defer os.Remove(configPath)

	m, err := migrate.NewWithDatabaseInstance(getSourceURL(), "ini", getConfig())
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

	m, err := migrate.NewWithDatabaseInstance(getSourceURL(), "ini", getConfig())
	if err != nil {
		t.Error(err)
		return
	}

	if err := m.Steps(4); err == nil {
		t.Error("expected error")
		return
	}

	result, err := readConfigFile(configPath)
	if err != nil {
		t.Error(err)
		return
	}

	expected := map[string]interface{}{
		"": map[string]interface{}{
			"version": "4", "force": "true", "str": "str", "number": "1", "boolean": "true",
		},
		"version": "4", "force": "true", "str": "str", "number": "1", "boolean": "true",
		"map": map[string]interface{}{
			"map_str": "map_str", "map_number": "2", "map_boolean": "false",
		},
		"app": map[string]interface{}{
			"name": "config", "debug": "true",
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected: %v, got: %v", expected, result)
	}

	v, f, err := m.Version()
	if err != nil {
		t.Error(err)
		return
	}

	if v != 4 {
		t.Errorf("Expected version %d, got: %d", 4, v)
	}

	if f != true {
		t.Errorf("Expected force %t, got: %t", true, f)
	}
}

func TestDrop(t *testing.T) {
	defer os.Remove(configPath)

	m, err := migrate.NewWithDatabaseInstance(getSourceURL(), "ini", getConfig())
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

	result, err := readConfigFile(configPath)
	if err != nil {
		t.Error(err)
		return
	}

	// After Drop, driver writes EmptyData() = version=0, force=false
	expectedMap := map[string]interface{}{}
	_ = (Ini{}).Unmarshal((Ini{}).EmptyData(), &expectedMap)
	if !reflect.DeepEqual(result, expectedMap) {
		t.Errorf("Expected after Drop: %v, got: %v", expectedMap, result)
	}
}
