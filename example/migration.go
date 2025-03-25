package main

import (
	"embed"
	"errors"

	_ "github.com/c2pc/config-migrate/config/yaml"
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
		return err // Возвращаем ошибку, если что-то пошло не так
	}

	return nil
}
