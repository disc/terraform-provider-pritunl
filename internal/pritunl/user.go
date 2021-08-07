package pritunl

type User struct {
	ID              string                   `json:"id,omitempty"`
	Name            string                   `json:"name"`
	Type            string                   `json:"type,omitempty"`
	AuthType        string                   `json:"auth_type,omitempty"`
	DnsServers      []string                 `json:"dns_servers,omitempty"`
	Pin             bool                     `json:"pin,omitempty"`
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
