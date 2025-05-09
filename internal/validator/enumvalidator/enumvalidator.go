package enumvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Enum is an enum type that can be used with the EnumValidator.
//
// It must be a string type and implement the IsValid() method.
type Enum interface {
	~string
}

type enumValidator struct {
	values []string
}

// Valid creates a new enum validator for the given enum type.
func Valid[T Enum](values ...T) validator.String {
	stringValues := make([]string, len(values))
	for i, v := range values {
		stringValues[i] = string(v)
	}
	return &enumValidator{values: stringValues}
}

func (v *enumValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v *enumValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("must be one of: %v", v.values)
}

func (v *enumValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	val := req.ConfigValue.ValueString()
	for _, allowedValue := range v.values {
		if val == allowedValue {
			return
		}
	}

	resp.Diagnostics.AddError(
		"Invalid value",
		fmt.Sprintf("Value %q must be one of: %v", val, v.values),
	)
}
