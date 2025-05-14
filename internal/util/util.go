package util

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MustMapToList converts a slice of elements of type T to a types.List.
func MustMapToList[T any](elements []T, elementType attr.Type, f func(T) attr.Value) types.List {
	values := make([]attr.Value, 0, len(elements))
	for _, v := range elements {
		values = append(values, f(v))
	}
	return types.ListValueMust(elementType, values)
}

// BoolValue parses a string or bool value into a types.Bool.
func BoolValue[T ~bool | ~string](value T, diagnostics *diag.Diagnostics) types.Bool {
	switch v := any(value).(type) {
	case bool:
		return types.BoolValue(v)
	case string:
		return types.BoolValue(v == "yes")
	default:
		diagnostics.AddError("Invalid value", fmt.Sprintf("Value %q must be a valid bool", v))
		return types.BoolNull()
	}
}

// Int64PointerValue parses a string or bool pointer value into a types.Int64.1.
func Int64PointerValue[T ~int64 | ~string](value *T, diagnostics *diag.Diagnostics) types.Int64 {
	switch v := any(value).(type) {
	case nil:
		return types.Int64Null()
	case *int64:
		return Int64Value(*v, diagnostics)
	case *string:
		return Int64Value(*v, diagnostics)
	default:
		diagnostics.AddError("Invalid value", fmt.Sprintf("Value %q must be a valid *int64", v))
		return types.Int64Null()
	}
}

// Int64Value parses a string or int value into a types.Int64.
func Int64Value[T ~int64 | ~string](value T, diagnostics *diag.Diagnostics) types.Int64 {
	switch v := any(value).(type) {
	case int64:
		return types.Int64Value(v)
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return types.Int64Value(i)
		}
		diagnostics.AddError("Invalid value", fmt.Sprintf("Value %q must be a valid int64", v))
	default:
		diagnostics.AddError("Invalid value", fmt.Sprintf("Unsupported type: %T", value))
	}
	return types.Int64Null()
}
