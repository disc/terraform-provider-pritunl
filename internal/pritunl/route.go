package pritunl

import (
	"encoding/hex"
)

type Route struct {
	Network        string `json:"network"`
	Nat            bool   `json:"nat"`
	Comment        string `json:"comment,omitempty"`
	VirtualNetwork bool   `json:"virtual_network,omitempty"`
	WgNetwork      string `json:"wg_network,omitempty"`
	NetworkLink    bool   `json:"network_link,omitempty"`
	ServerLink     bool   `json:"server_link,omitempty"`
	NetGateway     bool   `json:"net_gateway,omitempty"`
	VpcID          string `json:"vpc_id,omitempty"`
	VpcRegion      string `json:"vpc_region,omitempty"`
	Metric         string `json:"metric,omitempty"`
	Advertise      bool   `json:"advertise,omitempty"`
	NatInterface   string `json:"nat_interface,omitempty"`
	NatNetmap      string `json:"nat_netmap,omitempty"`
}

func (r Route) GetID() string {
	if len(r.Network) > 0 {
		return hex.EncodeToString([]byte(r.Network))
	}

	return ""
}

func ConvertMapToRoute(data map[string]interface{}) Route {
	var route Route

	if v, ok := data["network"]; ok {
		route.Network = v.(string)
	}
	if v, ok := data["comment"]; ok {
		route.Comment = v.(string)
	}
	if v, ok := data["nat"]; ok {
		route.Nat = v.(bool)
	}

	return route
}
