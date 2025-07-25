package utils

import (
	"encoding/json"
	"strings"
)

// ConvertKeysToCamelCase recursively converts all keys in a data structure to camelCase
func ConvertKeysToCamelCase(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			camelKey := ToCamelCase(key)
			result[camelKey] = ConvertKeysToCamelCase(value)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = ConvertKeysToCamelCase(item)
		}
		return result
	default:
		return v
	}
}

// ToCamelCase converts snake_case to camelCase
func ToCamelCase(s string) string {
	// Handle special cases
	if s == "" {
		return s
	}

	// Split by underscore
	parts := strings.Split(s, "_")
	if len(parts) == 1 {
		return s // No underscores, return as is
	}

	// First part stays lowercase, capitalize first letter of subsequent parts
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}

	return result
}

// ConvertStructToMap converts a struct to map[string]interface{} and then applies camelCase conversion
func ConvertStructToMap(v interface{}) (interface{}, error) {
	// Convert struct to JSON and back to map[string]interface{} to handle all types
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return v, err
	}

	var result interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return v, err
	}

	// Now apply camelCase conversion
	return ConvertKeysToCamelCase(result), nil
}
