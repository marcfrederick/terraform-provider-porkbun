package provider

import (
	"strconv"
)

// StringPtrToInt64Ptr converts a string pointer to an int64 pointer.
func StringPtrToInt64Ptr(s *string) *int64 {
	if s == nil {
		return nil
	}
	result, _ := strconv.ParseInt(*s, 10, 64)
	return &result
}

// Int64PtrToStringPtr converts an int64 pointer to a string pointer.
// Returns nil if the input is nil.
func Int64PtrToStringPtr(i *int64) *string {
	if i == nil {
		return nil
	}
	result := strconv.FormatInt(*i, 10)
	return &result
}
