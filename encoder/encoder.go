package encoder

import (
	"fmt"
)

func EncodeInteger(value int) string {
	return fmt.Sprintf("i%de", value)
}

func EncodeString(value string) string {
	return fmt.Sprintf("%d:%s", len(value), value)
}

func EncodeList(values []interface{}) string {
	var result string
	for _, value := range values {
		switch v := value.(type) {
		case int:
			result += EncodeInteger(v)
		case string:
			result += EncodeString(v)
		case []any:
			result += EncodeList(v)
		default:
			panic(fmt.Sprintf("unsupported type: %T", v))
		}
	}
	return fmt.Sprintf("l%se", result)
}

func EncodeDictionary(values map[string]interface{}) string {
	var result string
	for key, value := range values {
		result += EncodeString(key)
		switch v := value.(type) {
		case int:
			result += EncodeInteger(v)
		case string:
			result += EncodeString(v)
		case []any:
			result += EncodeList(v)
		default:
			panic(fmt.Sprintf("unsupported type: %T", v))
		}
	}
	return fmt.Sprintf("d%se", result)
}
