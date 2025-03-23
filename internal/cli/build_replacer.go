//go:build replacer
// +build replacer

package cli

import (
	_ "github.com/c2pc/config-migrate/replace/ip"
	_ "github.com/c2pc/config-migrate/replace/project_name"
	_ "github.com/c2pc/config-migrate/replace/random"
)
