package provider

import (
	"strings"
	"testing"
)

func TestNormalizeSAMLCert(t *testing.T) {
	naked := "MIIDqDCCApCgAwIBAgIGAZkxg7q2MA0GCSqGSIb3DQEBCwUAMIGUMQswCQYDVQQGEwJVUzETMBEGA1UECAwKQ2FsaWZvcm5pYTEWMBQGA1UEBwwNU2FuIEZyYW5jaXNjbw"
	armored := "-----BEGIN CERTIFICATE-----\n" + naked[:64] + "\n" + naked[64:128] + "\n" + naked[128:] + "\n-----END CERTIFICATE-----"

	cases := map[string]struct {
		in   string
		want string
	}{
		"empty stays empty":              {"", ""},
		"whitespace only stays empty":    {" \n\t ", ""},
		"naked base64 gets armored":      {naked, armored},
		"metadata line breaks are fine":  {naked[:40] + "\n" + naked[40:80] + " \n " + naked[80:], armored},
		"armored passes through trimmed": {"\n" + armored + "\n", armored},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got := normalizeSAMLCert(c.in)
			if got != c.want {
				t.Fatalf("normalizeSAMLCert(%q):\n got %q\nwant %q", c.in, got, c.want)
			}
			if again := normalizeSAMLCert(got); again != got {
				t.Fatalf("not idempotent:\nfirst %q\n then %q", got, again)
			}
			for i, line := range strings.Split(got, "\n") {
				if len(line) > 64 {
					t.Fatalf("line %d longer than 64 columns: %q", i+1, line)
				}
			}
		})
	}
}
