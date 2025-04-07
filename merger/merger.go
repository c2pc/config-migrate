package merger

import (
	"github.com/c2pc/config-migrate/replacer"
)

func Merge(new, old map[string]interface{}) map[string]interface{} {
	m := mergeMaps(new, old)

	if replacer.HasReplacers() {
		for k, v := range m {
			m[k] = replace(v)
		}
	}

	return m
}

func mergeMaps(out, old map[string]interface{}) map[string]interface{} {
	//If the new map is empty
	if out == nil || len(out) == 0 {
		return map[string]interface{}{}
	}

	//If the old map is empty, start merging,
	//otherwise start creating only a new one.
	if old != nil && len(old) > 0 {
		for oldKey, oldValue := range old {
			//If the key is present in the new map, otherwise skip
			if newValue, exists := out[oldKey]; exists {
				newValueMap, newValueMapOK := newValue.(map[string]interface{})
				oldValueMap, oldValueMapOK := oldValue.(map[string]interface{})

				newValueArray, newValueArrayOK := newValue.([]interface{})
				oldValueArray, oldValueArrayOK := oldValue.([]interface{})

				//If interface of new and old is map[string]interface{}
				if newValueMapOK && oldValueMapOK {
					//Run merger for maps
					out[oldKey] = mergeMaps(newValueMap, oldValueMap)
				} else if newValueArrayOK && oldValueArrayOK {
					//If the old and new arrays are not empty,
					//otherwise the new one is assigned
					if len(newValueArray) > 0 && len(oldValueArray) > 0 {
						//The interface of the first element of the new array is map[string]interface{},
						//otherwise a old one is assigned
						if newValueArray2Map, newValueArrayOK2Map := newValueArray[0].(map[string]interface{}); newValueArrayOK2Map {
							//Create a new migration results map for array elements
							var newArrayMap []map[string]interface{}
							for i, newValueArrayValue := range oldValueArray {
								//If interface of value is map[string]interface{},
								//otherwise the element is skipped
								if oldValueArray2Map, oldValueArrayOK2Map := newValueArrayValue.(map[string]interface{}); oldValueArrayOK2Map {
									//Run merger for maps
									newArrayMap = append(newArrayMap, mergeMaps(newValueArray2Map, oldValueArray2Map))
								} else {
									if i < len(newValueArray) {
										newArrayMap = append(newArrayMap, mergeMaps(newValueArray2Map, newValueArray2Map))
									}
								}
							}
							//Add new array to results map
							out[oldKey] = newArrayMap
						} else {
							//Add pld value to results map with replacer
							out[oldKey] = oldValue
						}
					}
				} else if isSameType(newValue, oldValue) {
					out[oldKey] = oldValue
				} else if oldValue == nil {
					out[oldKey] = nil
				} else if newValue == nil {
					out[oldKey] = oldValue
				}
			}
		}
	}

	return out
}

func replace(value interface{}) interface{} {
	if str, ok := value.(string); ok {
		return replacer.Replace(str)
	} else if strArray, ok := value.([]interface{}); ok {
		newArray := make([]interface{}, len(strArray))
		for k, val := range strArray {
			newArray[k] = replace(val)
		}
		return newArray
	} else if m, ok := value.(map[string]interface{}); ok {
		for k, val := range m {
			m[k] = replace(val)
		}
		return m
	}

	return value
}

func isSameType(a, b interface{}) bool {
	switch a.(type) {
	case string:
		_, ok := b.(string)
		return ok
	case int:
		_, ok := b.(int)
		return ok
	case float64:
		_, ok := b.(float64)
		return ok
	case bool:
		_, ok := b.(bool)
		return ok
	case map[string]interface{}:
		_, ok := b.(map[string]interface{})
		return ok
	case []interface{}:
		_, ok := b.([]interface{})
		return ok
	default:
		return false
	}
}
