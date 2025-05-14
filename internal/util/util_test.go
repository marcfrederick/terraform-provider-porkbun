package util_test

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/marcfrederick/terraform-provider-porkbun/internal/util"
)

func TestMustMapToList(t *testing.T) {
	type args[T any] struct {
		elements    []T
		elementType attr.Type
		f           func(T) attr.Value
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want types.List
	}
	tests := []testCase[string]{
		{
			name: "empty",
			args: args[string]{
				elements:    []string{},
				elementType: types.StringType,
				f:           func(s string) attr.Value { return types.StringValue(s) },
			},
			want: types.ListValueMust(types.StringType, []attr.Value{}),
		},
		{
			name: "non-empty",
			args: args[string]{
				elements:    []string{"a", "b", "c"},
				elementType: types.StringType,
				f:           func(s string) attr.Value { return types.StringValue(s) },
			},
			want: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("a"),
				types.StringValue("b"),
				types.StringValue("c"),
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := util.MustMapToList(tt.args.elements, tt.args.elementType, tt.args.f); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MustMapToList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoolValue(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		want     types.Bool
		wantDiag diag.Diagnostics
	}{
		{
			name:  "bool true",
			value: true,
			want:  types.BoolValue(true),
		},
		{
			name:  "bool false",
			value: false,
			want:  types.BoolValue(false),
		},
		{
			name:  "string yes",
			value: "yes",
			want:  types.BoolValue(true),
		},
		{
			name:  "string no",
			value: "no",
			want:  types.BoolValue(false),
		},
		{
			name:  "string empty",
			value: "",
			want:  types.BoolValue(false),
		},
		{
			name:  "string invalid",
			value: "invalid",
			want:  types.BoolValue(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				got     types.Bool
				gotDiag diag.Diagnostics
			)

			switch v := tt.value.(type) {
			case bool:
				got = util.BoolValue(v, &gotDiag)
			case string:
				got = util.BoolValue(v, &gotDiag)
			default:
				t.Fatalf("unsupported type: %T", v)
			}

			if got != tt.want {
				t.Errorf("BoolValue() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(gotDiag, tt.wantDiag) {
				t.Errorf("BoolValue() diag = %v, want %v", gotDiag, tt.wantDiag)
			}
		})
	}
}

func TestInt64Value(t *testing.T) {
	type testCase struct {
		name     string
		value    any
		want     types.Int64
		wantDiag diag.Diagnostics
	}
	tests := []testCase{
		{
			name:  "int64",
			value: int64(42),
			want:  types.Int64Value(42),
		},
		{
			name:  "string",
			value: "42",
			want:  types.Int64Value(42),
		},
		{
			name:  "string invalid",
			value: "invalid",
			want:  types.Int64Null(),
			wantDiag: diag.Diagnostics{
				diag.NewErrorDiagnostic("Invalid value", `Value "invalid" must be a valid int64`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				got     types.Int64
				gotDiag diag.Diagnostics
			)

			switch v := tt.value.(type) {
			case int64:
				got = util.Int64Value(v, &gotDiag)
			case string:
				got = util.Int64Value(v, &gotDiag)
			default:
				t.Fatalf("unsupported type: %T", v)
			}

			if got != tt.want {
				t.Errorf("Int64Value() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(gotDiag, tt.wantDiag) {
				t.Errorf("Int64Value() diag = %v, want %v", gotDiag, tt.wantDiag)
			}
		})
	}
}
