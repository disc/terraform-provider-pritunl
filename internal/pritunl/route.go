package pritunl

// {"comment": null, "vpc_region": null, "metric": null, "advertise": false, "nat_interface": null, "id": "382e382e382e322f3332", "nat_netmap": null, "network": "8.8.8.2/32", "server": "610332d44bce2ca96a757523", "nat": true, "vpc_id": null, "net_gateway": false}

type Route struct {
	ID           string `json:"id,omitempty"`
	Server       string `json:"server"`
	Network      string `json:"network"`
	NetGateway   string `json:"net_gateway"`
	Comment      string `json:"comment"`
	VpcID        string `json:"vpc_id"`
	VpcRegion    string `json:"vpc_region"`
	Metric       string `json:"metric"`
	Advertise    string `json:"advertise"`
	Nat          string `json:"nat"`
	NatInterface string `json:"nat_interface"`
	NatNetmap    string `json:"nat_netmap"`
}
