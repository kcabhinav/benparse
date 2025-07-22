package encoder

import (
	"fmt"
	"strconv"
	"strings"
)

func EncodeInteger(value int) string {
	var builder strings.Builder
	builder.Grow(20) // Pre-allocate for typical integer size
	builder.WriteByte('i')
	builder.WriteString(strconv.Itoa(value))
	builder.WriteByte('e')
	return builder.String()
}

func EncodeString(value string) string {
	var builder strings.Builder
	lengthStr := strconv.Itoa(len(value))
	builder.Grow(len(lengthStr) + 1 + len(value)) // Pre-allocate exact size
	builder.WriteString(lengthStr)
	builder.WriteByte(':')
	builder.WriteString(value)
	return builder.String()
}

func EncodeList(values []interface{}) string {
	var builder strings.Builder
	builder.Grow(estimateListSize(values)) // Pre-allocate estimated size
	builder.WriteByte('l')

	for _, value := range values {
		switch v := value.(type) {
		case int:
			writeIntegerToBuilder(&builder, v)
		case string:
			writeStringToBuilder(&builder, v)
		case []any:
			builder.WriteString(EncodeList(v))
		default:
			panic(fmt.Sprintf("unsupported type: %T", v))
		}
	}

	builder.WriteByte('e')
	return builder.String()
}

func EncodeDictionary(values map[string]interface{}) string {
	var builder strings.Builder
	builder.Grow(estimateDictSize(values)) // Pre-allocate estimated size
	builder.WriteByte('d')

	for key, value := range values {
		writeStringToBuilder(&builder, key)
		switch v := value.(type) {
		case int:
			writeIntegerToBuilder(&builder, v)
		case string:
			writeStringToBuilder(&builder, v)
		case []any:
			builder.WriteString(EncodeList(v))
		default:
			panic(fmt.Sprintf("unsupported type: %T", v))
		}
	}

	builder.WriteByte('e')
	return builder.String()
}

// Helper functions for direct writing to builder (more efficient)
func writeIntegerToBuilder(builder *strings.Builder, value int) {
	builder.WriteByte('i')
	builder.WriteString(strconv.Itoa(value))
	builder.WriteByte('e')
}

func writeStringToBuilder(builder *strings.Builder, value string) {
	lengthStr := strconv.Itoa(len(value))
	builder.WriteString(lengthStr)
	builder.WriteByte(':')
	builder.WriteString(value)
}

// Estimation functions for better memory pre-allocation
func estimateListSize(values []interface{}) int {
	estimate := 2 // 'l' and 'e'
	for _, value := range values {
		switch v := value.(type) {
		case int:
			estimate += 20 // Conservative estimate for integer encoding
		case string:
			estimate += len(strconv.Itoa(len(v))) + 1 + len(v) // length:string
		case []any:
			estimate += estimateListSize(v) // Recursive estimation
		}
	}
	return estimate
}

func estimateDictSize(values map[string]interface{}) int {
	estimate := 2 // 'd' and 'e'
	for key, value := range values {
		// Key size
		estimate += len(strconv.Itoa(len(key))) + 1 + len(key)

		// Value size
		switch v := value.(type) {
		case int:
			estimate += 20 // Conservative estimate for integer encoding
		case string:
			estimate += len(strconv.Itoa(len(v))) + 1 + len(v)
		case []any:
			estimate += estimateListSize(v)
		}
	}
	return estimate
}
