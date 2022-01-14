package pritunl

import (
	"encoding/json"
	"errors"
)

type User struct {
	ID              string                   `json:"id,omitempty"`
	Name            string                   `json:"name"`
	Type            string                   `json:"type,omitempty"`
	AuthType        string                   `json:"auth_type,omitempty"`
	DnsServers      []string                 `json:"dns_servers,omitempty"`
	Pin             Pin                      `json:"pin,omitempty"`
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
	SSO             string                   `json:"sso,omitempty"`
	BypassSecondary bool                     `json:"bypass_secondary,omitempty"`
	Groups          []string                 `json:"groups,omitempty"`
	Audit           bool                     `json:"audit,omitempty"`
	Gravatar        bool                     `json:"gravatar,omitempty"`
	OtpAuth         bool                     `json:"otp_auth,omitempty"`
	Organization    string                   `json:"organization,omitempty"`
}

type PortForwarding struct {
	Dport    string `json:"dport"`
	Protocol string `json:"protocol"`
	Port     string `json:"port"`
}

// Pin encodes a pin code which can be set or not set, as well as known or unknown.
type Pin struct {
	// IsSet defines whenever the pin code is set. It must be true if Value is non-empty.
	IsSet bool
	// Value defines whenever the pin code is known - nonempty string is a known pin.
	Value string
}

func NewPin(value string) Pin {
	return Pin{
		IsSet: value != "",
		Value: value,
	}
}

func (pin *Pin) MarshalJSON() ([]byte, error) {
	if pin.Value != "" {
		if !pin.IsSet {
			return nil, errors.New("invalid pin: Value is set but IsSet is false")
		}
		return json.Marshal(pin.Value)
	}
	return json.Marshal(pin.IsSet)
}

func (pin *Pin) UnmarshalJSON(data []byte) error {
	var i interface{}
	err := json.Unmarshal(data, &i)
	if err != nil {
		return err
	}
	switch v := i.(type) {
	case bool:
		pin.IsSet = v
		pin.Value = ""
	case string:
		pin.IsSet = true
		pin.Value = v
	default:
		return errors.New("invalid type for the pin, only bool or string is accepted")
	}
	return nil
}
