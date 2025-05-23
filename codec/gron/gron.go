package gron

import (
	"bytes"
	"fmt"
	"github.com/JFryy/qq/codec/util"
	"github.com/goccy/go-json"
	"reflect"
	"strconv"
	"strings"
)

type Codec struct{}

func (c *Codec) Unmarshal(data []byte, v any) error {
	lines := strings.Split(string(data), "\n")
	var isArray bool
	dataMap := make(map[string]any)
	arrayData := make([]any, 0)

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		parts := strings.SplitN(line, " = ", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid line format: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(parts[1], `";`)
		parsedValue := util.ParseValue(value)

		if strings.HasPrefix(key, "[") && strings.Contains(key, "]") {
			isArray = true
		}

		c.setValueJSON(dataMap, key, parsedValue)
	}

	if isArray && len(dataMap) == 1 {
		for _, val := range dataMap {
			if arrayVal, ok := val.([]any); ok {
				arrayData = arrayVal
			}
		}
	}

	vv := reflect.ValueOf(v)
	if vv.Kind() != reflect.Ptr || vv.IsNil() {
		return fmt.Errorf("provided value must be a non-nil pointer")
	}
	if isArray && len(arrayData) > 0 {
		vv.Elem().Set(reflect.ValueOf(arrayData))
	} else {
		vv.Elem().Set(reflect.ValueOf(dataMap))
	}

	return nil
}

func (c *Codec) Marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	c.traverseJSON("", v, &buf)
	return buf.Bytes(), nil
}

func (c *Codec) traverseJSON(prefix string, v any, buf *bytes.Buffer) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Map:
		for _, key := range rv.MapKeys() {
			strKey := fmt.Sprintf("%v", key)
			c.traverseJSON(addPrefix(prefix, strKey), rv.MapIndex(key).Interface(), buf)
		}
	case reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			c.traverseJSON(fmt.Sprintf("%s[%d]", prefix, i), rv.Index(i).Interface(), buf)
		}
	default:
		buf.WriteString(fmt.Sprintf("%s = %s;\n", prefix, formatJSONValue(v)))
	}
}

func addPrefix(prefix, name string) string {
	if prefix == "" {
		return name
	}
	if strings.Contains(name, "[") && strings.Contains(name, "]") {
		return prefix + name
	}
	return prefix + "." + name
}

func formatJSONValue(v any) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case bool:
		return strconv.FormatBool(val)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	default:
		if v == nil {
			return "null"
		}
		data, _ := json.Marshal(v)
		return string(data)
	}
}

func (c *Codec) setValueJSON(data map[string]any, key string, value any) {
	parts := strings.Split(key, ".")
	var m = data
	for i, part := range parts {
		if i == len(parts)-1 {
			if strings.Contains(part, "[") && strings.Contains(part, "]") {
				k := strings.Split(part, "[")[0]
				index := parseArrayIndex(part)
				if _, ok := m[k]; !ok {
					m[k] = make([]any, index+1)
				}
				arr := m[k].([]any)
				if len(arr) <= index {
					for len(arr) <= index {
						arr = append(arr, nil)
					}
					m[k] = arr
				}
				arr[index] = value
			} else {
				m[part] = value
			}
		} else {
			// fix index assignment nested map: this is needs optimization
			if strings.Contains(part, "[") && strings.Contains(part, "]") {
				k := strings.Split(part, "[")[0]
				index := parseArrayIndex(part)
				if _, ok := m[k]; !ok {
					m[k] = make([]any, index+1)
				}
				arr := m[k].([]any)
				if len(arr) <= index {
					for len(arr) <= index {
						arr = append(arr, nil)
					}
					m[k] = arr
				}
				if arr[index] == nil {
					arr[index] = make(map[string]any)
				}
				m = arr[index].(map[string]any)
			} else {
				if _, ok := m[part]; !ok {
					m[part] = make(map[string]any)
				}
				m = m[part].(map[string]any)
			}
		}
	}
}

func parseArrayIndex(part string) int {
	indexStr := strings.Trim(part[strings.Index(part, "[")+1:strings.Index(part, "]")], " ")
	index, _ := strconv.Atoi(indexStr)
	return index
}
