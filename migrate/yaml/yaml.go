package yaml

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/c2pc/golang-file-migrate/internal/merge"
	"github.com/c2pc/golang-file-migrate/internal/url"
	migration "github.com/golang-migrate/migrate/v4/database"
	"github.com/pkg/errors"
	lockedFile "github.com/rogpeppe/go-internal/lockedfile"
	"gopkg.in/yaml.v3"
)

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

	yml := &Yaml{
		config: &Config{
			Path: path,
		},
	}

	return yml, nil
}

// Open returns a new driver instance configured with parameters
// coming from the URL string. Migrate will call this function
// only once per instance.
func (y *Yaml) Open(filePath string) (migration.Driver, error) {
	js, err := New(&Config{Path: filePath})
	if err != nil {
		return nil, err
	}

	return js, nil
}

// Close closes the underlying database instance managed by the driver.
// Migrate will call this function only once per instance.
func (y *Yaml) Close() error {
	return y.lockedFile.Close()
}

// Lock should acquire a database lock so that only one migration process
// can run at a time. Migrate will call this function before Run is called.
// If the implementation can't provide this functionality, return nil.
// Return database.ErrLocked if database is already locked.
func (y *Yaml) Lock() error {
	f, err := lockedFile.OpenFile(y.config.Path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	y.mu.Lock()

	y.lockedFile = f

	return nil
}

// Unlock should release the lock. Migrate will call this function after
// all migrations have been run.
func (y *Yaml) Unlock() error {
	y.mu.Unlock()
	return y.Close()
}

// Run applies a migration to the database. migration is guaranteed to be not nil.
func (y *Yaml) Run(migration io.Reader) error {
	migrData, err := io.ReadAll(migration)
	if err != nil {
		return err
	}

	migrMap := map[string]interface{}{}
	if err := yaml.Unmarshal(migrData, &migrMap); err != nil {
		return errors.Wrapf(err, "failed to parse migration file")
	}

	if _, err = y.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	fileData, err := io.ReadAll(y.lockedFile)
	if err != nil {
		return err
	}

	fileMap := map[string]interface{}{}
	if err := yaml.Unmarshal(fileData, &fileMap); err != nil {
		return errors.Wrapf(err, "failed to parse %s", y.config.Path)
	}

	base := map[string]interface{}{}
	base = merge.Merge(migrMap, fileMap)

	delete(base, "version")
	delete(base, "force")

	data, err := yaml.Marshal(base)
	if err != nil {
		return err
	}
	newData := strings.ReplaceAll(string(data), "'", "")
	newData = strings.ReplaceAll(newData, "null", "")

	err = y.lockedFile.Truncate(0)
	if err != nil {
		return err
	}

	if _, err = y.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	_, err = y.lockedFile.Write([]byte(newData))
	if err != nil {
		return err
	}

	return nil
}

// SetVersion saves version and dirty state.
// Migrate will call this function before and after each call to Run.
// version must be >= -1. -1 means NilVersion.
func (y *Yaml) SetVersion(version int, dirty bool) error {
	if _, err := y.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	fileData, err := io.ReadAll(y.lockedFile)
	if err != nil {
		return err
	}

	fileMap := map[string]interface{}{}
	if err := yaml.Unmarshal(fileData, &fileMap); err != nil {
		return errors.Wrapf(err, "failed to parse %s", y.config.Path)
	}

	delete(fileMap, "version")
	delete(fileMap, "force")

	data, err := yaml.Marshal(fileMap)
	if err != nil {
		return err
	}

	newData := strings.ReplaceAll(string(data), "null", "")

	if len(fileMap) == 0 {
		newData = ""
	}

	newData = fmt.Sprintf("version: %v\nforce: %v\n", version, dirty) + newData

	err = y.lockedFile.Truncate(0)
	if err != nil {
		return err
	}

	if _, err = y.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	_, err = y.lockedFile.Write([]byte(newData))
	if err != nil {
		return err
	}

	return nil
}

// Version returns the currently active version and if the database is dirty.
// When no migration has been applied, it must return version -1.
// Dirty means, a previous migration failed and user interaction is required.
func (y *Yaml) Version() (int, bool, error) {
	if _, err := y.lockedFile.Seek(0, 0); err != nil {
		return 0, false, err
	}

	r, err := io.ReadAll(y.lockedFile)
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

// Drop deletes everything in the database.
// Note that this is a breaking action, a new call to Open() is necessary to
// ensure subsequent calls work as expected.
func (y *Yaml) Drop() error {
	err := y.lockedFile.Truncate(0)
	if err != nil {
		return err
	}

	if _, err = y.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	return nil
}
