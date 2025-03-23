//go:build replacer
// +build replacer

package cli

import (
	_ "github.com/c2pc/golang-file-migrate/replace/ip"
	_ "github.com/c2pc/golang-file-migrate/replace/project_name"
	_ "github.com/c2pc/golang-file-migrate/replace/random"
)
