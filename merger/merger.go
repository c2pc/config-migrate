package merger

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/c2pc/config-migrate/replacer"
)

const deprecatedSuffix = "_deprecated"

// Keys ending with _deprecated_expand: value is "path->field" — take array at path from old,
// put each element into the given field of the template objects from new[targetKey].
const deprecatedExpandSuffix = "_deprecated_expand"

// Keys ending with _deprecated_collapse: value is "arrayPath.field->targetPath" — take array at
// arrayPath from old, extract field from each element, write array of scalars at targetPath.
const deprecatedCollapseSuffix = "_deprecated_collapse"

// Keys ending with _replace force the target to use the new value (override merged old).
const deprecatedReplaceSuffix = "_deprecated_replace"

// Keys ending with _deprecated_concat: value is "path1,path2->template". Values at path1, path2 (from the
// already-merged map m, after _deprecated/_deprecated_expand/_deprecated_collapse) are substituted into
// template as {0}, {1}, ... and the result is written to the target key.
const deprecatedConcatSuffix = "_deprecated_concat"

func Merge(new, old map[string]interface{}) map[string]interface{} {
	newCopy := deepCopyMap(new)
	m := mergeMaps(newCopy, old)
	applyDeprecatedInto(m, newCopy, old)
	applyDeprecatedExpandInto(m, new, old)
	applyDeprecatedCollapseInto(m, m, new, old)
	applyDeprecatedConcatInto(m, m, new)
	// Use original new so _replace sees the intended new values (merge overwrote newCopy).
	applyReplaceInto(m, new)

	if replacer.HasReplacers() {
		for k, v := range m {
			m[k] = replace(v)
		}
	}

	return m
}

// deepCopyMap recursively copies a map (and nested maps/slices) so merge can mutate the copy.
func deepCopyMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = deepCopyValue(v)
	}
	return out
}

func deepCopyValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]interface{}); ok {
		return deepCopyMap(m)
	}
	if s, ok := v.([]interface{}); ok {
		cp := make([]interface{}, len(s))
		for i := range s {
			cp[i] = deepCopyValue(s[i])
		}
		return cp
	}
	return v
}

// getValueByPath returns value at dot-separated path (e.g. "sql.db.url") in m. Root is m.
func getValueByPath(m map[string]interface{}, path string) (interface{}, bool) {
	if m == nil || path == "" {
		return nil, false
	}
	parts := strings.Split(path, ".")
	var current interface{} = m
	for _, key := range parts {
		mp, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		var exists bool
		current, exists = mp[key]
		if !exists {
			return nil, false
		}
	}
	return current, true
}

// setValueByPath sets value at dot-separated path in m, creating nested maps as needed.
func setValueByPath(m map[string]interface{}, path string, value interface{}) {
	if m == nil || path == "" {
		return
	}
	parts := strings.Split(path, ".")
	current := m
	for i, key := range parts {
		if i == len(parts)-1 {
			current[key] = value
			return
		}
		next, ok := current[key].(map[string]interface{})
		if !ok || next == nil {
			next = make(map[string]interface{})
			current[key] = next
		}
		current = next
	}
}

// mergeDeprecatedIntoTarget merges deprecated value into the target key value.
// newTarget is the new config value for the key (defines desired type: array vs scalar).
// - target array + scalar deprecated -> [..., deprecated]
// - target array + array deprecated -> [...target, ...deprecated]
// - target scalar + array deprecated -> first element of deprecated
// - target scalar + scalar deprecated -> deprecated
func mergeDeprecatedIntoTarget(oldTarget, deprecatedVal, newTarget interface{}) interface{} {
	oldArr, oldIsArr := toSlice(oldTarget)
	depArr, depIsArr := toSlice(deprecatedVal)
	_, newIsArr := toSlice(newTarget)

	if oldIsArr && depIsArr {
		return appendSlice(oldArr, depArr...)
	}
	if oldIsArr {
		return appendSlice(oldArr, deprecatedVal)
	}
	if depIsArr {
		if len(depArr) == 0 {
			return oldTarget
		}
		return depArr[0]
	}
	// both scalar: if new expects array, wrap deprecated in array
	if newIsArr {
		return appendSlice(nil, deprecatedVal)
	}
	return deprecatedVal
}

func toSlice(v interface{}) ([]interface{}, bool) {
	if v == nil {
		return nil, false
	}
	if sl, ok := v.([]interface{}); ok {
		return sl, true
	}
	// Handle other slice types (e.g. []string from YAML/JSON decode)
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice {
		n := rv.Len()
		out := make([]interface{}, n)
		for i := 0; i < n; i++ {
			out[i] = rv.Index(i).Interface()
		}
		return out, true
	}
	return nil, false
}

func appendSlice(a []interface{}, els ...interface{}) []interface{} {
	if a == nil {
		a = []interface{}{}
	}
	return append(a, els...)
}

// applyDeprecatedInto applies deprecated rules into m in place: for each *_deprecated in new,
// pulls value from old by path and merges into m[targetKey]. Uses root old for paths.
func applyDeprecatedInto(m, new, old map[string]interface{}) {
	if m == nil || new == nil || old == nil {
		return
	}
	for k, v := range new {
		if !strings.HasSuffix(k, deprecatedSuffix) {
			if newMap, ok := v.(map[string]interface{}); ok {
				if mChild, ok := m[k].(map[string]interface{}); ok {
					applyDeprecatedInto(mChild, newMap, old)
				}
			}
			continue
		}
		path, ok := v.(string)
		if !ok {
			continue
		}
		targetKey := strings.TrimSuffix(k, deprecatedSuffix)
		deprecatedVal, found := getValueByPath(old, path)
		if !found {
			continue
		}
		oldTarget, _ := getValueByPath(old, targetKey)
		newTarget := new[targetKey]

		_, depIsMap := deprecatedVal.(map[string]interface{})
		newMap, newIsMap := newTarget.(map[string]interface{})

		if depIsMap && newIsMap {
			// Nested object: m[targetKey] already has new structure from merge; just recurse to apply inner _deprecated
			mChild, _ := m[targetKey].(map[string]interface{})
			if mChild != nil {
				applyDeprecatedInto(mChild, newMap, old)
			}
		} else {
			merged := mergeDeprecatedIntoTarget(oldTarget, deprecatedVal, newTarget)
			m[targetKey] = merged
		}
	}
	// Remove _deprecated keys from m
	for k := range m {
		if strings.HasSuffix(k, deprecatedSuffix) {
			delete(m, k)
		}
	}
}

// parseExpandSpec parses "path->field" and returns path, field. If no "->", returns "", "".
func parseExpandSpec(s string) (path, field string) {
	i := strings.Index(s, "->")
	if i < 0 {
		return "", ""
	}
	return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+2:])
}

// applyDeprecatedExpandInto handles key_deprecated_expand: "path->field". Source array at path in old
// is merged with template array from new[targetKey]: for each index i, result[i] = template[i] with [field] = source[i].
func applyDeprecatedExpandInto(m, new, old map[string]interface{}) {
	if m == nil || new == nil || old == nil {
		return
	}
	for k, v := range new {
		if !strings.HasSuffix(k, deprecatedExpandSuffix) {
			continue
		}
		spec, ok := v.(string)
		if !ok {
			continue
		}
		path, field := parseExpandSpec(spec)
		if path == "" || field == "" {
			continue
		}
		targetKey := strings.TrimSuffix(k, deprecatedExpandSuffix)
		sourceVal, found := getValueByPath(old, path)
		if !found {
			continue
		}
		sourceArr, ok := toSlice(sourceVal)
		if !ok {
			continue
		}
		templateVal := new[targetKey]
		templateArr, _ := toSlice(templateVal)
		result := make([]interface{}, 0, len(sourceArr))
		for i, scalar := range sourceArr {
			var obj map[string]interface{}
			if i < len(templateArr) {
				if t, ok := templateArr[i].(map[string]interface{}); ok {
					obj = deepCopyMap(t)
				} else {
					obj = make(map[string]interface{})
				}
			} else if len(templateArr) > 0 {
				if t, ok := templateArr[len(templateArr)-1].(map[string]interface{}); ok {
					obj = deepCopyMap(t)
				} else {
					obj = make(map[string]interface{})
				}
			} else {
				obj = make(map[string]interface{})
			}
			obj[field] = scalar
			result = append(result, obj)
		}
		m[targetKey] = result
	}
	for k := range m {
		if strings.HasSuffix(k, deprecatedExpandSuffix) {
			delete(m, k)
		}
	}
	// Recurse into nested maps
	for k, v := range new {
		if strings.HasSuffix(k, deprecatedExpandSuffix) || strings.HasSuffix(k, deprecatedSuffix) || strings.HasSuffix(k, deprecatedReplaceSuffix) {
			continue
		}
		if newMap, ok := v.(map[string]interface{}); ok {
			if mChild, ok := m[k].(map[string]interface{}); ok {
				applyDeprecatedExpandInto(mChild, newMap, old)
			}
		}
	}
}

// parseCollapseSpec parses "arrayPath.field->targetPath" and returns arrayPath, field, targetPath.
// Example: "fps.urls.url->fps.urls" → ("fps.urls", "url", "fps.urls").
func parseCollapseSpec(s string) (arrayPath, field, targetPath string) {
	i := strings.Index(s, "->")
	if i < 0 {
		return "", "", ""
	}
	left := strings.TrimSpace(s[:i])
	targetPath = strings.TrimSpace(s[i+2:])
	if left == "" || targetPath == "" {
		return "", "", ""
	}
	j := strings.LastIndex(left, ".")
	if j < 0 {
		return "", "", ""
	}
	arrayPath = strings.TrimSpace(left[:j])
	field = strings.TrimSpace(left[j+1:])
	return arrayPath, field, targetPath
}

// applyDeprecatedCollapseInto handles key_deprecated_collapse: "arrayPath.field->targetPath".
// Reads array at arrayPath from old, extracts field from each element, writes []scalars at targetPath in rootM.
func applyDeprecatedCollapseInto(rootM, m, new, old map[string]interface{}) {
	if rootM == nil || m == nil || new == nil || old == nil {
		return
	}
	for k, v := range new {
		if !strings.HasSuffix(k, deprecatedCollapseSuffix) {
			continue
		}
		spec, ok := v.(string)
		if !ok {
			continue
		}
		arrayPath, field, targetPath := parseCollapseSpec(spec)
		if arrayPath == "" || field == "" || targetPath == "" {
			continue
		}
		sourceVal, found := getValueByPath(old, arrayPath)
		if !found {
			continue
		}
		sourceArr, ok := toSlice(sourceVal)
		if !ok {
			continue
		}
		result := make([]interface{}, 0, len(sourceArr))
		for _, elem := range sourceArr {
			obj, ok := elem.(map[string]interface{})
			if !ok {
				continue
			}
			if val, exists := obj[field]; exists {
				result = append(result, val)
			}
		}
		setValueByPath(rootM, targetPath, result)
	}
	for k := range m {
		if strings.HasSuffix(k, deprecatedCollapseSuffix) {
			delete(m, k)
		}
	}
	// Recurse into nested maps
	for k, v := range new {
		if strings.HasSuffix(k, deprecatedCollapseSuffix) || strings.HasSuffix(k, deprecatedExpandSuffix) || strings.HasSuffix(k, deprecatedSuffix) || strings.HasSuffix(k, deprecatedReplaceSuffix) {
			continue
		}
		if newMap, ok := v.(map[string]interface{}); ok {
			if mChild, ok := m[k].(map[string]interface{}); ok {
				applyDeprecatedCollapseInto(rootM, mChild, newMap, old)
			}
		}
	}
}

// parseConcatSpec parses "path1,path2,...->template" and returns paths (trimmed) and template.
func parseConcatSpec(s string) (paths []string, template string) {
	i := strings.Index(s, "->")
	if i < 0 {
		return nil, ""
	}
	left := strings.TrimSpace(s[:i])
	template = strings.TrimSpace(s[i+2:])
	if left == "" || template == "" {
		return nil, ""
	}
	for _, p := range strings.Split(left, ",") {
		paths = append(paths, strings.TrimSpace(p))
	}
	return paths, template
}

// applyDeprecatedConcatInto applies key_deprecated_concat: "path1,path2->template". Paths are resolved
// against rootM (the merged map after _deprecated/_deprecated_expand/_deprecated_collapse), so concat
// glues already-transformed fields. Template uses {0}, {1}, ...; result is written to target key in m.
func applyDeprecatedConcatInto(rootM, m, new map[string]interface{}) {
	if rootM == nil || m == nil || new == nil {
		return
	}
	for k, v := range new {
		if !strings.HasSuffix(k, deprecatedConcatSuffix) {
			continue
		}
		spec, ok := v.(string)
		if !ok {
			continue
		}
		paths, template := parseConcatSpec(spec)
		if len(paths) == 0 || template == "" {
			continue
		}
		var parts []string
		for _, path := range paths {
			val, found := getValueByPath(rootM, path)
			if !found {
				parts = append(parts, "")
				continue
			}
			parts = append(parts, fmt.Sprint(val))
		}
		result := template
		for i, part := range parts {
			result = strings.ReplaceAll(result, "{"+strconv.Itoa(i)+"}", part)
		}
		targetKey := strings.TrimSuffix(k, deprecatedConcatSuffix)
		m[targetKey] = result
	}
	for k := range m {
		if strings.HasSuffix(k, deprecatedConcatSuffix) {
			delete(m, k)
		}
	}
	// Recurse into nested maps
	for k, v := range new {
		if strings.HasSuffix(k, deprecatedConcatSuffix) || strings.HasSuffix(k, deprecatedCollapseSuffix) ||
			strings.HasSuffix(k, deprecatedExpandSuffix) || strings.HasSuffix(k, deprecatedSuffix) ||
			strings.HasSuffix(k, deprecatedReplaceSuffix) {
			continue
		}
		if newMap, ok := v.(map[string]interface{}); ok {
			if mChild, ok := m[k].(map[string]interface{}); ok {
				applyDeprecatedConcatInto(rootM, mChild, newMap)
			}
		}
	}
}

// applyReplaceInto applies _replace keys from new into m: for each key_replace in new,
// set m[key] = new[key] (the target key's value in the new config), so the new value
// overwrites whatever the merge kept. The _replace key is just a marker (e.g. empty);
// the actual value is taken from the sibling key in new.
func applyReplaceInto(m, new map[string]interface{}) {
	if m == nil || new == nil {
		return
	}
	for k := range new {
		if strings.HasSuffix(k, deprecatedReplaceSuffix) {
			targetKey := strings.TrimSuffix(k, deprecatedReplaceSuffix)
			if newVal, exists := new[targetKey]; exists {
				m[targetKey] = newVal
			}
			continue
		}
		v := new[k]
		if newMap, ok := v.(map[string]interface{}); ok {
			if mChild, ok := m[k].(map[string]interface{}); ok {
				applyReplaceInto(mChild, newMap)
			}
		}
	}
	for k := range m {
		if strings.HasSuffix(k, deprecatedReplaceSuffix) {
			delete(m, k)
		}
	}
}

func mergeMaps(out, old map[string]interface{}) map[string]interface{} {
	if len(out) == 0 {
		return map[string]interface{}{}
	}
	if len(old) == 0 {
		return out
	}

	for key, oldVal := range old {
		newVal, exists := out[key]
		if !exists {
			continue
		}

		out[key] = mergeValues(newVal, oldVal)
	}

	return out
}

// mergeValues returns the merged value for one key. Precedence: both maps → recurse; both arrays → merge; same type → old; nils handled.
func mergeValues(newVal, oldVal interface{}) interface{} {
	newMap, newIsMap := newVal.(map[string]interface{})
	oldMap, oldIsMap := oldVal.(map[string]interface{})
	newArr, newIsArr := newVal.([]interface{})
	oldArr, oldIsArr := oldVal.([]interface{})

	switch {
	case newIsMap && oldIsMap:
		return mergeMaps(newMap, oldMap)

	case newIsArr && oldIsArr:
		return mergeArrays(newArr, oldArr)

	case isSameType(newVal, oldVal):
		return oldVal

	case oldVal == nil:
		return nil

	case newVal == nil:
		return oldVal
	}

	return newVal
}

func mergeArrays(newArr, oldArr []interface{}) interface{} {
	if len(newArr) == 0 || len(oldArr) == 0 {
		return oldArr
	}
	newElem, newElemIsMap := newArr[0].(map[string]interface{})
	if !newElemIsMap {
		return oldArr
	}

	var result []map[string]interface{}
	for i, oldItem := range oldArr {
		oldElem, oldIsMap := oldItem.(map[string]interface{})
		if oldIsMap {
			templateCopy := deepCopyMap(newElem)
			result = append(result, mergeMaps(templateCopy, oldElem))
		} else if i < len(newArr) {
			templateCopy := deepCopyMap(newElem)
			result = append(result, mergeMaps(templateCopy, templateCopy))
		}
	}
	return result
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
