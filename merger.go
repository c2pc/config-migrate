package config_migrate

import (
	"github.com/c2pc/config-migrate/internal/merger"
)

func MergeMaps(new, old map[string]interface{}) map[string]interface{} {
	return merger.Merge(new, old)
}
