package migrator

import (
	"io"
	"io/fs"
	"os"
	"strings"
	"sync"

	"github.com/c2pc/config-migrate/internal/merger"
	"github.com/c2pc/config-migrate/internal/url"
	migration "github.com/golang-migrate/migrate/v4/database"
	"github.com/pkg/errors"
	lockedFile "github.com/rogpeppe/go-internal/lockedfile"
)

const DefaultPerm fs.FileMode = 0666

type Config struct {
	Path string
	Perm fs.FileMode
}

type Migrator interface {
	Unmarshal([]byte, interface{}) error
	Marshal(interface{}) ([]byte, error)
	Version([]byte) (int, bool, error)
	EmptyData() []byte
}

func New(migrator Migrator, config Config) *Migrate {
	perm := DefaultPerm
	if config.Perm != 0 {
		perm = config.Perm
	}

	yml := &Migrate{
		migrator: migrator,
		path:     config.Path,
		perm:     perm,
	}

	return yml
}

type Migrate struct {
	migrator   Migrator
	lockedFile *lockedFile.File
	mu         sync.Mutex
	path       string
	perm       fs.FileMode
}

func (m *Migrate) Open(filePath string) (migration.Driver, error) {
	path, err := url.ParseURL(filePath)
	if err != nil {
		return nil, err
	}

	m.path = path
	return m, nil
}

func (m *Migrate) Close() error {
	if err := m.lockedFile.Close(); err != nil {
		return err
	}

	return nil
}

func (m *Migrate) Lock() error {
	f, err := lockedFile.OpenFile(m.path, os.O_RDWR|os.O_CREATE, m.perm)
	if err != nil {
		return err
	}
	m.mu.Lock()

	m.lockedFile = f

	return nil
}

func (m *Migrate) Unlock() error {
	m.mu.Unlock()
	return m.Close()
}

func (m *Migrate) Run(migration io.Reader) error {
	migrData, err := io.ReadAll(migration)
	if err != nil {
		return err
	}

	migrMap := map[string]interface{}{}
	if err := m.migrator.Unmarshal(migrData, &migrMap); err != nil {
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
	if err := m.migrator.Unmarshal(fileData, &fileMap); err != nil {
		return errors.Wrapf(err, "failed to parse %s", m.path)
	}

	base := map[string]interface{}{}
	base = merger.Merge(migrMap, fileMap)

	delete(base, "version")
	delete(base, "force")

	data, err := m.migrator.Marshal(base)
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

func (m *Migrate) SetVersion(version int, dirty bool) error {
	if _, err := m.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	fileData, err := io.ReadAll(m.lockedFile)
	if err != nil {
		return err
	}

	fileMap := map[string]interface{}{}
	if err := m.migrator.Unmarshal(fileData, &fileMap); err != nil {
		return errors.Wrapf(err, "failed to parse %s", m.path)
	}

	fileMap["version"] = version
	fileMap["force"] = dirty

	data, err := m.migrator.Marshal(fileMap)
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

func (m *Migrate) Version() (int, bool, error) {
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
		return migration.NilVersion, false, nil
	}

	version, force, err := m.migrator.Version(r)
	if err != nil {
		return 0, false, err
	}

	if version == 0 {
		return migration.NilVersion, false, nil
	}

	return version, force, nil
}

func (m *Migrate) Drop() error {
	err := m.lockedFile.Truncate(0)
	if err != nil {
		return err
	}

	if _, err = m.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	emptyData := m.migrator.EmptyData()
	if emptyData != nil && len(emptyData) > 0 {
		_, err = m.lockedFile.Write(emptyData)
		if err != nil {
			return err
		}
	}

	return nil
}
