package config

import (
	"io/fs"

	"github.com/golang-migrate/migrate/v4/database"
)

// DefaultPerm defines the default file permissions (0666) for created config files.
const DefaultPerm fs.FileMode = 0666

// CommentSuffix defines the special suffix used to identify comment-like keys.
// Keys ending with this suffix will be treated as comments and converted to actual '#'(YAML) comments in output.
const CommentSuffix = "___comment___"

// Settings represents the configuration settings for a config driver.
type Settings struct {
	// Path — the path to the configuration file.
	Path string

	// Perm — file permissions for reading/writing the config file.
	Perm fs.FileMode

	// UnableToReplaceComments True if some comments could be replaced
	UnableToReplaceComments bool
}

// Driver is the interface that every config driver must implement.
type Driver interface {
	// Unmarshal — deserializes data from []byte into a map.
	Unmarshal([]byte, interface{}) error

	// Marshal — serializes a map into []byte.
	Marshal(interface{}, bool) ([]byte, error)

	// Version — extracts the version number from the data.
	Version([]byte) (int, bool, error)

	// EmptyData — returns the default empty data content.
	EmptyData() []byte
}

// Open returns a new instance of a migration database driver using the given URL.
func Open(url string) (database.Driver, error) {
	return database.Open(url)
}

// Register globally registers a new config driver with the specified name and settings.
func Register(name string, driver Driver, cfg Settings) {
	m := New(driver, cfg)
	database.Register(name, m)
}

// List returns a list of all registered migration drivers.
func List() []string {
	return database.List()
}
