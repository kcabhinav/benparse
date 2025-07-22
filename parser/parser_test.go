package parser

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestParseInteger(t *testing.T) {
	t.Run("Testing For 0 input", func(t *testing.T) {
		str := "i0e"
		got, _ := ParseInteger(str)
		want := 0

		if got != want {
			t.Errorf("Got %d Wanted %d", got, want)
		}
	})

	t.Run("Testing for 25 Input", func(t *testing.T) {
		str := "i25e"
		got, _ := ParseInteger(str)
		want := 25

		if got != want {
			t.Errorf("Got %d Wanted %d", got, want)
		}
	})

	t.Run("Testing invalid input", func(t *testing.T) {
		str := "iabce"
		_, err := ParseInteger(str)

		if err == nil {
			t.Error("Error expected. Got nil")
		}

		errMsg := "integer parsing error: invalid format or value \"abc\""

		if err.Error() != errMsg {
			t.Errorf("Unexpected Error message. Wanted %q, got %q", errMsg, err.Error())
		}
	})
}

func TestParseString(t *testing.T) {
	t.Run("Testing for abc Input", func(t *testing.T) {
		str := "3:abc"
		got, _ := ParseString(str)
		want := "abc"

		if got != want {
			t.Errorf("Got %q Wanted %q", got, want)
		}
	})

	t.Run("Testing for abceft123 Input", func(t *testing.T) {
		str := "9:abceft123"
		got, _ := ParseString(str)
		want := "abceft123"

		if got != want {
			t.Errorf("Got %q Wanted %q", got, want)
		}
	})
}

func TestParseBencodedValue(t *testing.T) {
	t.Run("Testing integer parsing", func(t *testing.T) {
		str := "i42e"
		got, remaining, err := parseBencodedValue(str)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if remaining != "" {
			t.Errorf("Expected empty remaining, got %q", remaining)
		}
		want := int64(42)
		if got != want {
			t.Errorf("Got %v Wanted %v", got, want)
		}
	})

	t.Run("Testing string parsing", func(t *testing.T) {
		str := "4:spam"
		got, remaining, err := parseBencodedValue(str)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if remaining != "" {
			t.Errorf("Expected empty remaining, got %q", remaining)
		}
		want := "spam"
		if got != want {
			t.Errorf("Got %v Wanted %v", got, want)
		}
	})

	t.Run("Testing list parsing", func(t *testing.T) {
		str := "li1ei2ee"
		got, remaining, err := parseBencodedValue(str)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if remaining != "" {
			t.Errorf("Expected empty remaining, got %q", remaining)
		}
		want := []any{int64(1), int64(2)}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Got %v Wanted %v", got, want)
		}
	})

	t.Run("Testing empty dictionary parsing", func(t *testing.T) {
		str := "de"
		got, remaining, err := parseBencodedValue(str)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if remaining != "" {
			t.Errorf("Expected empty remaining, got %q", remaining)
		}
		want := map[string]interface{}{}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Got %v Wanted %v", got, want)
		}
	})
}

func TestParseList(t *testing.T) {
	t.Run("Testing valid list li1ei2ee", func(t *testing.T) {
		str := "li1ei2ee"
		got, err := ParseList(str)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		// Bencode integers are typically parsed into int64 in Go for consistency
		want := []any{int64(1), int64(2)}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("ParseList(%q) = %v, want %v", str, got, want)
		}
	})

	t.Run("Testing valid list l7:bencodei-20ee", func(t *testing.T) {
		str := "l7:bencodei-20ee"
		want := []any{"bencode", int64(-20)} // Expected valid parsed output

		got, err := ParseList(str)

		if err != nil {
			t.Errorf("Unexpected error for valid list %q: %v", str, err)
			return
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("ParseList(%q) = %v, want %v", str, got, want)
		}
	})

	t.Run("Testing nested list", func(t *testing.T) {
		str := "lli1ei2eeli3ei4eee"
		want := []any{
			[]any{int64(1), int64(2)},
			[]any{int64(3), int64(4)},
		}

		got, err := ParseList(str)

		if err != nil {
			t.Errorf("Unexpected error for nested list %q: %v", str, err)
			return
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("ParseList(%q) = %v, want %v", str, got, want)
		}
	})

	t.Run("Testing invalid input - not a list", func(t *testing.T) {
		str := "i42e"
		_, err := ParseList(str)
		if err == nil {
			t.Error("Expected error for non-list input, got nil")
		}
	})

	t.Run("Testing list with extra data", func(t *testing.T) {
		str := "li1ei2eextra"
		_, err := ParseList(str)
		if err == nil {
			t.Error("Expected error for list with extra data, got nil")
		}
	})
}

// BenchmarkParseInteger benchmarks integer parsing with different sizes
func BenchmarkParseInteger(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"Small", "i42e"},
		{"Medium", "i123456e"},
		{"Large", "i9223372036854775807e"},     // max int64
		{"Negative", "i-9223372036854775808e"}, // min int64
		{"Zero", "i0e"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := ParseInteger(tc.input)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkParseString benchmarks string parsing with different sizes
func BenchmarkParseString(b *testing.B) {
	// Generate test strings of different sizes
	small := "5:hello"
	medium := fmt.Sprintf("%d:%s", 100, strings.Repeat("a", 100))
	large := fmt.Sprintf("%d:%s", 10000, strings.Repeat("b", 10000))
	xlarge := fmt.Sprintf("%d:%s", 100000, strings.Repeat("c", 100000))

	testCases := []struct {
		name  string
		input string
	}{
		{"Small", small},
		{"Medium", medium},
		{"Large", large},
		{"XLarge", xlarge},
		{"Empty", "0:"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := ParseString(tc.input)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkParseList benchmarks list parsing with different sizes and nesting
func BenchmarkParseList(b *testing.B) {
	// Generate lists of different sizes
	small := "li1ei2ei3ee" // [1, 2, 3]

	// Medium list with 100 integers
	var mediumBuilder strings.Builder
	mediumBuilder.WriteString("l")
	for i := 0; i < 100; i++ {
		mediumBuilder.WriteString(fmt.Sprintf("i%de", i))
	}
	mediumBuilder.WriteString("e")
	medium := mediumBuilder.String()

	// Large list with 1000 integers
	var largeBuilder strings.Builder
	largeBuilder.WriteString("l")
	for i := 0; i < 1000; i++ {
		largeBuilder.WriteString(fmt.Sprintf("i%de", i))
	}
	largeBuilder.WriteString("e")
	large := largeBuilder.String()

	// Nested list
	nested := "lli1ei2eeli3ei4eeli5ei6eee" // [[1,2], [3,4], [5,6]]

	// Deep nested list
	deepNested := "llli1ei2eeli3ei4eeeli5ei6eeli7ei8eee" // [[[1,2], [3,4]], [[5,6], [7,8]]]

	// Mixed types list
	mixed := "li42e4:spam3:fooli1ei2eee" // [42, "spam", "foo", [1, 2]]

	testCases := []struct {
		name  string
		input string
	}{
		{"Small", small},
		{"Medium100", medium},
		{"Large1000", large},
		{"Nested", nested},
		{"DeepNested", deepNested},
		{"Mixed", mixed},
		{"Empty", "le"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := ParseList(tc.input)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkParseDictionary benchmarks dictionary parsing with different sizes and nesting
func BenchmarkParseDictionary(b *testing.B) {
	small := "d3:foo3:bare" // {"foo": "bar"}

	// Medium dictionary with 50 key-value pairs
	var mediumBuilder strings.Builder
	mediumBuilder.WriteString("d")
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		mediumBuilder.WriteString(fmt.Sprintf("%d:%s%d:%s", len(key), key, len(value), value))
	}
	mediumBuilder.WriteString("e")
	medium := mediumBuilder.String()

	// Large dictionary with 500 key-value pairs
	var largeBuilder strings.Builder
	largeBuilder.WriteString("d")
	for i := 0; i < 500; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		largeBuilder.WriteString(fmt.Sprintf("%d:%s%d:%s", len(key), key, len(value), value))
	}
	largeBuilder.WriteString("e")
	large := largeBuilder.String()

	// Nested dictionary
	nested := "d3:food3:bar3:bazee" // {"foo": {"bar": "baz"}}

	// Mixed types dictionary
	mixed := "d3:fooi42e3:bar4:spam4:listli1ei2eee" // {"foo": 42, "bar": "spam", "list": [1, 2]}

	testCases := []struct {
		name  string
		input string
	}{
		{"Small", small},
		{"Medium50", medium},
		{"Large500", large},
		{"Nested", nested},
		{"Mixed", mixed},
		{"Empty", "de"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := ParseDictionary(tc.input)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkParse benchmarks the general Parse function with various data types
func BenchmarkParse(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"Integer", "i42e"},
		{"String", "4:spam"},
		{"List", "li1ei2ei3ee"},
		{"Dictionary", "d3:foo3:bare"},
		{"ComplexNested", "d4:listli1ei2ee4:dictd3:foo3:bar7:integeri42eee"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := Parse(tc.input)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkRealWorldTorrent benchmarks parsing a realistic torrent-like structure
func BenchmarkRealWorldTorrent(b *testing.B) {
	// Simulate a simplified torrent file structure
	torrentData := "d8:announce8:test.com4:infod4:name4:test12:piece_lengthi262144e6:pieces20:xxxxxxxxxxxxxxxxxxxxee"

	b.Run("TorrentFile", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := Parse(torrentData)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkParseStringWithLargeLength benchmarks string parsing with large length values
func BenchmarkParseStringWithLargeLength(b *testing.B) {
	// Test parsing strings with large length prefixes
	testCases := []struct {
		name   string
		length int
	}{
		{"1KB", 1024},
		{"10KB", 10240},
		{"100KB", 102400},
	}

	for _, tc := range testCases {
		input := fmt.Sprintf("%d:%s", tc.length, strings.Repeat("x", tc.length))
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := ParseString(input)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkParseComplexStructure benchmarks parsing deeply nested and complex structures
func BenchmarkParseComplexStructure(b *testing.B) {
	// Create a complex nested structure
	var builder strings.Builder
	builder.WriteString("d")

	// Add multiple lists
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("list%d", i)
		builder.WriteString(fmt.Sprintf("%d:%sl", len(key), key))
		for j := 0; j < 20; j++ {
			builder.WriteString(fmt.Sprintf("i%de", j))
		}
		builder.WriteString("e")
	}

	// Add nested dictionaries
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("dict%d", i)
		builder.WriteString(fmt.Sprintf("%d:%sd", len(key), key))
		for j := 0; j < 10; j++ {
			subkey := fmt.Sprintf("key%d", j)
			subval := fmt.Sprintf("val%d", j)
			builder.WriteString(fmt.Sprintf("%d:%s%d:%s", len(subkey), subkey, len(subval), subval))
		}
		builder.WriteString("e")
	}

	builder.WriteString("e")
	complexStructure := builder.String()

	b.Run("ComplexStructure", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := Parse(complexStructure)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	// Large list that will cause significant allocations
	var builder strings.Builder
	builder.WriteString("l")
	for i := 0; i < 10000; i++ {
		str := fmt.Sprintf("item%d", i)
		builder.WriteString(fmt.Sprintf("%d:%s", len(str), str))
	}
	builder.WriteString("e")
	largeList := builder.String()

	b.Run("LargeList", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := Parse(largeList)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func TestParseDictionary(t *testing.T) {
	t.Run("Testing valid empty dictionary de", func(t *testing.T) {
		str := "de"
		got, remaining, err := parseBencodedValue(str)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if remaining != "" {
			t.Errorf("Expected empty remaining, got %q", remaining)
		}
		want := map[string]interface{}{}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Got %v Wanted %v", got, want)
		}
	})

	t.Run("Testing dictionary with string pair d3:foo3:bare", func(t *testing.T) {
		str := "d3:foo3:bare"
		got, remaining, err := parseBencodedValue(str)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if remaining != "" {
			t.Errorf("Expected empty remaining, got %q", remaining)
		}
		want := map[string]interface{}{"foo": "bar"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Got %v Wanted %v", got, want)
		}
	})

	t.Run("Testing dictionary with integer value d3:fooi123ee", func(t *testing.T) {
		str := "d3:fooi123ee"
		got, remaining, err := parseBencodedValue(str)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if remaining != "" {
			t.Errorf("Expected empty remaining, got %q", remaining)
		}
		want := map[string]interface{}{"foo": int64(123)}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Got %v Wanted %v", got, want)
		}
	})

	t.Run("Testing dictionary with list value d3:fooli1ei2eee", func(t *testing.T) {
		str := "d3:fooli1ei2eee"
		got, remaining, err := parseBencodedValue(str)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if remaining != "" {
			t.Errorf("Expected empty remaining, got %q", remaining)
		}
		want := map[string]interface{}{"foo": []interface{}{int64(1), int64(2)}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Got %v Wanted %v", got, want)
		}
	})

	t.Run("Testing nested dictionary d3:food3:bar3:bazee", func(t *testing.T) {
		str := "d3:food3:bar3:bazee"
		got, remaining, err := parseBencodedValue(str)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if remaining != "" {
			t.Errorf("Expected empty remaining, got %q", remaining)
		}
		want := map[string]interface{}{"foo": map[string]interface{}{"bar": "baz"}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Got %v Wanted %v", got, want)
		}
	})

	t.Run("Testing dictionary with multiple keys d3:fooi42e3:bar4:spame", func(t *testing.T) {
		str := "d3:fooi42e3:bar4:spame"
		got, remaining, err := parseBencodedValue(str)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if remaining != "" {
			t.Errorf("Expected empty remaining, got %q", remaining)
		}
		want := map[string]interface{}{"foo": int64(42), "bar": "spam"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Got %v Wanted %v", got, want)
		}
	})

	t.Run("Testing malformed dictionary missing end e", func(t *testing.T) {
		str := "d3:foo3:bar"
		_, _, err := parseBencodedValue(str)
		if err == nil {
			t.Error("Expected error for malformed dictionary, got nil")
		}
	})

	t.Run("Testing dictionary with non-string key di42e3:bare", func(t *testing.T) {
		str := "di42e3:bare"
		_, _, err := parseBencodedValue(str)
		if err == nil {
			t.Error("Expected error for non-string key, got nil")
		}
	})

	t.Run("Testing dictionary with odd number of elements d3:fooe", func(t *testing.T) {
		str := "d3:fooe"
		_, _, err := parseBencodedValue(str)
		if err == nil {
			t.Error("Expected error for odd number of elements, got nil")
		}
	})

	t.Run("Testing dictionary with extra data d3:foo3:bareextra", func(t *testing.T) {
		str := "d3:foo3:bareextra"
		got, remaining, err := parseBencodedValue(str)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if remaining != "extra" {
			t.Errorf("Expected remaining 'extra', got %q", remaining)
		}
		want := map[string]interface{}{"foo": "bar"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Got %v Wanted %v", got, want)
		}
	})
}
