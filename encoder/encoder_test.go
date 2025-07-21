package encoder

import (
	"testing"
)

func TestEncodeInteger(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{123, "i123e"},
		{-456, "i-456e"},
		{0, "i0e"},
	}

	for _, test := range tests {
		result := EncodeInteger(test.input)
		if result != test.expected {
			t.Errorf("EncodeInteger(%d) = %s; want %s", test.input, result, test.expected)
		}
	}
}

func TestEncodeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "5:hello"},
		{"world", "5:world"},
		{"", "0:"},
	}

	for _, test := range tests {
		result := EncodeString(test.input)
		if result != test.expected {
			t.Errorf("EncodeString(%s) = %s; want %s", test.input, result, test.expected)
		}
	}
}

func TestEncodeList(t *testing.T) {
	tests := []struct {
		input    []interface{}
		expected string
	}{
		{[]interface{}{1, "hello", []interface{}{2, "world"}}, "li1e5:helloli2e5:worldee"},
		{[]interface{}{}, "le"},
	}

	for _, test := range tests {
		result := EncodeList(test.input)
		if result != test.expected {
			t.Errorf("EncodeList(%v) = %s; want %s", test.input, result, test.expected)
		}
	}
}

func TestEncodeDictionary(t *testing.T) {
	tests := []struct {
		input    map[string]interface{}
		expected string
	}{
		{
			map[string]interface{}{
				"key": "value",
			},
			"d3:key5:valuee",
		},
		{
			map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			"d4:key16:value14:key26:value2e",
		},
	}

	for _, test := range tests {
		result := EncodeDictionary(test.input)
		if result != test.expected {
			t.Errorf("EncodeDictionary(%v) = %s; want %s", test.input, result, test.expected)
		}
	}
}
