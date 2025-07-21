package parser

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseInteger(str string) (int, error) {
	str = strings.TrimPrefix(str, "i")
	str = strings.TrimSuffix(str, "e")

	res, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("integer parsing error: invalid format or value %q", str)
	}

	if (len(str) > 1 && str[0] == '0') || (len(str) > 2 && str[0] == '-' && str[1] == '0') {
		return 0, fmt.Errorf("integer parsing error: malformed integer with leading zero %q", str)
	}

	return res, nil
}

func ParseString(str string) (string, error) {
	colonIndex := strings.Index(str, ":")
	if colonIndex == -1 {
		return "", fmt.Errorf("string parsing error: missing colon in %q", str)
	}

	lengthStr := str[:colonIndex]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", fmt.Errorf("string parsing error: invalid length %q in %q", lengthStr, str)
	}
	if length < 0 {
		return "", fmt.Errorf("string parsing error: negative length %d in %q", length, str)
	}

	// Check for leading zeros in length (e.g., "05:hello" is invalid)
	if len(lengthStr) > 1 && lengthStr[0] == '0' {
		return "", fmt.Errorf("string parsing error: malformed length with leading zero %q in %q", lengthStr, str)
	}

	stringValueStartIndex := colonIndex + 1
	bencodedStringFullLength := stringValueStartIndex + length

	if len(str) < bencodedStringFullLength {
		return "", fmt.Errorf("string parsing error: declared length %d exceeds actual string length in %q", length, str)
	}

	stringValue := str[stringValueStartIndex : stringValueStartIndex+length]

	if len(str) > bencodedStringFullLength {
		return "", fmt.Errorf("string parsing error: extra data %q after declared string length in %q", str[bencodedStringFullLength:], str)
	}

	return stringValue, nil
}

func parseBencodedValue(s string) (any, string, error) {
	if len(s) == 0 {
		return nil, "", fmt.Errorf("empty string for parsing Bencode value")
	}

	switch s[0] {
	case 'i':
		eIndex := strings.IndexByte(s, 'e')
		if eIndex == -1 {
			return nil, "", fmt.Errorf("integer parsing error: missing 'e' in %q", s)
		}
		// Basic check for malformed 'i' (like "ie" or "i-e" without digits)
		if eIndex == 1 && (s[1] == 'e' || s[1] == '-') {
			return nil, "", fmt.Errorf("integer parsing error: malformed integer %q", s[:eIndex+1])
		}
		intStr := s[:eIndex+1]
		val, err := ParseInteger(intStr)
		if err != nil {
			return nil, "", err
		}
		return int64(val), s[eIndex+1:], nil // Convert int to int64 for consistency with bencode specs
	case 'l':
		current := s[1:] // Skip 'l'
		list := []any{}

		for len(current) > 0 && current[0] != 'e' {
			val, remaining, err := parseBencodedValue(current)
			if err != nil {
				return nil, "", err
			}
			list = append(list, val)
			current = remaining
		}

		if len(current) == 0 || current[0] != 'e' {
			return nil, "", fmt.Errorf("list parsing error: missing 'e' at end of list elements in %q", s)
		}

		return list, current[1:], nil // Return the list and string after 'e'
	case 'd':
		current := s[1:] // Skip 'd'
		dict := make(map[string]any)

		for len(current) > 0 && current[0] != 'e' {
			// Parse key (must be a string)
			key, remaining, err := parseBencodedValue(current)
			if err != nil {
				return nil, "", err
			}
			keyStr, ok := key.(string)
			if !ok {
				return nil, "", fmt.Errorf("dictionary key must be a string, got %T in %q", key, s)
			}
			current = remaining

			// Check if we have a value
			if len(current) == 0 {
				return nil, "", fmt.Errorf("dictionary missing value for key %q in %q", keyStr, s)
			}

			// Parse value
			value, remaining, err := parseBencodedValue(current)
			if err != nil {
				return nil, "", err
			}
			dict[keyStr] = value
			current = remaining
		}

		if len(current) == 0 || current[0] != 'e' {
			return nil, "", fmt.Errorf("dictionary parsing error: missing 'e' at end of dictionary in %q", s)
		}

		return dict, current[1:], nil // Return the dict and string after 'e'
	default: // Must be a string (starts with a digit)
		colonIndex := strings.IndexByte(s, ':')
		if colonIndex == -1 {
			return nil, "", fmt.Errorf("string parsing error: missing colon in %q", s)
		}
		lengthStr := s[:colonIndex]
		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return nil, "", fmt.Errorf("string parsing error: invalid length %q in %q", lengthStr, s)
		}
		if length < 0 {
			return nil, "", fmt.Errorf("string parsing error: negative length %d in %q", length, s)
		}

		stringValueStartIndex := colonIndex + 1
		stringValueEndIndex := stringValueStartIndex + length

		if stringValueEndIndex > len(s) {
			return nil, "", fmt.Errorf("string parsing error: declared length %d (%q) exceeds actual string length in %q", length, lengthStr, s)
		}

		val := s[stringValueStartIndex:stringValueEndIndex]
		return val, s[stringValueEndIndex:], nil
	}
}

func ParseList(str string) ([]any, error) {
	val, remaining, err := parseBencodedValue(str)
	if err != nil {
		return nil, err
	}

	listVal, ok := val.([]any)
	if !ok {
		return nil, fmt.Errorf("input %q was not a bencoded list, parsed as %T", str, val)
	}

	if len(remaining) > 0 {
		return nil, fmt.Errorf("extra data %q after list %q", remaining, str)
	}

	return listVal, nil
}

func ParseDictionary(str string) (map[string]any, error) {
	val, remaining, err := parseBencodedValue(str)
	if err != nil {
		return nil, err
	}

	dictVal, ok := val.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("input %q was not a bencoded dictionary, parsed as %T", str, val)
	}

	if len(remaining) > 0 {
		return nil, fmt.Errorf("extra data %q after dictionary %q", remaining, str)
	}

	return dictVal, nil
}

func Parse(str string) (any, error) {
	val, remaining, err := parseBencodedValue(str)
	if err != nil {
		return nil, err
	}

	if len(remaining) > 0 {
		return nil, fmt.Errorf("extra data %q after value %q", remaining, str)
	}

	return val, nil
}
