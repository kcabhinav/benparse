package parser

import (
	"reflect"
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
