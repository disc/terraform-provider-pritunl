package pritunl

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ErrAdministratorNotFound is returned by GetAdministrator when the instance
// holds no administrator with the given id, so that a resource whose
// administrator was removed outside of Terraform can be told apart from an
// instance that is merely unreachable.
var ErrAdministratorNotFound = errors.New("the administrator does not exist")

// Administrator is one administrator account of a Pritunl instance, the login
// of its web console and of its API, kept as decoded JSON instead of a typed
// struct.
//
// It is the only type of this package besides Settings that is not a plain Go
// struct, and for the very same reason: PUT /admin/<id> is a full replace and
// not a partial update. Organization, Server and User are ordinary structs
// because their handlers take a partial body, this one cannot.
//
// The Pritunl web console talks to pritunl-web, which deserialises the request
// body into a fixed struct of plain values, no pointers and no omitempty, and
// re-serialises it before handing the request over to the Pritunl backend.
// Every field left out of the request therefore reaches the backend carrying
// its zero value, and the backend, whose handler only checks whether a key is
// present, applies it. A request that only means to rename an administrator
// consequently strips its super user flag, turns its API access off and, as a
// side effect of that, rotates the API credentials it authenticates with. That
// is not a theoretical hazard, it is what a PUT carrying nothing but a
// username does to an administrator on a stock instance.
//
// The console lives with that by reading the whole administrator back and
// sending all of it again on every save, and so does this package. Keeping the
// object raw is what makes it possible: a field this provider does not model
// is never inspected and never converted, it is handed back exactly as the API
// returned it, which also covers the fields a newer Pritunl adds to the
// endpoint. Numbers are decoded as json.Number so that they are written back
// with the very same representation.
type Administrator map[string]interface{}

// String returns the value of a string field, or an empty string when the
// field is absent, null or not a string.
func (a Administrator) String(key string) string {
	value, _ := a[key].(string)

	return value
}

// Bool returns the value of a boolean field, or false when the field is
// absent, null or not a boolean. Pritunl reports the flags of an administrator
// that has never had them written as null rather than as false, which carries
// the same meaning to it.
func (a Administrator) Bool(key string) bool {
	value, _ := a[key].(bool)

	return value
}

// Has reports whether the API returned the field at all, which is how the
// fields a Pritunl version does not know about are told apart from the ones it
// reports as null. local_otp_auth is the field this matters for: instances
// older than its introduction leave it out of the response entirely.
func (a Administrator) Has(key string) bool {
	_, ok := a[key]

	return ok
}

// The fields GET /admin returns that describe the administrator instead of
// configuring it. None of them is read by the create or the update handler,
// and default is decided by Pritunl alone: the flag marks the administrator
// the instance was seeded with, and Pritunl clears it by itself the first time
// that administrator changes its password.
var administratorReadOnlyFields = []string{
	"id",
	"default",
	// only the list endpoint reports it, it mirrors the auditing setting of
	// the instance rather than anything about the administrator
	"audit",
}

// normalize makes an administrator read from the API acceptable as a request
// body again.
//
// It only ever clears fields that this package never sends a value for, which
// is what keeps it from undoing an overlay: the fields it touches are the ones
// the API returns but the write path must not hand back.
//
//   - token and secret are triggers rather than values. The backend regenerates
//     the credential as soon as the key carries anything truthy and throws the
//     value away, so handing the credentials that were just read back would
//     rotate the API credentials of the administrator on every single write.
//
//   - otp_secret is a trigger too, and the one field whose type in the response
//     differs from the one the request accepts: the response carries the
//     two-step authentication secret itself, a string, as soon as the
//     administrator has one, while pritunl-web deserialises the request body
//     into a struct that declares it as a boolean and answers an empty 400 for
//     anything else. False is both the type the request wants and the value
//     that leaves the secret alone, the backend only regenerates it for a
//     literal true.
//
// password is deliberately left alone: GET never returns it, so there is
// nothing to strip, and clearing it here would drop the password an update is
// meant to apply.
func (a Administrator) normalize() {
	for _, field := range administratorReadOnlyFields {
		delete(a, field)
	}

	a["token"] = ""
	a["secret"] = ""
	a["otp_secret"] = false
}

// ApiError is a request the Pritunl API refused, carrying the reason the API
// itself gave for it. Pritunl answers a refused administrator request with an
// error code and a message rather than with a bare status, and those are what
// tell a lockout guard apart from a malformed request, so they are kept
// instead of being flattened into a string.
type ApiError struct {
	// what was being done, phrased to read after "Non-200 response on"
	Operation  string
	StatusCode int
	// the machine readable error of the API, such as no_admins
	Code string
	// the message the API pairs with the code
	Message string
	Body    string
}

func (e *ApiError) Error() string {
	if e.Code == "" {
		return fmt.Sprintf("Non-200 response on %s\ncode=%d\nbody=%s", e.Operation, e.StatusCode, e.Body)
	}

	return fmt.Sprintf("Non-200 response on %s\ncode=%d\nerror=%s\nerror_msg=%s",
		e.Operation, e.StatusCode, e.Code, e.Message)
}

// newApiError builds the error of a refused request, parsing the error code
// and message out of the body when the API sent one. A body it cannot parse is
// kept as it is: pritunl-web answers a request it could not deserialise with
// an empty 400 of its own, which carries no code at all.
func newApiError(operation string, statusCode int, body []byte) *ApiError {
	apiError := &ApiError{
		Operation:  operation,
		StatusCode: statusCode,
		Body:       string(body),
	}

	var parsed struct {
		Error    string `json:"error"`
		ErrorMsg string `json:"error_msg"`
	}

	if err := json.Unmarshal(body, &parsed); err == nil {
		apiError.Code = parsed.Error
		apiError.Message = parsed.ErrorMsg
	}

	return apiError
}
