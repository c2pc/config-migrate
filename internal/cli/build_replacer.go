//go:build replacer
// +build replacer

package cli

import (
	_ "github.com/c2pc/config-migrate/replacer/ip"
	_ "github.com/c2pc/config-migrate/replacer/project_name"
	_ "github.com/c2pc/config-migrate/replacer/random"
)
