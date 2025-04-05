package json

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/c2pc/config-migrate/driver"
	"github.com/golang-migrate/migrate/v4/database"
)

type Json struct {
}

func init() {
	config.Register("json", &Json{}, config.Settings{})
}

func New(cfg config.Settings) database.Driver {
	return config.New(&Json{}, cfg)
}

func (m Json) Unmarshal(bytes []byte, i interface{}) error {
	if bytes == nil || len(bytes) == 0 {
		bytes = []byte(`{}`)
	}
	return json.Unmarshal(bytes, i)
}

func (m Json) Marshal(i interface{}, replaceComments bool) ([]byte, error) {
	b, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		return nil, err
	}

	if replaceComments {
		re := regexp.MustCompile(`^(\s*)"(.*)` + config.CommentSuffix + `"(.*)$`)

		scanner := bufio.NewScanner(bytes.NewBuffer(b))
		var result []string

		for scanner.Scan() {
			line := scanner.Text()
			if matches := re.FindStringSubmatch(line); matches != nil {
				indent := matches[1]
				key := matches[2]
				comment := matches[3]
				if len(result) > 0 {
					k := strings.Repeat("_", len(strings.TrimRight(key, "_"))) + key
					for c := len(result) - 1; c >= 0; c-- {
						var com string
						if comment != `: ""` && comment != `: "",` {
							com = fmt.Sprintf(`%s"%s"%s`, indent, k, comment)
						}

						if strings.Contains(result[c], strings.TrimRight(key, "_")+`":`) {
							if c == 0 {
								result = append([]string{com}, result...)
							} else if c == len(result)-1 {
								var result2 []string
								result2 = append(result2, result[:c]...)
								last := result[c]
								if comment[len(comment)-1:] != "," && last[len(last)-1:] == "," {
									if com != "" {
										com = com + ","
									}
									last = last[:len(last)-1]
								}
								result2 = append(result2, com)
								result2 = append(result2, last)
								result = result2
							} else {
								var result2 []string
								result2 = append(result2, result[:c]...)
								result2 = append(result2, com)
								result2 = append(result2, result[c])
								result2 = append(result2, result[c+1:]...)
								result = result2
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
	Version int  `json:"version"`
	Force   bool `json:"force"`
}

func (m Json) Version(bytes []byte) (int, bool, error) {
	v := new(version)
	if err := m.Unmarshal(bytes, v); err != nil {
		return 0, false, err
	}

	return v.Version, v.Force, nil
}

func (m Json) EmptyData() []byte {
	return []byte("{}")
}
