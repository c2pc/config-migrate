package replacer

import "github.com/c2pc/config-migrate/internal/replacer"

type Replacer func() string

func RegisterReplacer(name string, fn Replacer) {
	replacer.Register(name, replacer.Replacer(fn))
}
