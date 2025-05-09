package enumvalidator

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.String = (*enumValidator)(nil)
	_ validator.Int64  = (*enumValidator)(nil)
)

type EnumValidator interface {
	validator.String
	validator.Int64
}

type enumValidator struct {
	allowedValues []string
}

// Valid creates a new EnumValidator to ensure the value matches one of the allowed options.
//
// The provided allowedValues must be of type string or convertible to string.
// The Validate* methods will verify if the string representation of the value
// is within the allowed options.
func Valid[T ~string](allowedValues ...T) EnumValidator {
	stringAllowedValues := make([]string, len(allowedValues))
	for i, v := range allowedValues {
		stringAllowedValues[i] = string(v)
	}
	return &enumValidator{allowedValues: stringAllowedValues}
}

func (v *enumValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v *enumValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("must be one of: %v", v.allowedValues)
}

func (v *enumValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	val := req.ConfigValue.ValueString()
	if !v.isValidValue(val) {
		resp.Diagnostics.AddError(
			"Invalid value",
			fmt.Sprintf("Value %q must be one of: %v", val, v.allowedValues),
		)
	}
}

func (v *enumValidator) ValidateInt64(_ context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	val := strconv.FormatInt(req.ConfigValue.ValueInt64(), 10)
	if !v.isValidValue(val) {
		resp.Diagnostics.AddError(
			"Invalid value",
			fmt.Sprintf("Value %q must be one of: %v", val, v.allowedValues),
		)
	}
}

// isValidValue checks if the given value is in the list of allowed allowedValues.
func (v *enumValidator) isValidValue(value string) bool {
	for _, allowedValue := range v.allowedValues {
		if value == allowedValue {
			return true
		}
	}
	return false
}
