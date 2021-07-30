package pritunl

// Server
// {
//    "port_wg": null,
//    "dns_servers": [
//        "8.8.8.8"
//    ],
//    "protocol": "tcp",
//    "max_devices": 0,
//    "max_clients": 2000,
//    "link_ping_timeout": 5,
//    "ping_timeout": 60,
//    "ipv6": false,
//    "vxlan": true,
//    "network_mode": "tunnel",
//    "bind_address": "",
//    "block_outside_dns": false,
//    "network_start": "",
//    "name": "Alice-TCPnoTLS",
//    "ping_interval": 10,
//    "allowed_devices": null,
//    "users_online": 1,
//    "ipv6_firewall": true,
//    "session_timeout": null,
//    "otp_auth": false,
//    "multi_device": false,
//    "search_domain": null,
//    "lzo_compression": "adaptive",
//    "pre_connect_msg": null,
//    "inactive_timeout": null,
//    "link_ping_interval": 1,
//    "id": "60d06624c36cc9d1d673304b",
//    "ping_timeout_wg": 360,
//    "uptime": 1295821,
//    "network_end": "",
//    "network": "192.168.249.0/24",
//    "dh_param_bits": 2048,
//    "wg": false,
//    "port": 17490,
//    "devices_online": 1,
//    "network_wg": null,
//    "status": "online",
//    "dns_mapping": false,
//    "hash": "sha1",
//    "debug": false,
//    "restrict_routes": true,
//    "user_count": 1,
//    "groups": [],
//    "inter_client": true,
//    "replica_count": 1,
//    "cipher": "aes128",
//    "mss_fix": null,
//    "jumbo_frames": false
//}
//*/
type Server struct {
	ID               string   `json:"id,omitempty"`
	Name             string   `json:"name"`
	Protocol         string   `json:"protocol"`
	Cipher           string   `json:"cipher"`
	Hash             string   `json:"hash"`
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
}
