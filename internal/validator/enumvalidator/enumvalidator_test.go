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

func TestEnumValidator_ValidateString(t *testing.T) {
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

func TestEnumValidator_ValidateInt64(t *testing.T) {
	t.Parallel()

	type testCase struct {
		in          types.Int64
		enumValues  []porkbun.DnssecAlgorithm
		expectError bool
	}

	testCases := map[string]testCase{
		"simple-match": {
			in: types.Int64Value(1),
			enumValues: []porkbun.DnssecAlgorithm{
				porkbun.DnssecAlgorithmRsaMd5,
				porkbun.DnssecAlgorithmDsaSha1,
			},
		},
		"simple-mismatch": {
			in: types.Int64Value(100),
			enumValues: []porkbun.DnssecAlgorithm{
				porkbun.DnssecAlgorithmRsaMd5,
				porkbun.DnssecAlgorithmDsaSha1,
			},
			expectError: true,
		},
		"skip-validation-on-null": {
			in: types.Int64Null(),
			enumValues: []porkbun.DnssecAlgorithm{
				porkbun.DnssecAlgorithmRsaMd5,
				porkbun.DnssecAlgorithmDsaSha1,
			},
		},
		"skip-validation-on-unknown": {
			in: types.Int64Unknown(),
			enumValues: []porkbun.DnssecAlgorithm{
				porkbun.DnssecAlgorithmRsaMd5,
				porkbun.DnssecAlgorithmDsaSha1,
			},
		},
	}

	for name, test := range testCases {
		t.Run(fmt.Sprintf("ValidateString - %s", name), func(t *testing.T) {
			t.Parallel()
			req := validator.Int64Request{
				ConfigValue: test.in,
			}
			res := validator.Int64Response{}
			enumvalidator.Valid(test.enumValues...).ValidateInt64(context.TODO(), req, &res)

			if !res.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if res.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", res.Diagnostics)
			}
		})
	}
}
