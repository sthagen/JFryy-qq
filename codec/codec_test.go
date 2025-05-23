package codec

import (
	"fmt"
	"github.com/goccy/go-json"
	"reflect"
	"testing"
)

func TestGetEncodingType(t *testing.T) {
	tests := []struct {
		input    string
		expected EncodingType
	}{
		{"json", JSON},
		{"yaml", YAML},
		{"yml", YML},
		{"toml", TOML},
		{"hcl", HCL},
		{"tf", TF},
		{"csv", CSV},
		{"xml", XML},
		{"ini", INI},
		{"gron", GRON},
		//		{"html", HTML},
	}

	for _, tt := range tests {
		result, err := GetEncodingType(tt.input)
		if err != nil {
			t.Errorf("unexpected error for type %s: %v", tt.input, err)
		} else if result != tt.expected {
			t.Errorf("expected %v, got %v", tt.expected, result)
		}
	}

	unsupportedResult, err := GetEncodingType("unsupported")
	if err == nil {
		t.Errorf("expected error for unsupported type, got result: %v", unsupportedResult)
	}
}

func TestMarshal(t *testing.T) {
	data := map[string]any{"key": "value"}
	tests := []struct {
		encodingType EncodingType
	}{
		{JSON}, {YAML}, {YML}, {TOML}, {HCL}, {TF}, {CSV}, {XML}, {INI}, {GRON}, {HTML},
	}

	for _, tt := range tests {
		// wrap in an interface for things like CSV that require the basic test data be a []map[string]any
		var currentData any
		currentData = data
		if tt.encodingType == CSV {
			currentData = []any{data}
		}

		_, err := Marshal(currentData, tt.encodingType)
		if err != nil {
			t.Errorf("marshal failed for %v: %v", tt.encodingType, err)
		}
	}
}

func TestUnmarshal(t *testing.T) {
	jsonData := `{"key": "value"}`
	xmlData := `<root><key>value</key></root>`
	yamlData := "key: value"
	tomlData := "key = \"value\""
	gronData := `key = "value";`
	tfData := `key = "value"`
	// note: html and csv tests are not yet functional
	//	htmlData := `<html><body><key>value</key></body></html>`
	//	csvData := "key1,key2\nvalue1,value2\nvalue3,value4"

	tests := []struct {
		input        []byte
		encodingType EncodingType
		expected     any
	}{
		{[]byte(jsonData), JSON, map[string]any{"key": "value"}},
		{[]byte(xmlData), XML, map[string]any{"root": map[string]any{"key": "value"}}},
		{[]byte(yamlData), YAML, map[string]any{"key": "value"}},
		{[]byte(tomlData), TOML, map[string]any{"key": "value"}},
		{[]byte(gronData), GRON, map[string]any{"key": "value"}},
		{[]byte(tfData), TF, map[string]any{"key": "value"}},
		//		{[]byte(htmlData), HTML, map[string]any{"html": map[string]any{"body": map[string]any{"key": "value"}}}},
		//		{[]byte(csvData), CSV, []map[string]any{
		//			{"key1": "value1", "key2": "value2"},
		//			{"key1": "value3", "key2": "value4"},
		//		}},
	}

	for _, tt := range tests {
		var data any
		err := Unmarshal(tt.input, tt.encodingType, &data)
		if err != nil {
			t.Errorf("unmarshal failed for %v: %v", tt.encodingType, err)
		}

		expectedJSON, _ := json.Marshal(tt.expected)
		actualJSON, _ := json.Marshal(data)

		if !reflect.DeepEqual(data, tt.expected) {
			fmt.Printf("expected: %s\n", string(expectedJSON))
			fmt.Printf("got: %s\n", string(actualJSON))
			t.Errorf("%s: expected %v, got %v", tt.encodingType, tt.expected, data)
		}
	}
}
