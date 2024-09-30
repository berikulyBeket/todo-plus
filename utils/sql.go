package utils

import (
	"fmt"
	"strings"
)

// CreatePlaceholders generates $1, $2, $3,... placeholders for SQL queries
func CreatePlaceholders(count int) string {
	placeholders := make([]string, count)
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	return strings.Join(placeholders, ", ")
}

// ConvertToInterfaceSlice converts a slice of ints to a slice of empty interfaces
func ConvertToInterfaceSlice(ids []int) []interface{} {
	result := make([]interface{}, len(ids))
	for i, id := range ids {
		result[i] = id
	}
	return result
}
