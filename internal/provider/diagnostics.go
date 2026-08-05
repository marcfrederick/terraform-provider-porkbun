package provider

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/tuzzmaniandevil/porkbun-go"
)

// addAPIDiagnostics adds a diagnostic with enhanced error information from the
// Porkbun API. If the error is a *porkbun.ErrorResponse, it extracts the
// machine-readable error code, message, and remediation hint. Otherwise, it
// falls back to the error's string representation.
//
// This produces more actionable diagnostics than err.Error() alone:
//   - Code: machine-readable identifier (e.g., "INSUFFICIENT_FUNDS")
//   - Message: human-readable description from the API
//   - Hint: suggested remediation when available (e.g., "Add account credit and retry")
func addAPIDiagnostics(diags *diag.Diagnostics, summary string, err error) {
	var apiErr *porkbun.ErrorResponse
	if errors.As(err, &apiErr) {
		detail := fmt.Sprintf("Code: %s\nMessage: %s", apiErr.Code, apiErr.Message)
		if apiErr.NextAction != nil && apiErr.NextAction.Hint != "" {
			detail += fmt.Sprintf("\nSuggested action: %s", apiErr.NextAction.Hint)
		}
		diags.AddError(summary, detail)
		return
	}

	diags.AddError(summary, err.Error())
}
