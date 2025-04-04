package yaml

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/c2pc/config-migrate/config"
	"github.com/golang-migrate/migrate/v4/database"
	"gopkg.in/yaml.v3"
)

type Yaml struct {
}

func init() {
	config.Register("yaml", &Yaml{}, config.Settings{})
}

func New(cfg config.Settings) database.Driver {
	return config.New(&Yaml{}, cfg)
}

func (m Yaml) Unmarshal(bytes []byte, i interface{}) error {
	return yaml.Unmarshal(bytes, i)
}

func (m Yaml) Marshal(i interface{}, replaceComments bool) ([]byte, error) {
	b, err := yaml.Marshal(i)
	if err != nil {
		return nil, err
	}

	if replaceComments {
		re := regexp.MustCompile(`^(\s*)(.*)` + config.CommentSuffix + `\d*:\s*(.*)$`)

		scanner := bufio.NewScanner(bytes.NewBuffer(b))
		var result []string

		for scanner.Scan() {
			line := scanner.Text()
			if matches := re.FindStringSubmatch(line); matches != nil {
				indent := matches[1]
				key := matches[2]
				comment := matches[3]
				if len(result) > 0 {
					for c := len(result) - 1; c >= 0; c-- {
						com := fmt.Sprintf("%s# %s", indent, comment)
						if strings.Contains(result[c], key+":") {
							if c == 0 {
								result = append([]string{com}, result...)
							} else if c == len(result)-1 {
								var result2 []string
								result2 = append(result2, result[:c]...)
								result2 = append(result2, com)
								result2 = append(result2, result[c])
								result = append([]string{}, result2...)
							} else {
								var result2 []string
								result2 = append(result2, result[:c]...)
								result2 = append(result2, com)
								result2 = append(result2, result[c:]...)
								result = append([]string{}, result2...)
							}
							break
						}
					}
				}
			} else {
				result = append(result, line)
			}
		}

		output := strings.Join(result, "\n")
		return []byte(output), nil
	}

	return b, nil
}

type version struct {
	Version int  `yaml:"version"`
	Force   bool `yaml:"force"`
}

func (m Yaml) Version(bytes []byte) (int, bool, error) {
	v := new(version)
	if err := m.Unmarshal(bytes, v); err != nil {
		return 0, false, err
	}

	return v.Version, v.Force, nil
}

func (m Yaml) EmptyData() []byte {
	return []byte{}
}
