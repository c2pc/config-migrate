package migrate

import (
	"github.com/c2pc/golang-file-migrate/internal/merger"
)

func MergeMaps(new, old map[string]interface{}) map[string]interface{} {
	return merger.Merge(new, old)
}
