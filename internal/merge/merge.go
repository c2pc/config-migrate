package merge

import (
	"reflect"

	"github.com/c2pc/golang-file-migrate/internal/replacer"
)

func Merge(new, old map[string]interface{}) map[string]interface{} {
	delete(old, "force")
	delete(old, "version")
	return mergeMaps(new, old)
}

func mergeMaps(new, old map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for key, value := range new {
		out[key] = replace(value)
	}

	if old != nil && len(old) > 0 {
		for oldKey, oldValue := range old {
			if newValue, exists := out[oldKey]; exists {
				newValueMap, okNewValueMap := newValue.(map[string]interface{})
				oldValueMap, okOldValueMap := oldValue.(map[string]interface{})

				newValueArray, okNewValueArray := newValue.([]interface{})
				oldValueArray, okOldValueArray := oldValue.([]interface{})

				if okNewValueMap && okOldValueMap {
					out[oldKey] = mergeMaps(newValueMap, oldValueMap)
				} else if okNewValueArray && okOldValueArray {
					if len(newValueArray) > 0 {
						if len(oldValueArray) > 0 {
							newValueArray2Map, okNewValueArray2 := newValueArray[0].(map[string]interface{})
							if okNewValueArray2 {
								newMap := make([]map[string]interface{}, len(oldValueArray))
								for i, newValueArrayValue := range oldValueArray {
									if i < len(newValueArray) {
										newValueArray2Map = newValueArray[i].(map[string]interface{})
									}
									oldValueArray2Map, okOldValueArray2 := newValueArrayValue.(map[string]interface{})
									if okOldValueArray2 {
										newMap[i] = mergeMaps(newValueArray2Map, oldValueArray2Map)
									}
								}
								out[oldKey] = newMap
							} else {
								out[oldKey] = replace(oldValue)
							}
						} else {
							continue
						}
					}
				} else if reflect.TypeOf(newValue) == reflect.TypeOf(oldValue) {
					out[oldKey] = replace(oldValue)
				} else if reflect.TypeOf(oldValue) == nil {
					out[oldKey] = oldValue
				}
			}
		}
	} else {
		return mergeMaps(out, out)
	}

	return out
}

func replace(value interface{}) interface{} {
	if v2, ok2 := value.(string); ok2 {
		return replacer.Replace(v2)
	} else if v3, ok3 := value.([]string); ok3 {
		for i, c3 := range v3 {
			v3[i] = replacer.Replace(c3)
		}
		return v3
	} else {
		return value
	}
}
