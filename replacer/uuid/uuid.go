package random

import (
	"github.com/c2pc/config-migrate/replacer"
	"github.com/google/uuid"
)

func init() {
	replacer.Register("___uuid___", uuidReplacer)
}

func uuidReplacer() string {
	return uuid.New().String()
}
