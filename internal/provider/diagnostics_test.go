package provider

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/tuzzmaniandevil/porkbun-go"
)

func TestAddAPIDiagnostics_WithAPIError(t *testing.T) {
	var diags diag.Diagnostics

	apiErr := &porkbun.ErrorResponse{
		BaseResponse: porkbun.BaseResponse{
			HTTPResponse: &http.Response{StatusCode: http.StatusNotFound},
			Status:       "ERROR",
			Code:         porkbun.ErrCodeDomainNotFound,
			Message:      "Domain not found in your account.",
		},
	}

	addAPIDiagnostics(&diags, "Error Reading Domain", apiErr)

	if !diags.HasError() {
		t.Fatal("expected diagnostic error")
	}

	if diags.Errors()[0].Summary() != "Error Reading Domain" {
		t.Errorf("unexpected summary: %s", diags.Errors()[0].Summary())
	}

	detail := diags.Errors()[0].Detail()
	if detail != "Code: DOMAIN_NOT_FOUND\nMessage: Domain not found in your account." {
		t.Errorf("unexpected detail: %s", detail)
	}
}

func TestAddAPIDiagnostics_WithNextAction(t *testing.T) {
	var diags diag.Diagnostics

	apiErr := &porkbun.ErrorResponse{
		BaseResponse: porkbun.BaseResponse{
			HTTPResponse: &http.Response{StatusCode: http.StatusPaymentRequired},
			Status:       "ERROR",
			Code:         porkbun.ErrCodeInsufficientFunds,
			Message:      "Insufficient account balance.",
		},
		NextAction: &porkbun.NextAction{
			Type: porkbun.NextActionAddFunds,
			Hint: "Add account credit and retry.",
		},
	}

	addAPIDiagnostics(&diags, "Error Creating Domain", apiErr)

	if !diags.HasError() {
		t.Fatal("expected diagnostic error")
	}

	detail := diags.Errors()[0].Detail()
	expected := "Code: INSUFFICIENT_FUNDS\nMessage: Insufficient account balance.\nSuggested action: Add account credit and retry."
	if detail != expected {
		t.Errorf("unexpected detail:\ngot:  %s\nwant: %s", detail, expected)
	}
}

func TestAddAPIDiagnostics_WithWrappedAPIError(t *testing.T) {
	var diags diag.Diagnostics

	apiErr := &porkbun.ErrorResponse{
		BaseResponse: porkbun.BaseResponse{
			HTTPResponse: &http.Response{StatusCode: http.StatusForbidden},
			Status:       "ERROR",
			Code:         porkbun.ErrCodeIPNotAllowed,
			Message:      "Source IP is not in the key's allowlist.",
		},
	}

	// Simulate a wrapped error, as returned by helper functions like getDNSRecord.
	wrapped := fmt.Errorf("error fetching DNS record: %w", apiErr)

	addAPIDiagnostics(&diags, "Error Fetching DNS Record", wrapped)

	if !diags.HasError() {
		t.Fatal("expected diagnostic error")
	}

	detail := diags.Errors()[0].Detail()
	if detail != "Code: IP_NOT_ALLOWED\nMessage: Source IP is not in the key's allowlist." {
		t.Errorf("unexpected detail: %s", detail)
	}
}

func TestAddAPIDiagnostics_WithNonAPIError(t *testing.T) {
	var diags diag.Diagnostics

	err := errors.New("connection refused")

	addAPIDiagnostics(&diags, "Error Reading Domain", err)

	if !diags.HasError() {
		t.Fatal("expected diagnostic error")
	}

	if diags.Errors()[0].Summary() != "Error Reading Domain" {
		t.Errorf("unexpected summary: %s", diags.Errors()[0].Summary())
	}

	if diags.Errors()[0].Detail() != "connection refused" {
		t.Errorf("unexpected detail: %s", diags.Errors()[0].Detail())
	}
}
