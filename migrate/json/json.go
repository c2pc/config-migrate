package json

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/c2pc/golang-file-migrate/internal/merge"
	"github.com/c2pc/golang-file-migrate/internal/url"
	migration "github.com/golang-migrate/migrate/v4/database"
	"github.com/pkg/errors"
	lockedFile "github.com/rogpeppe/go-internal/lockedfile"
)

const emptyFileTemplate = `{
    "version": %v,
    "force": %v
    `

const nonEmptyFileTemplate = `{
    "version": %v,
    "force": %v,`

type version struct {
	Version int  `json:"version"`
	Force   bool `json:"force"`
}

func init() {
	j := Json{}
	migration.Register("json", &j)
}

type Config struct {
	Path string
}

type Json struct {
	lockedFile *lockedFile.File
	mu         sync.Mutex
	config     *Config
}

func New(config *Config) (*Json, error) {
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

	js := &Json{
		config: &Config{
			Path: path,
		},
	}

	return js, nil
}

// Open returns a new driver instance configured with parameters
// coming from the URL string. Migrate will call this function
// only once per instance.
func (j *Json) Open(filePath string) (migration.Driver, error) {
	js, err := New(&Config{Path: filePath})
	if err != nil {
		return nil, err
	}

	return js, nil
}

// Close closes the underlying database instance managed by the driver.
// Migrate will call this function only once per instance.
func (j *Json) Close() error {
	return j.lockedFile.Close()
}

// Lock should acquire a database lock so that only one migration process
// can run at a time. Migrate will call this function before Run is called.
// If the implementation can't provide this functionality, return nil.
// Return database.ErrLocked if database is already locked.
func (j *Json) Lock() error {
	f, err := lockedFile.OpenFile(j.config.Path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	j.mu.Lock()

	j.lockedFile = f

	return nil
}

// Unlock should release the lock. Migrate will call this function after
// all migrations have been run.
func (j *Json) Unlock() error {
	j.mu.Unlock()
	return j.Close()
}

// Run applies a migration to the database. migration is guaranteed to be not nil.
func (j *Json) Run(migration io.Reader) error {
	migrData, err := io.ReadAll(migration)
	if err != nil {
		return err
	}

	migrMap := map[string]interface{}{}
	if err := json.Unmarshal(migrData, &migrMap); err != nil {
		return errors.Wrapf(err, "failed to parse migration file")
	}

	if _, err = j.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	fileData, err := io.ReadAll(j.lockedFile)
	if err != nil {
		return err
	}

	if len(fileData) == 0 {
		fileData = []byte("{}")
	}

	fileMap := map[string]interface{}{}
	if err := json.Unmarshal(fileData, &fileMap); err != nil {
		return errors.Wrapf(err, "failed to parse %s", j.config.Path)
	}

	base := map[string]interface{}{}
	base = merge.Merge(migrMap, fileMap)

	delete(base, "version")
	delete(base, "force")

	data, err := json.MarshalIndent(base, "", "    ")
	if err != nil {
		return err
	}

	err = j.lockedFile.Truncate(0)
	if err != nil {
		return err
	}

	if _, err = j.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	_, err = j.lockedFile.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// SetVersion saves version and dirty state.
// Migrate will call this function before and after each call to Run.
// version must be >= -1. -1 means NilVersion.
func (j *Json) SetVersion(version int, dirty bool) error {
	if _, err := j.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	fileData, err := io.ReadAll(j.lockedFile)
	if err != nil {
		return err
	}

	if len(fileData) == 0 {
		fileData = []byte("{}")
	}

	fileMap := map[string]interface{}{}
	if err := json.Unmarshal(fileData, &fileMap); err != nil {
		return errors.Wrapf(err, "failed to parse %s", j.config.Path)
	}

	if version >= 0 || (version == migration.NilVersion && dirty) {
		delete(fileMap, "version")
		delete(fileMap, "force")

		data, err := json.MarshalIndent(fileMap, "", "    ")
		if err != nil {
			return err
		}

		newData := string(data)
		if len(fileMap) == 0 {
			newData = fmt.Sprintf(emptyFileTemplate, version, dirty) + newData[1:]
		} else {
			newData = fmt.Sprintf(nonEmptyFileTemplate, version, dirty) + newData[1:]
		}

		err = j.lockedFile.Truncate(0)
		if err != nil {
			return err
		}

		if _, err = j.lockedFile.Seek(0, 0); err != nil {
			return err
		}

		_, err = j.lockedFile.Write([]byte(newData))
		if err != nil {
			return err
		}
	}

	return nil
}

// Version returns the currently active version and if the database is dirty.
// When no migration has been applied, it must return version -1.
// Dirty means, a previous migration failed and user interaction is required.
func (j *Json) Version() (int, bool, error) {
	if _, err := j.lockedFile.Seek(0, 0); err != nil {
		return 0, false, err
	}

	r, err := io.ReadAll(j.lockedFile)
	if err != nil {
		return 0, false, err
	}

	if len(r) == 0 {
		return migration.NilVersion, false, nil
	}

	v := new(version)
	if err := json.Unmarshal(r, v); err != nil {
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
func (j *Json) Drop() error {
	err := j.lockedFile.Truncate(0)
	if err != nil {
		return err
	}

	if _, err = j.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	return nil
}
