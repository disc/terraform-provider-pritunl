package pritunl

import (
	"encoding/json"
	"strconv"
)

const (
	ServerStatusOnline  = "online"
	ServerStatusOffline = "offline"

	ServerNetworkModeTunnel = "tunnel"
	ServerNetworkModeBridge = "bridge"
)

type Server struct {
	ID               string   `json:"id,omitempty"`
	Name             string   `json:"name"`
	Protocol         string   `json:"protocol,omitempty"`
	Cipher           string   `json:"cipher,omitempty"`
	Hash             string   `json:"hash,omitempty"`
	Port             int      `json:"port,omitempty"`
	Network          string   `json:"network,omitempty"`
	WG               bool     `json:"wg,omitempty"`
	PortWG           int      `json:"port_wg,omitempty"`
	NetworkWG        string   `json:"network_wg,omitempty"`
	NetworkMode      string   `json:"network_mode,omitempty"`
	NetworkStart     string   `json:"network_start,omitempty"`
	NetworkEnd       string   `json:"network_end,omitempty"`
	RestrictRoutes   bool     `json:"restrict_routes,omitempty"`
	IPv6             bool     `json:"ipv6,omitempty"`
	IPv6Firewall     bool     `json:"ipv6_firewall,omitempty"`
	BindAddress      string   `json:"bind_address,omitempty"`
	DhParamBits      int      `json:"dh_param_bits,omitempty"`
	Groups           []string `json:"groups,omitempty"`
	MultiDevice      bool     `json:"multi_device,omitempty"`
	DnsServers       []string `json:"dns_servers,omitempty"`
	SearchDomain     string   `json:"search_domain,omitempty"`
	InterClient      bool     `json:"inter_client,omitempty"`
	PingInterval     int      `json:"ping_interval,omitempty"`
	PingTimeout      int      `json:"ping_timeout,omitempty"`
	LinkPingInterval int      `json:"link_ping_interval,omitempty"`
	LinkPingTimeout  int      `json:"link_ping_timeout,omitempty"`
	InactiveTimeout  int      `json:"inactive_timeout,omitempty"`
	SessionTimeout   int      `json:"session_timeout,omitempty"`
	AllowedDevices   string   `json:"allowed_devices,omitempty"`
	MaxClients       int      `json:"max_clients,omitempty"`
	MaxDevices       int      `json:"max_devices,omitempty"`
	ReplicaCount     int      `json:"replica_count,omitempty"`
	VxLan            bool     `json:"vxlan,omitempty"`
	DnsMapping       bool     `json:"dns_mapping,omitempty"`
	PreConnectMsg    string   `json:"pre_connect_msg,omitempty"`
	OtpAuth          bool     `json:"otp_auth,omitempty"`
	MssFix           int      `json:"mss_fix,omitempty"`
	LzoCompression   bool     `json:"lzo_compression,omitempty"`
	BlockOutsideDns  bool     `json:"block_outside_dns,omitempty"`
	JumboFrames      bool     `json:"jumbo_frames,omitempty"`
	Debug            bool     `json:"debug,omitempty"`
	Status           string   `json:"status,omitempty"`
}

func (s *Server) MarshalJSON() ([]byte, error) {
	type Alias Server
	return json.Marshal(&struct {
		// Pritunl API expects input mss_fix value as a string, but returns as an int
		MssFix string `json:"mss_fix"`
		*Alias
	}{
		MssFix: strconv.Itoa(s.MssFix),
		Alias:  (*Alias)(s),
	})
}
