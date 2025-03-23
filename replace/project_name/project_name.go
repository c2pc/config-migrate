package project_name

import (
	"os"
	"path"
	"regexp"

	"github.com/c2pc/config-migrate/internal/replacer"
)

func init() {
	replacer.Register("___project_name___", projectNameReplacer)
}

func projectNameReplacer() string {
	m := regexp.MustCompile("^([a-zA-Z0-9]+)(.*)")
	template := "${1}"

	e, err := os.Executable()
	if err != nil {
		return m.ReplaceAllString(os.Args[0], template)
	}

	return m.ReplaceAllString(path.Base(path.Dir(e)), template)
}
