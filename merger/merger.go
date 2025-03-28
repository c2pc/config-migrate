package merger

import (
	"reflect"

	"github.com/c2pc/config-migrate/replacer"
)

func Merge(new, old map[string]interface{}) map[string]interface{} {
	delete(old, "force")
	delete(new, "force")
	delete(old, "version")
	delete(new, "version")
	return mergeMaps(new, old)
}

func mergeMaps(new, old map[string]interface{}) map[string]interface{} {
	//If the new map is empty
	if new == nil || len(new) == 0 {
		return map[string]interface{}{}
	}

	//Create a new migration results map
	out := make(map[string]interface{})
	for key, value := range new {
		out[key] = replace(value)
	}

	//If the old map is empty, start merging,
	//otherwise start creating only a new one.
	if old != nil && len(old) > 0 {
		for oldKey, oldValue := range old {
			//If the key is present in the new map, otherwise skip
			if newValue, exists := out[oldKey]; exists {
				newValueMap, newValueMapOK := newValue.(map[string]interface{})
				oldValueMap, oldValueMapOK := oldValue.(map[string]interface{})

				//If interface of new and old is map[string]interface{}
				if newValueMapOK && oldValueMapOK {
					//Run merger for maps
					out[oldKey] = mergeMaps(newValueMap, oldValueMap)
				} else {
					newValueArray, newValueArrayOK := newValue.([]interface{})
					oldValueArray, oldValueArrayOK := oldValue.([]interface{})

					//If interface of new and old is []interface{}
					if newValueArrayOK && oldValueArrayOK {
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
								out[oldKey] = replace(oldValue)
							}
						}
					} else
					//If the types of the new and old values are equal use old value
					if reflect.TypeOf(newValue) == reflect.TypeOf(oldValue) {
						out[oldKey] = replace(oldValue)
					} else
					//If old value is empty,
					//otherwise a new one is assigned
					if reflect.TypeOf(oldValue) == nil {
						out[oldKey] = nil
					}
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
