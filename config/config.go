package config

import (
	"io"
	"io/fs"
	"os"
	"strings"
	"sync"

	"github.com/c2pc/config-migrate/internal/url"
	"github.com/c2pc/config-migrate/merger"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/pkg/errors"
	lockedFile "github.com/rogpeppe/go-internal/lockedfile"
)

const DefaultPerm fs.FileMode = 0666

type Settings struct {
	Path string
	Perm fs.FileMode
}

type Driver interface {
	Unmarshal([]byte, interface{}) error
	Marshal(interface{}) ([]byte, error)
	Version([]byte) (int, bool, error)
	EmptyData() []byte
}

type Config struct {
	driver     Driver
	lockedFile *lockedFile.File
	mu         sync.Mutex
	path       string
	perm       fs.FileMode
}

func New(driver Driver, cfg Settings) database.Driver {
	perm := DefaultPerm
	if cfg.Perm != 0 {
		perm = cfg.Perm
	}

	yml := &Config{
		driver: driver,
		path:   cfg.Path,
		perm:   perm,
	}

	return yml
}

func Register(name string, driver Driver, cfg Settings) {
	m := New(driver, cfg)
	database.Register(name, m)
}

func (m *Config) Open(filePath string) (database.Driver, error) {
	path, err := url.ParseURL(filePath)
	if err != nil {
		return nil, err
	}

	m.path = path
	return m, nil
}

func (m *Config) Close() error {
	if err := m.lockedFile.Close(); err != nil {
		return err
	}

	return nil
}

func (m *Config) Lock() error {
	f, err := lockedFile.OpenFile(m.path, os.O_RDWR|os.O_CREATE, m.perm)
	if err != nil {
		return err
	}
	m.mu.Lock()

	m.lockedFile = f

	return nil
}

func (m *Config) Unlock() error {
	m.mu.Unlock()
	return m.Close()
}

func (m *Config) Run(migration io.Reader) error {
	migrData, err := io.ReadAll(migration)
	if err != nil {
		return err
	}

	migrMap := map[string]interface{}{}
	if err := m.driver.Unmarshal(migrData, &migrMap); err != nil {
		return errors.Wrapf(err, "failed to parse migration file")
	}

	if _, err = m.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	fileData, err := io.ReadAll(m.lockedFile)
	if err != nil {
		return err
	}

	fileMap := map[string]interface{}{}
	if err := m.driver.Unmarshal(fileData, &fileMap); err != nil {
		return errors.Wrapf(err, "failed to parse %s", m.path)
	}

	base := map[string]interface{}{}
	base = merger.Merge(migrMap, fileMap)

	delete(base, "version")
	delete(base, "force")

	data, err := m.driver.Marshal(base)
	if err != nil {
		return err
	}
	newData := strings.ReplaceAll(string(data), "'", "")
	newData = strings.ReplaceAll(newData, "null", "")

	err = m.lockedFile.Truncate(0)
	if err != nil {
		return err
	}

	if _, err = m.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	_, err = m.lockedFile.Write([]byte(newData))
	if err != nil {
		return err
	}

	return nil
}

func (m *Config) SetVersion(version int, dirty bool) error {
	if _, err := m.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	fileData, err := io.ReadAll(m.lockedFile)
	if err != nil {
		return err
	}

	fileMap := map[string]interface{}{}
	if err := m.driver.Unmarshal(fileData, &fileMap); err != nil {
		return errors.Wrapf(err, "failed to parse %s", m.path)
	}

	fileMap["version"] = version
	fileMap["force"] = dirty

	data, err := m.driver.Marshal(fileMap)
	if err != nil {
		return err
	}

	newData := strings.ReplaceAll(string(data), "null", "")

	err = m.lockedFile.Truncate(0)
	if err != nil {
		return err
	}

	if _, err = m.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	_, err = m.lockedFile.Write([]byte(newData))
	if err != nil {
		return err
	}

	return nil
}

func (m *Config) Version() (int, bool, error) {
	if _, err := m.lockedFile.Seek(0, 0); err != nil {
		if errors.Is(err, fs.ErrClosed) {
			if err := m.Lock(); err != nil {
				return 0, false, err
			}
			defer m.Unlock()

			if _, err := m.lockedFile.Seek(0, 0); err != nil {
				return 0, false, err
			}
		} else {
			return 0, false, err
		}
	}

	r, err := io.ReadAll(m.lockedFile)
	if err != nil {
		return 0, false, err
	}

	if len(r) == 0 {
		return database.NilVersion, false, nil
	}

	version, force, err := m.driver.Version(r)
	if err != nil {
		return 0, false, err
	}

	if version == 0 {
		return database.NilVersion, false, nil
	}

	return version, force, nil
}

func (m *Config) Drop() error {
	err := m.lockedFile.Truncate(0)
	if err != nil {
		return err
	}

	if _, err = m.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	emptyData := m.driver.EmptyData()
	if emptyData != nil && len(emptyData) > 0 {
		_, err = m.lockedFile.Write(emptyData)
		if err != nil {
			return err
		}
	}

	return nil
}
