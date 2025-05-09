package provider_test

import (
	"reflect"
	"testing"

	"github.com/marcfrederick/terraform-provider-porkbun/internal/provider"
)

func TestStringPtrToInt64Ptr(t *testing.T) {
	tests := []struct {
		name string
		s    *string
		want *int64
	}{
		{"TestNil", nil, nil},
		{"TestStringEmpty", ptr(""), ptr(int64(0))},
		{"TestStringZero", ptr("0"), ptr(int64(0))},
		{"TestStringOne", ptr("1"), ptr(int64(1))},
		{"TestStringNegativeOne", ptr("-1"), ptr(int64(-1))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := provider.StringPtrToInt64Ptr(tt.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stringPtrToInt64Ptr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}

func TestInt64PtrToStringPtr(t *testing.T) {
	tests := []struct {
		name string
		i    *int64
		want *string
	}{
		{"TestNil", nil, nil},
		{"TestInt64Zero", ptr(int64(0)), ptr("0")},
		{"TestInt64One", ptr(int64(1)), ptr("1")},
		{"TestInt64NegativeOne", ptr(int64(-1)), ptr("-1")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := provider.Int64PtrToStringPtr(tt.i); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Int64PtrToStringPtr() = %v, want %v", got, tt.want)
			}
		})
	}
}
