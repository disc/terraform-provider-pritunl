package pritunl

import (
	"encoding/json"
)

type User struct {
	ID              string                   `json:"id,omitempty"`
	Name            string                   `json:"name"`
	Type            string                   `json:"type,omitempty"`
	AuthType        string                   `json:"auth_type,omitempty"`
	DnsServers      []string                 `json:"dns_servers,omitempty"`
	DnsSuffix       string                   `json:"dns_suffix,omitempty"`
	DnsMapping      string                   `json:"dns_mapping,omitempty"`
	Disabled        bool                     `json:"disabled,omitempty"`
	NetworkLinks    []string                 `json:"network_links,omitempty"`
	PortForwarding  []map[string]interface{} `json:"port_forwarding,omitempty"`
	Email           string                   `json:"email,omitempty"`
	Status          bool                     `json:"status,omitempty"`
	OtpSecret       string                   `json:"otp_secret,omitempty"`
	ClientToClient  bool                     `json:"client_to_client,omitempty"`
	MacAddresses    []string                 `json:"mac_addresses,omitempty"`
	YubicoID        string                   `json:"yubico_id,omitempty"`
	SSO             interface{}              `json:"sso,omitempty"`
	BypassSecondary bool                     `json:"bypass_secondary,omitempty"`
	Groups          []string                 `json:"groups,omitempty"`
	Audit           bool                     `json:"audit,omitempty"`
	Gravatar        bool                     `json:"gravatar,omitempty"`
	OtpAuth         bool                     `json:"otp_auth,omitempty"`
	DeviceAuth      bool                     `json:"device_auth,omitempty"`
	Organization    string                   `json:"organization,omitempty"`
	Pin             *Pin                      `json:"pin,omitempty"`
}

type PortForwarding struct {
	Dport    string `json:"dport"`
	Protocol string `json:"protocol"`
	Port     string `json:"port"`
}

type Pin struct {
	IsSet  bool
	Secret string
}

// MarshalJSON customizes the JSON encoding of the Pin struct.
//
// When marshaling a User JSON, the "pin" field will contain the PIN secret
// if it is set, otherwise the field is excluded. This is used when making
// a user create or update request to the Pritunl API.
func (p *Pin) MarshalJSON() ([]byte, error) {
	if p.Secret != "" {
		return json.Marshal(p.Secret)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON customizes the JSON decoding of the Pin struct.
//
// When unmarshaling a User JSON, the "pin" field will contain a boolean
// indicating whether the user has a PIN set or not. This is used when
// reading a user response from the Pritunl API.
func (p *Pin) UnmarshalJSON(data []byte) error {
	var b bool
	err := json.Unmarshal(data, &b)
	if err == nil {
		p.IsSet = b
		p.Secret = ""
	}
	return err
}
