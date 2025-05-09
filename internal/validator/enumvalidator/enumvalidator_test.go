package enumvalidator_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tuzzmaniandevil/porkbun-go"

	"github.com/marcfrederick/terraform-provider-porkbun/internal/validator/enumvalidator"
)

func TestEnumValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		in          types.String
		enumValues  []porkbun.ForwardType
		expectError bool
	}

	testCases := map[string]testCase{
		"simple-match": {
			in: types.StringValue("temporary"),
			enumValues: []porkbun.ForwardType{
				porkbun.Temporary,
				porkbun.Permanent,
			},
		},
		"simple-mismatch": {
			in: types.StringValue("301"),
			enumValues: []porkbun.ForwardType{
				porkbun.Temporary,
				porkbun.Permanent,
			},
			expectError: true,
		},
		"skip-validation-on-null": {
			in: types.StringNull(),
			enumValues: []porkbun.ForwardType{
				porkbun.Temporary,
				porkbun.Permanent,
			},
		},
		"skip-validation-on-unknown": {
			in: types.StringUnknown(),
			enumValues: []porkbun.ForwardType{
				porkbun.Temporary,
				porkbun.Permanent,
			},
		},
	}

	for name, test := range testCases {
		t.Run(fmt.Sprintf("ValidateString - %s", name), func(t *testing.T) {
			t.Parallel()
			req := validator.StringRequest{
				ConfigValue: test.in,
			}
			res := validator.StringResponse{}
			enumvalidator.Valid(test.enumValues...).ValidateString(context.TODO(), req, &res)

			if !res.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if res.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", res.Diagnostics)
			}
		})
	}
}
