# Config-Migrate

`config-migrate` is a plugin for [golang-migrate](https://github.com/golang-migrate/migrate) that enables versioned migrations for configuration files like YAML and JSON.

## Configs

Config drivers apply migrations to configuration files instead of traditional databases.  
[Want to add a new config driver?](driver/driver.go)

Currently supported config drivers:

* [JSON](driver/json)
* [YAML](driver/yaml)

## Why use `config-migrate`?

This is useful when:

- You want to apply versioned updates to your application configs.
- You need a reproducible and auditable history of config changes.
- You want to manage infrastructure or system configs as part of CI/CD.

## Features

* Seamless integration with [`golang-migrate`](https://github.com/golang-migrate/migrate)
* File-based locking to prevent concurrent writes
* Config merging with support for version tracking
* Supports `version`, `force`, and `drop` commands
* Graceful file handling using `io.Reader` / `io.Writer`
* Thread-safe with `sync.Mutex`
* Built-in support for YAML and JSON config formats

## Getting Started

### Install

Make sure you have Go 1.18+ and Go modules enabled.

```bash
go get github.com/c2pc/config-migrate
```

## Use in your Go project

* API is stable and frozen for this release (v3 & v4).
* Uses [Go modules](https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more) to manage dependencies.
* Supports graceful stop via `GracefulStop chan bool`.
* Bring your own logger.
* Uses `io.Reader` streams internally for low memory overhead.
* Thread-safe and no goroutine leaks.

📚 __[Go Documentation](https://pkg.go.dev/github.com/golang-migrate/migrate/v4)__

### Basic usage with config file:

```go
import (
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/c2pc/config-migrate/driver/json"
    _ "github.com/golang-migrate/migrate/v4/source/github"
)

func main() {
    m, err := migrate.New(
        "github://username:token@your-repo/json-migrations",
        "json://./config.json")
    if err != nil {
        panic(err)
    }

    // Apply the next 2 migration steps
    m.Steps(2)
}
```

### Advanced usage with an existing config client:

```go
import (
    "github.com/golang-migrate/migrate/v4"
    "github.com/c2pc/config-migrate/driver/yaml"
    "github.com/c2pc/config-migrate/config"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
    driver := yaml.New(driver.Settings{
        Path: "config.yml",
        Perm: 0777,
    })

    m, err := migrate.NewWithDatabaseInstance(
        "file://driver/migrations",
        "yaml", driver)
    if err != nil {
        panic(err)
    }

    m.Up()
}
```

## Dynamic Replacers

You can use dynamic placeholders in your config files and define how they should be replaced at runtime using `replacer`.

### Built-in Example: Random

The following example registers a replacer that replaces `___random___` with random string:

```go
package random

import (
	"crypto/rand"

	"github.com/c2pc/config-migrate/replacer"
)

const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func init() {
	replacer.Register("___random___", randomReplacer(16))
}

func randomReplacer(n int) func() string {
	return func() string {
		bytes := make([]byte, n)
		_, err := rand.Read(bytes)
		if err != nil {
			return ""
		}

		for i, b := range bytes {
			bytes[i] = letters[b%byte(len(letters))]
		}

		return string(bytes)
	}
}

```

### Writing Your Own Replacers

You can create and register your own replacers by calling `replacer.Register`:

```go
package main

import (
    "github.com/c2pc/config-migrate/replacer"
)

func init() {
    replacer.Register("___my_placeholder___", func() string {
        return "dynamic-value"
    })
}
```

These placeholders will be replaced automatically when configs are processed during migrations.

### When to Use Replacers

Use replacers when you want to inject dynamic values (like IPs, ports, timestamps, environment info) into your config files at the time of applying a migration.


## Comments

you can enable comment replacer in settings `UnableToReplaceComments: true`
```go
package main

import (
	"embed"
	"errors"

	"github.com/c2pc/config-migrate/driver"
	"github.com/c2pc/config-migrate/driver/yaml"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.yml
var fs embed.FS

func runMigration(path string) error {
	//You can set a comment suffix
	driver.CommentSuffix = "___comment___"
	
	//Enable to replace comments
	yamlMigr := yaml.New(driver.Settings{
		Path:                    path,
		Perm:                    0777,
		UnableToReplaceComments: true,
	})

	//Using iofs as a source
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}

	//Create yaml migration
	m, err := migrate.NewWithInstance("iofs", d, "yaml", yamlMigr)
	if err != nil {
		return err
	}

	//Run migrations
	if err := m.Up(); errors.Is(err, migrate.ErrNoChange) {
		return nil
	} else if err != nil {
		return err
	}

	return nil
}

```


### YAML
You can add comments to your parameters. Add suffix `______` to parameter name
Set empty comment `host______:` to add `\n` to file
```yaml
http______:
http_______: HTTP server configuration1
http________: HTTP server configuration2
http:
  host______:
  host_______: The IP address the server will bind to
  host: ___ip_address___
  port______:
  port_______: The port the server will listen on
  port: 8052
```

As a result we get
```yaml

# HTTP server configuration1
# HTTP server configuration2
http:

  # The IP address the server will bind to
  host: ___ip_address___

  # The port the server will listen on
  port: 8052
```

### JSON
You can add comments to your parameters. Add suffix `______` to parameter name.
Set empty comment `"http______": ""` to add `\n` to file
```json
{
  "http______": "",
  "http_______": "HTTP server configuration",
  "http________": "HTTP server configuration1",
  "http_________": "HTTP server configuration2",
  "http": {
    "host______": "",
    "host_______": "The IP address the server will bind to",
    "host": "___ip_address___",

    "port______": "",
    "port_______": "The port the server will listen on",
    "port": 8052
  }
}
```

As a result we get
```json
{
  
  "____http": "HTTP server configuration",
  "____http_": "HTTP server configuration1",
  "____http__": "HTTP server configuration2",
  "http": {
    
    "____host": "The IP address the server will bind to",
    "host": "___ip_address___",

    "____port": "The port the server will listen on",
    "port": 8052
  }
}
```


### Examples

* [JSON](driver/json/examples/migrations) - JSON migrations with replacers and comments
* [YAML](driver/yaml/examples/migrations) - YAML migrations with replacers and comments
* [EXAMPLE](example) - Example application with YAML migrations