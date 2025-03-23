package yaml

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
	"gopkg.in/yaml.v3"
)

const DefaultPerm fs.FileMode = 0666

type version struct {
	Version int  `yaml:"version"`
	Force   bool `json:"force"`
}

func init() {
	y := Yaml{}
	migration.Register("yaml", &y)
}

type Config struct {
	Path string
	Perm fs.FileMode
}

type Yaml struct {
	lockedFile *lockedFile.File
	mu         sync.Mutex
	config     *Config
}

func New(config *Config) (*Yaml, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.Path == "" {
		return nil, errors.New("config path is empty")
	}

	path, err := url.ParseURL(config.Path)
	if err != nil {
		return nil, err
	}

	perm := DefaultPerm
	if config.Perm != 0 {
		perm = config.Perm
	}

	yml := &Yaml{
		config: &Config{
			Path: path,
			Perm: perm,
		},
	}

	return yml, nil
}

// Open returns a new driver instance configured with parameters
// coming from the URL string. Migrate will call this function
// only once per instance.
func (m *Yaml) Open(filePath string) (migration.Driver, error) {
	js, err := New(&Config{Path: filePath})
	if err != nil {
		return nil, err
	}

	return js, nil
}

// Close closes the underlying file instance managed by the driver.
// Migrate will call this function only once per instance.
func (m *Yaml) Close() error {
	if m.lockedFile != nil {
		return m.lockedFile.Close()
	}

	return nil
}

// Lock should acquire a file lock so that only one migration process
// can run at a time. Migrate will call this function before Run is called.
// If the implementation can't provide this functionality, return nil.
// Return file.ErrLocked if file is already locked.
func (m *Yaml) Lock() error {
	f, err := lockedFile.OpenFile(m.config.Path, os.O_RDWR|os.O_CREATE, m.config.Perm)
	if err != nil {
		return err
	}
	m.mu.Lock()

	m.lockedFile = f

	return nil
}

// Unlock should release the lock. Migrate will call this function after
// all migrations have been run.
func (m *Yaml) Unlock() error {
	m.mu.Unlock()
	return m.Close()
}

// Run applies a migration to the file. migration is guaranteed to be not nil.
func (m *Yaml) Run(migration io.Reader) error {
	migrData, err := io.ReadAll(migration)
	if err != nil {
		return err
	}

	migrMap := map[string]interface{}{}
	if err := yaml.Unmarshal(migrData, &migrMap); err != nil {
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
	if err := yaml.Unmarshal(fileData, &fileMap); err != nil {
		return errors.Wrapf(err, "failed to parse %s", m.config.Path)
	}

	base := map[string]interface{}{}
	base = merger.Merge(migrMap, fileMap)

	delete(base, "version")
	delete(base, "force")

	data, err := yaml.Marshal(base)
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

// SetVersion saves version and dirty state.
// Migrate will call this function before and after each call to Run.
// version must be >= -1. -1 means NilVersion.
func (m *Yaml) SetVersion(version int, dirty bool) error {
	if _, err := m.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	fileData, err := io.ReadAll(m.lockedFile)
	if err != nil {
		return err
	}

	fileMap := map[string]interface{}{}
	if err := yaml.Unmarshal(fileData, &fileMap); err != nil {
		return errors.Wrapf(err, "failed to parse %s", m.config.Path)
	}

	fileMap["version"] = version
	fileMap["force"] = dirty

	data, err := yaml.Marshal(fileMap)
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

// Version returns the currently active version and if the file is dirty.
// When no migration has been applied, it must return version -1.
// Dirty means, a previous migration failed and user interaction is required.
func (m *Yaml) Version() (int, bool, error) {
	if _, err := m.lockedFile.Seek(0, 0); err != nil {
		return 0, false, err
	}

	r, err := io.ReadAll(m.lockedFile)
	if err != nil {
		return 0, false, err
	}

	if len(r) == 0 {
		return migration.NilVersion, false, nil
	}

	v := new(version)
	if err := yaml.Unmarshal(r, v); err != nil {
		return 0, false, err
	}

	if v.Version == 0 {
		return migration.NilVersion, false, nil
	}

	return v.Version, v.Force, nil
}

// Drop deletes everything in the file.
// Note that this is a breaking action, a new call to Open() is necessary to
// ensure subsequent calls work as expected.
func (m *Yaml) Drop() error {
	err := m.lockedFile.Truncate(0)
	if err != nil {
		return err
	}

	if _, err = m.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	return nil
}
