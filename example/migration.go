package main

import (
	"embed"
	"errors"

	migrator "github.com/c2pc/config-migrate/driver"
	"github.com/c2pc/config-migrate/driver/yaml"
	_ "github.com/c2pc/config-migrate/replacer/ip"
	_ "github.com/c2pc/config-migrate/replacer/project_name"
	_ "github.com/c2pc/config-migrate/replacer/random"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.yml
var fs embed.FS

func runMigration(path string) error {
	//Using iofs as a source
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}

	//Create yaml migration
	m, err := migrate.NewWithSourceInstance("iofs", d, "yaml://"+path)
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

func runMigrationWithComments(path string) error {
	//Enable to replace comments
	yamlMigr := yaml.New(migrator.Settings{
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
