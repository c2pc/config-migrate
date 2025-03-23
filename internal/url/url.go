package url

import (
	nurl "net/url"
	"os"
	"path/filepath"
)

func ParseURL(url string) (string, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return "", err
	}
	p := u.Opaque
	if len(p) == 0 {
		p = u.Host + u.Path
	}

	if len(p) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		p = wd
	} else if !filepath.IsAbs(p) {
		abs, err := filepath.Abs(p)
		if err != nil {
			return "", err
		}
		p = abs
	}
	return p, nil
}
