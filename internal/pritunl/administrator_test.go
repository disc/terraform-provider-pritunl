package pritunl

import (
	"encoding/json"
	"testing"
)

// The body of an administrator read from a Pritunl instance, with the two-step
// authentication turned on so that otp_secret carries the secret itself, which
// is the shape the round trip has to survive.
const administratorResponse = `{
	"id": "6a9b8b734f5041e3afd7a11b",
	"username": "tfacc-admin",
	"yubikey_id": null,
	"otp_auth": true,
	"otp_secret": "NGN6M5AQX6KDANHR",
	"local_otp_auth": false,
	"auth_api": true,
	"token": "ST2c3wTMiDprQzpH6cwsZpkUHUjjreD5",
	"secret": "nngx1cMRXmkip3ip3Y5cHT1zFmRZhTU3",
	"default": false,
	"disabled": false,
	"super_user": true,
	"audit": false
}`

func decodeAdministrator(t *testing.T, body string) Administrator {
	t.Helper()

	administrator := Administrator{}
	if err := decodeAdministrators([]byte(body), &administrator); err != nil {
		t.Fatalf("failed to decode the administrator: %s", err)
	}

	return administrator
}

func TestAdministratorAccessors(t *testing.T) {
	administrator := decodeAdministrator(t, administratorResponse)

	if got := administrator.String("username"); got != "tfacc-admin" {
		t.Errorf("String(username) = %q, want %q", got, "tfacc-admin")
	}

	// null is how Pritunl reports a field an account never had written, and it
	// means the same thing to it as the empty value
	if got := administrator.String("yubikey_id"); got != "" {
		t.Errorf("String(yubikey_id) = %q, want an empty string", got)
	}

	if got := administrator.Bool("super_user"); !got {
		t.Errorf("Bool(super_user) = %v, want true", got)
	}

	if got := administrator.Bool("nonexistent"); got {
		t.Errorf("Bool(nonexistent) = %v, want false", got)
	}

	if !administrator.Has("local_otp_auth") {
		t.Error("Has(local_otp_auth) = false, want true")
	}

	// what tells a field a Pritunl version does not know about apart from one
	// it reports as null
	if administrator.Has("nonexistent") {
		t.Error("Has(nonexistent) = true, want false")
	}

	if !administrator.Has("yubikey_id") {
		t.Error("Has(yubikey_id) = false for a field reported as null, want true")
	}
}

// normalize is what keeps the full object round trip from doing damage of its
// own, so what it clears and what it leaves alone are both part of the
// contract.
func TestAdministratorNormalize(t *testing.T) {
	administrator := decodeAdministrator(t, administratorResponse)

	// the overlay of the resource runs before the request is built, and the
	// password it puts there has to survive normalize, unlike everything else
	// the write path must not send
	administrator["password"] = "tfacc-Passw0rd!"
	administrator["some_field_a_newer_pritunl_added"] = "kept"

	administrator.normalize()

	// handing the credentials that were just read back would make Pritunl
	// generate a new pair on every write, the value of the field is a trigger
	// and not a value
	if got := administrator["token"]; got != "" {
		t.Errorf("token = %v after normalize, want an empty string", got)
	}

	if got := administrator["secret"]; got != "" {
		t.Errorf("secret = %v after normalize, want an empty string", got)
	}

	// the response carries the two-step authentication secret itself, a
	// string, while the request wants a boolean and the Pritunl web frontend
	// answers anything else with an empty 400
	if got := administrator["otp_secret"]; got != false {
		t.Errorf("otp_secret = %v after normalize, want false", got)
	}

	for _, field := range administratorReadOnlyFields {
		if administrator.Has(field) {
			t.Errorf("%s is still in the request body after normalize", field)
		}
	}

	if got := administrator["password"]; got != "tfacc-Passw0rd!" {
		t.Errorf("password = %v after normalize, want the one the overlay set", got)
	}

	// the whole point of keeping the object raw: a field this provider knows
	// nothing about is handed back exactly as it was read
	if got := administrator["some_field_a_newer_pritunl_added"]; got != "kept" {
		t.Errorf("an unmodelled field = %v after normalize, want it kept", got)
	}

	// and the managed ones are handed back too, which is what keeps a write
	// that only renames the account from stripping everything else off it
	for field, want := range map[string]interface{}{
		"username":       "tfacc-admin",
		"super_user":     true,
		"auth_api":       true,
		"disabled":       false,
		"otp_auth":       true,
		"local_otp_auth": false,
	} {
		if got := administrator[field]; got != want {
			t.Errorf("%s = %v after normalize, want %v", field, got, want)
		}
	}

	// the request body still has to be serialisable, a value the Pritunl web
	// frontend cannot deserialise is answered with an empty 400
	if _, err := json.Marshal(administrator); err != nil {
		t.Errorf("the normalized administrator cannot be marshalled: %s", err)
	}
}

func TestNewApiError(t *testing.T) {
	err := newApiError("deleting the administrator", 400,
		[]byte(`{"error": "no_admins", "error_msg": "At least one super administrator must exist."}`))

	if err.Code != "no_admins" {
		t.Errorf("Code = %q, want %q", err.Code, "no_admins")
	}

	if err.Message != "At least one super administrator must exist." {
		t.Errorf("Message = %q, want the message of the api", err.Message)
	}

	if got := err.Error(); got != "Non-200 response on deleting the administrator\ncode=400\n"+
		"error=no_admins\nerror_msg=At least one super administrator must exist." {
		t.Errorf("Error() = %q", got)
	}

	// pritunl-web answers a request it could not deserialise with an empty 400
	// of its own, which carries no code to report
	empty := newApiError("updating the administrator", 400, []byte(""))

	if empty.Code != "" {
		t.Errorf("Code = %q for a body that is not an api error, want an empty string", empty.Code)
	}

	if got := empty.Error(); got != "Non-200 response on updating the administrator\ncode=400\nbody=" {
		t.Errorf("Error() = %q", got)
	}
}
