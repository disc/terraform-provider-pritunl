package pritunl

import (
	"encoding/json"
)

// Settings is the whole settings object of a Pritunl instance, kept as decoded
// JSON instead of a typed struct.
//
// PUT /settings is a full replace, not a partial update. The official web
// console talks to pritunl-web, which deserialises the request body into a
// fixed struct of plain values (no pointers, no omitempty) and re-serialises it
// before handing the request over to the Pritunl backend. Every setting left
// out of the request therefore reaches the backend carrying an empty value,
// and the backend, whose handler only checks whether a key is present, clears
// it. The console lives with that by reading the whole settings object and
// sending all of it back on every save, and so does this provider.
//
// Keeping the object raw is what makes that possible: an unmanaged setting is
// never modelled, never inspected and never converted, it is handed back
// exactly as the API returned it. Numbers are decoded as json.Number so that
// they are written back with the very same representation.
type Settings map[string]interface{}

// String returns the value of a string setting, or an empty string when the
// setting is absent or null.
func (s Settings) String(key string) string {
	value, _ := s[key].(string)

	return value
}

// Bool returns the value of a boolean setting, or false when the setting is
// absent, null or not a boolean.
func (s Settings) Bool(key string) bool {
	value, _ := s[key].(bool)

	return value
}

// Int returns the value of a numeric setting, or zero when the setting is
// absent, null or not a number.
func (s Settings) Int(key string) int {
	switch value := s[key].(type) {
	case json.Number:
		number, err := value.Int64()
		if err != nil {
			return 0
		}

		return int(number)
	case float64:
		return int(value)
	}

	return 0
}

// normalize makes a settings object read from the API acceptable as a request
// body again.
//
// Pritunl ships exactly one setting whose factory default has a different type
// than the one pritunl-web accepts: sso holds the name of the single sign-on
// provider, a string, but its default value is the boolean false, and GET
// /settings keeps returning that boolean until single sign-on has been
// configured once. pritunl-web deserialises the request body into a struct that
// declares sso as a string, so handing the boolean back makes it reject the
// whole request with an empty 400 before the Pritunl backend ever sees it.
//
// An empty string carries the same meaning without the type clash: the backend
// reads any falsy value as "single sign-on disabled". Only false is rewritten,
// a boolean true never occurs and is better surfaced as an error than silently
// turned into a disabled provider.
func (s Settings) normalize() {
	if value, ok := s["sso"].(bool); ok && !value {
		s["sso"] = ""
	}
}
