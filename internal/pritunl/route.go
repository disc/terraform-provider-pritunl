package pritunl

import "strconv"

type Route struct {
	ID             string `json:"id,omitempty"`
	Network        string `json:"network"`
	Nat            bool   `json:"nat"`
	Comment        string `json:"comment,omitempty"`
	Server         string `json:"server,omitempty"`
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

func ConvertMapToRoute(data map[string]interface{}) Route {
	var route Route
	if v, ok := data["id"]; ok {
		route.ID = v.(string)
	}
	if v, ok := data["network"]; ok {
		route.Network = v.(string)
	}
	if v, ok := data["comment"]; ok {
		route.Comment = v.(string)
	}
	if v, ok := data["nat"]; ok {
		boolVal, _ := strconv.ParseBool(v.(string))
		route.Nat = boolVal
	}

	return route
}
