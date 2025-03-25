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

// Config represents the core struct used to manage config-based migrations.
// It contains a driver for reading/writing config data and a locked file to prevent concurrent access.
type Config struct {
	driver     Driver           // Custom config driver implementing (Un)Marshal and Version logic
	lockedFile *lockedFile.File // File handle with locking to avoid race conditions
	mu         sync.Mutex       // Mutex to synchronize file access
	path       string           // Path to the configuration file
	perm       fs.FileMode      // File permissions
}

// New returns a new instance of the config driver using the given settings.
func New(driver Driver, cfg Settings) database.Driver {
	path, err := url.ParseURL(cfg.Path)
	if err != nil {
		panic(err)
	}

	perm := DefaultPerm
	if cfg.Perm != 0 {
		perm = cfg.Perm
	}

	yml := &Config{
		driver: driver,
		path:   path,
		perm:   perm,
	}

	return yml
}

// Open sets the file path from a URL and returns the current instance.
func (m *Config) Open(filePath string) (database.Driver, error) {
	path, err := url.ParseURL(filePath)
	if err != nil {
		return nil, err
	}

	m.path = path
	return m, nil
}

// Close closes the locked file if open.
func (m *Config) Close() error {
	if err := m.lockedFile.Close(); err != nil {
		return err
	}

	return nil
}

// Lock opens the file with locking and stores the handle for later operations.
func (m *Config) Lock() error {
	f, err := lockedFile.OpenFile(m.path, os.O_RDWR|os.O_CREATE, m.perm)
	if err != nil {
		return err
	}
	m.mu.Lock()

	m.lockedFile = f

	return nil
}

// Unlock releases the mutex and closes the locked file.
func (m *Config) Unlock() error {
	m.mu.Unlock()
	return m.Close()
}

// Run applies a migration by reading the migration data and merging it with existing config data.
func (m *Config) Run(migration io.Reader) error {
	// Read migration file content
	migrData, err := io.ReadAll(migration)
	if err != nil {
		return err
	}

	// Unmarshal migration data into a map
	migrMap := map[string]interface{}{}
	if err := m.driver.Unmarshal(migrData, &migrMap); err != nil {
		return errors.Wrapf(err, "failed to parse migration file")
	}

	// Reset file cursor and read existing file content
	if _, err = m.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	fileData, err := io.ReadAll(m.lockedFile)
	if err != nil {
		return err
	}

	// Unmarshal current file into map
	fileMap := map[string]interface{}{}
	if err := m.driver.Unmarshal(fileData, &fileMap); err != nil {
		return errors.Wrapf(err, "failed to parse %s", m.path)
	}

	// Merge current config and migration changes
	base := merger.Merge(migrMap, fileMap)

	// Remove migration-specific metadata
	delete(base, "version")
	delete(base, "force")

	// Marshal merged data to bytes
	data, err := m.driver.Marshal(base)
	if err != nil {
		return err
	}

	// Clean up unwanted values in output
	newData := strings.ReplaceAll(string(data), "'", "")
	newData = strings.ReplaceAll(newData, "null", "")

	// Truncate and overwrite the file with new content
	err = m.lockedFile.Truncate(0)
	if err != nil {
		return err
	}

	if _, err = m.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	_, err = m.lockedFile.Write([]byte(newData))
	return err
}

// SetVersion updates the current config file with version and dirty (force) flags.
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

	// Set versioning fields
	fileMap["version"] = version
	fileMap["force"] = dirty

	data, err := m.driver.Marshal(fileMap)
	if err != nil {
		return err
	}

	newData := strings.ReplaceAll(string(data), "null", "")

	// Truncate and overwrite with updated version
	err = m.lockedFile.Truncate(0)
	if err != nil {
		return err
	}

	if _, err = m.lockedFile.Seek(0, 0); err != nil {
		return err
	}

	_, err = m.lockedFile.Write([]byte(newData))
	return err
}

// Version reads and returns the current migration version and dirty flag.
func (m *Config) Version() (int, bool, error) {
	if _, err := m.lockedFile.Seek(0, 0); err != nil {
		if errors.Is(err, fs.ErrClosed) {
			// If file is closed, reopen and lock it
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

// Drop resets the config file by truncating it and writing empty/default content.
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
