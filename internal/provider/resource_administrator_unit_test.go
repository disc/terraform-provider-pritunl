package provider

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/lfventura/terraform-provider-pritunl/internal/pritunl"
)

// The schema of the resource has to hold together on its own: ConflictsWith
// pointing at an attribute that does not exist, or at a required one, is only
// caught here and not by the compiler.
func TestPritunlProviderSchema(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("the provider schema is invalid: %s", err)
	}
}

func TestAdministratorUsername(t *testing.T) {
	cases := []struct {
		configured string
		want       string
	}{
		{"tfacc-admin", "tfacc-admin"},
		{"TFACC-Admin", "tfacc-admin"},
		{"terraform admin", "terraformadmin"},
		{"admin@example.com", "admin@example.com"},
		{" spaced ", "spaced"},
		{"quoted\"name", "quotedname"},
		{"", ""},
	}

	for _, testCase := range cases {
		if got := administratorUsername(testCase.configured); got != testCase.want {
			t.Errorf("administratorUsername(%q) = %q, want %q", testCase.configured, got, testCase.want)
		}
	}
}

func TestAdministratorYubikeyId(t *testing.T) {
	cases := []struct {
		configured string
		want       string
	}{
		// a whole OTP, of which Pritunl keeps the public id only
		{"ccccccbcdefgtrljvnvtdjhbvrhlchgcbeghdthflukj", "ccccccbcdefg"},
		{"ccccccbcdefg", "ccccccbcdefg"},
		{" ccccccbcdefg ", "ccccccbcdefg"},
		{"cccccc", "cccccc"},
		{"", ""},
	}

	for _, testCase := range cases {
		if got := administratorYubikeyId(testCase.configured); got != testCase.want {
			t.Errorf("administratorYubikeyId(%q) = %q, want %q", testCase.configured, got, testCase.want)
		}
	}
}

// The guards Pritunl puts on the last super user of an instance cannot be
// tripped by the acceptance suite: they count the super users of the whole
// instance, and the account the suite authenticates as is one of them, so
// reaching them would mean demoting, disabling or deleting the very account
// every other test needs. What this provider owns of them is the diagnostic it
// builds out of the answer, and that is what is checked here, against the
// exact bodies the Pritunl backend sends.
func TestAdministratorDiagnostics(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		summary   string
		detail    string
		unchanged bool
	}{
		{
			name: "the last super user cannot be deleted",
			err: administratorApiError("deleting the administrator", 400, "no_admins",
				"At least one super administrator must exist.",
				`{"error": "no_admins", "error_msg": "At least one super administrator must exist."}`),
			summary: "At least one super administrator must exist.",
			detail:  "refuses to remove the last super user",
		},
		{
			name: "the last super user cannot be demoted",
			err: administratorApiError("updating the administrator", 400, "no_super_users",
				"There must be at least one super user.",
				`{"error": "no_super_users", "error_msg": "There must be at least one super user."}`),
			summary: "There must be at least one super user.",
			detail:  "refuses to take the super user flag off",
		},
		{
			name: "the last enabled super user cannot be disabled",
			err: administratorApiError("updating the administrator", 400, "no_admins_enabled",
				"At least one super administrator must be enabled.",
				`{"error": "no_admins_enabled", "error_msg": "At least one super administrator must be enabled."}`),
			summary: "At least one super administrator must be enabled.",
			detail:  "refuses to disable the last enabled super user",
		},
		{
			name: "the endpoint requires a super user",
			err: administratorApiError("getting the administrator", 400, "requires_super_user",
				"This administrator action can only be performed by a super user.",
				`{"error": "requires_super_user", "error_msg": "This administrator action can only be performed by a super user."}`),
			summary: "This administrator action can only be performed by a super user.",
			detail:  "requires a super user",
		},
		{
			name: "usernames are unique",
			err: administratorApiError("creating the administrator", 400, "admin_username_exists",
				"Administrator username already exists.",
				`{"error": "admin_username_exists", "error_msg": "Administrator username already exists."}`),
			summary: "Administrator username already exists.",
			detail:  "terraform import pritunl_administrator.example",
		},
		{
			name: "the two-step authentication modes are exclusive",
			err: administratorApiError("updating the administrator", 400, "admin_invalid_otp",
				"Cannot enable both local and authenticator two-step authentication.",
				`{"error": "admin_invalid_otp", "error_msg": "Cannot enable both local and authenticator two-step authentication."}`),
			summary: "Cannot enable both local and authenticator two-step authentication.",
			detail:  "never with both",
		},
		{
			name: "revoked credentials are explained",
			err:  administratorApiError("getting the administrator", 401, "", "", "401: Unauthorized"),
			// pritunl-web answers on its own here, with a body that carries no
			// error code at all
			summary: "The Pritunl API refused the administrator request",
			detail:  "the write revoked them",
		},
		{
			name: "an unknown refusal is reported as it is",
			err: administratorApiError("updating the administrator", 400, "something_else",
				"Something else went wrong.",
				`{"error": "something_else", "error_msg": "Something else went wrong."}`),
			summary:   "Non-200 response on updating the administrator",
			unchanged: true,
		},
		{
			name:      "an error that is not an api answer is reported as it is",
			err:       errors.New("UpdateAdministrator: Error on HTTP request: connection refused"),
			summary:   "UpdateAdministrator: Error on HTTP request: connection refused",
			unchanged: true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			diagnostics := administratorDiagnostics(testCase.err)

			if len(diagnostics) != 1 {
				t.Fatalf("got %d diagnostics, want 1", len(diagnostics))
			}

			diagnostic := diagnostics[0]

			if testCase.unchanged {
				// diag.FromErr puts the whole error in the summary and leaves
				// the detail empty, which is what an answer this provider has
				// nothing to add to has to keep doing
				if diagnostic.Detail != "" {
					t.Errorf("got the detail %q, want none", diagnostic.Detail)
				}

				if !strings.Contains(diagnostic.Summary, testCase.summary) {
					t.Errorf("got the summary %q, want it to contain %q", diagnostic.Summary, testCase.summary)
				}

				return
			}

			if diagnostic.Summary != testCase.summary {
				t.Errorf("got the summary %q, want %q", diagnostic.Summary, testCase.summary)
			}

			if !strings.Contains(diagnostic.Detail, testCase.detail) {
				t.Errorf("got the detail %q, want it to contain %q", diagnostic.Detail, testCase.detail)
			}
		})
	}
}

// administratorApiError rebuilds the error the client hands over for a request
// the API refused. The parsing of the body the Pritunl backend answers with is
// covered by TestNewApiError in the pritunl package, here the fields it yields
// are what matters.
func administratorApiError(operation string, statusCode int, code, message, body string) error {
	err := &pritunl.ApiError{
		Operation:  operation,
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Body:       body,
	}

	// the client hands the error over as it is, wrapping it here proves the
	// diagnostics look through a wrapped one as well
	return fmt.Errorf("the request failed: %w", err)
}
