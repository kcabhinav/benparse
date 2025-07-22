package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseInteger parses bencode integers with reduced string operations
func ParseInteger(str string) (int, error) {
	if len(str) < 3 || str[0] != 'i' || str[len(str)-1] != 'e' {
		return 0, fmt.Errorf("integer parsing error: invalid format %q", str)
	}

	// Parse directly without string trimming operations
	numStr := str[1 : len(str)-1] // Remove 'i' and 'e' without allocations

	// Check for leading zeros before calling strconv.Atoi
	if (len(numStr) > 1 && numStr[0] == '0') || (len(numStr) > 2 && numStr[0] == '-' && numStr[1] == '0') {
		return 0, fmt.Errorf("integer parsing error: malformed integer with leading zero %q", numStr)
	}

	res, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("integer parsing error: invalid format or value %q", numStr)
	}

	return res, nil
}

// ParseString parses bencode strings with optimized operations
func ParseString(str string) (string, error) {
	colonIndex := strings.IndexByte(str, ':') // Use IndexByte instead of Index
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

// parseBencodedValue is the core optimized parsing function with pre-allocation
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
		// Pre-allocate slice with reasonable capacity to reduce reallocations
		list := make([]any, 0, 16)

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
		// Pre-allocate map with reasonable capacity to reduce hash table resizing
		dict := make(map[string]any, 8)

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

// ParseList parses bencode lists with pre-allocation optimizations
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

// ParseDictionary parses bencode dictionaries with pre-allocation optimizations
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

// Parse is the main optimized parsing function
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

// EstimateCapacity provides heuristic-based capacity estimation for better pre-allocation
func EstimateCapacity(s string) (listCap, dictCap int) {
	// Simple heuristics based on string analysis
	listMarkers := strings.Count(s, "l")
	dictMarkers := strings.Count(s, "d")
	colonCount := strings.Count(s, ":")
	integerCount := strings.Count(s, "i")

	// Estimate list capacity based on markers and content
	estimatedElements := colonCount + integerCount
	if listMarkers > 0 {
		listCap = max(16, estimatedElements/listMarkers)
	} else {
		listCap = 16
	}

	// Estimate dictionary capacity based on key-value pairs
	if dictMarkers > 0 {
		dictCap = max(8, colonCount/(dictMarkers*2)) // Assuming string keys
	} else {
		dictCap = 8
	}

	return listCap, dictCap
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ParseWithEstimation uses capacity estimation for even better performance
func ParseWithEstimation(str string) (any, error) {
	// For very large inputs, use estimation
	if len(str) > 1000 {
		listCap, dictCap := EstimateCapacity(str)
		return parseWithCapacities(str, listCap, dictCap)
	}

	// For smaller inputs, use standard optimized parsing
	return Parse(str)
}

// parseWithCapacities uses estimated capacities for pre-allocation
func parseWithCapacities(s string, listCap, dictCap int) (any, error) {
	val, remaining, err := parseBencodedValueWithCapacities(s, listCap, dictCap)
	if err != nil {
		return nil, err
	}

	if len(remaining) > 0 {
		return nil, fmt.Errorf("extra data %q after value %q", remaining, s)
	}

	return val, nil
}

// parseBencodedValueWithCapacities uses custom capacities for pre-allocation
func parseBencodedValueWithCapacities(s string, listCap, dictCap int) (any, string, error) {
	if len(s) == 0 {
		return nil, "", fmt.Errorf("empty string for parsing Bencode value")
	}

	switch s[0] {
	case 'i':
		eIndex := strings.IndexByte(s, 'e')
		if eIndex == -1 {
			return nil, "", fmt.Errorf("integer parsing error: missing 'e' in %q", s)
		}
		if eIndex == 1 && (s[1] == 'e' || s[1] == '-') {
			return nil, "", fmt.Errorf("integer parsing error: malformed integer %q", s[:eIndex+1])
		}
		intStr := s[:eIndex+1]
		val, err := ParseInteger(intStr)
		if err != nil {
			return nil, "", err
		}
		return int64(val), s[eIndex+1:], nil

	case 'l':
		current := s[1:] // Skip 'l'
		// Use estimated capacity
		list := make([]any, 0, listCap)

		for len(current) > 0 && current[0] != 'e' {
			val, remaining, err := parseBencodedValueWithCapacities(current, listCap, dictCap)
			if err != nil {
				return nil, "", err
			}
			list = append(list, val)
			current = remaining
		}

		if len(current) == 0 || current[0] != 'e' {
			return nil, "", fmt.Errorf("list parsing error: missing 'e' at end of list elements in %q", s)
		}

		return list, current[1:], nil

	case 'd':
		current := s[1:] // Skip 'd'
		// Use estimated capacity
		dict := make(map[string]any, dictCap)

		for len(current) > 0 && current[0] != 'e' {
			// Parse key (must be a string)
			key, remaining, err := parseBencodedValueWithCapacities(current, listCap, dictCap)
			if err != nil {
				return nil, "", err
			}
			keyStr, ok := key.(string)
			if !ok {
				return nil, "", fmt.Errorf("dictionary key must be a string, got %T in %q", key, s)
			}
			current = remaining

			if len(current) == 0 {
				return nil, "", fmt.Errorf("dictionary missing value for key %q in %q", keyStr, s)
			}

			// Parse value
			value, remaining, err := parseBencodedValueWithCapacities(current, listCap, dictCap)
			if err != nil {
				return nil, "", err
			}
			dict[keyStr] = value
			current = remaining
		}

		if len(current) == 0 || current[0] != 'e' {
			return nil, "", fmt.Errorf("dictionary parsing error: missing 'e' at end of dictionary in %q", s)
		}

		return dict, current[1:], nil

	default: // String parsing
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
