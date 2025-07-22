package encoder

import (
	"fmt"
	"strings"
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
		input    []any
		expected string
	}{
		{[]any{1, "hello", []any{2, "world"}}, "li1e5:helloli2e5:worldee"},
		{[]any{}, "le"},
	}

	for _, test := range tests {
		result := EncodeList(test.input)
		if result != test.expected {
			t.Errorf("EncodeList(%v) = %s; want %s", test.input, result, test.expected)
		}
	}
}

func TestEncodeDictionary(t *testing.T) {
	// Test single key dictionary (deterministic)
	singleKeyDict := map[string]any{
		"key": "value",
	}
	result := EncodeDictionary(singleKeyDict)
	expected := "d3:key5:valuee"
	if result != expected {
		t.Errorf("EncodeDictionary(%v) = %s; want %s", singleKeyDict, result, expected)
	}

	// Test multi-key dictionary (check components due to non-deterministic map iteration)
	multiKeyDict := map[string]any{
		"key1": "value1",
		"key2": "value2",
	}
	result = EncodeDictionary(multiKeyDict)

	// Check that result starts with 'd' and ends with 'e'
	if len(result) < 2 || result[0] != 'd' || result[len(result)-1] != 'e' {
		t.Errorf("EncodeDictionary result should start with 'd' and end with 'e', got: %s", result)
	}

	// Check that both key-value pairs are present
	if !strings.Contains(result, "4:key16:value1") || !strings.Contains(result, "4:key26:value2") {
		t.Errorf("EncodeDictionary(%v) = %s; should contain both key-value pairs", multiKeyDict, result)
	}
}

// Benchmark functions
func BenchmarkEncodeInteger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EncodeInteger(123456)
	}
}

func BenchmarkEncodeIntegerNegative(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EncodeInteger(-123456)
	}
}

func BenchmarkEncodeString(b *testing.B) {
	testString := "hello world this is a test string"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeString(testString)
	}
}

func BenchmarkEncodeStringLarge(b *testing.B) {
	// Create a larger string for more realistic benchmarking
	largeString := ""
	for i := 0; i < 1000; i++ {
		largeString += "abcdefghij"
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeString(largeString)
	}
}

func BenchmarkEncodeList(b *testing.B) {
	testList := []any{1, 2, 3, "hello", "world", []any{4, 5, "nested"}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeList(testList)
	}
}

func BenchmarkEncodeListLarge(b *testing.B) {
	// Create a larger list
	largeList := make([]any, 0, 1000)
	for i := 0; i < 500; i++ {
		largeList = append(largeList, i)
		largeList = append(largeList, "string"+string(rune(i)))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeList(largeList)
	}
}

func BenchmarkEncodeDictionary(b *testing.B) {
	testDict := map[string]any{
		"key1":   "value1",
		"key2":   123,
		"key3":   []any{1, 2, "nested"},
		"longer": "this is a longer value to test",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeDictionary(testDict)
	}
}

func BenchmarkEncodeDictionaryLarge(b *testing.B) {
	// Create a larger dictionary
	largeDict := make(map[string]any)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)
		if i%3 == 0 {
			largeDict[key] = i
		} else if i%3 == 1 {
			largeDict[key] = fmt.Sprintf("value%d", i)
		} else {
			largeDict[key] = []any{i, fmt.Sprintf("nested%d", i)}
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncodeDictionary(largeDict)
	}
}
