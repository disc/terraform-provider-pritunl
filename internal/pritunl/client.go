package pritunl

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client interface {
	TestApiCall() error

	GetOrganizations() ([]Organization, error)
	GetOrganization(id string) (*Organization, error)
	CreateOrganization(name string) (*Organization, error)
	UpdateOrganization(id string, organization *Organization) error
	DeleteOrganization(name string) error

	GetUser(id string, orgId string) (*User, error)
	CreateUser(newUser User) (*User, error)
	UpdateUser(id string, user *User) error
	DeleteUser(id string, orgId string) error

	GetServers() ([]Server, error)
	GetServer(id string) (*Server, error)
	CreateServer(serverData map[string]interface{}) (*Server, error)
	UpdateServer(id string, server *Server) error
	DeleteServer(id string) error

	GetOrganizationsByServer(serverId string) ([]Organization, error)
	AttachOrganizationToServer(organizationId, serverId string) error
	DetachOrganizationFromServer(organizationId, serverId string) error

	GetRoutesByServer(serverId string) ([]Route, error)
	AddRouteToServer(serverId string, route Route) error
	AddRoutesToServer(serverId string, route []Route) error
	DeleteRouteFromServer(serverId string, route Route) error
	UpdateRouteOnServer(serverId string, route Route) error

	GetHosts() ([]Host, error)
	GetHostsByServer(serverId string) ([]Host, error)
	AttachHostToServer(hostId, serverId string) error
	DetachHostFromServer(hostId, serverId string) error

	StartServer(serverId string) error
	StopServer(serverId string) error

	GetSettings() (*Settings, error)
	UpdateSettings(settings *Settings) error
}

type client struct {
	httpClient *http.Client
	baseUrl    string
}

func (c client) TestApiCall() error {
	url := fmt.Sprintf("/state")
	req, err := http.NewRequest("GET", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("TestApiCall: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response on the tests api call\nbody=%s", body)
	}

	// 401 - invalid credentials
	if resp.StatusCode == 401 {
		return fmt.Errorf("unauthorized: Invalid token or secret")
	}

	return nil
}

func (c client) GetOrganization(id string) (*Organization, error) {
	url := fmt.Sprintf("/organization/%s", id)
	req, err := http.NewRequest("GET", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetOrganization: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Non-200 response on getting the organization\nbody=%s", body)
	}

	var organization Organization

	err = json.Unmarshal(body, &organization)
	if err != nil {
		return nil, fmt.Errorf("GetOrganization: %s: %+v, id=%s, body=%s", err, organization, id, body)
	}

	return &organization, nil
}

func (c client) GetOrganizations() ([]Organization, error) {
	url := fmt.Sprintf("/organization")
	req, err := http.NewRequest("GET", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetOrganization: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Non-200 response on getting the organization\nbody=%s", body)
	}

	var organizations []Organization

	err = json.Unmarshal(body, &organizations)
	if err != nil {
		return nil, fmt.Errorf("GetOrganization: %s: %+v, body=%s", err, organizations, body)
	}

	return organizations, nil
}

func (c client) CreateOrganization(name string) (*Organization, error) {
	var jsonStr = []byte(`{"name": "` + name + `"}`)

	url := "/organization"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CreateOrganization: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Non-200 response on creating the organization\nbody=%s", body)
	}

	var organization Organization
	err = json.Unmarshal(body, &organization)
	if err != nil {
		return nil, fmt.Errorf("CreateOrganization: %s: %+v, name=%s, body=%s", err, organization, name, body)
	}

	return &organization, nil
}

func (c client) UpdateOrganization(id string, organization *Organization) error {
	jsonData, err := json.Marshal(organization)
	if err != nil {
		return fmt.Errorf("UpdateOrganization: Error on marshalling data: %s", err)
	}

	url := fmt.Sprintf("/organization/%s", id)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("UpdateOrganization: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response on updating the organization\nbody=%s", body)
	}

	return nil
}

func (c client) DeleteOrganization(id string) error {
	url := fmt.Sprintf("/organization/%s", id)
	req, err := http.NewRequest("DELETE", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("DeleteOrganization: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response on deleting the organization\nbody=%s", body)
	}

	return nil
}

func (c client) GetServer(id string) (*Server, error) {
	url := fmt.Sprintf("/server/%s", id)
	req, err := http.NewRequest("GET", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Non-200 response on getting the server\nbody=%s", body)
	}

	var server Server
	err = json.Unmarshal(body, &server)

	if err != nil {
		return nil, fmt.Errorf("GetServer: %s: %+v, id=%s, body=%s", err, server, id, body)
	}

	return &server, nil
}

func (c client) GetServers() ([]Server, error) {
	url := fmt.Sprintf("/server")
	req, err := http.NewRequest("GET", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetServers: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Non-200 response on getting servers\nbody=%s", body)
	}

	var servers []Server
	err = json.Unmarshal(body, &servers)

	if err != nil {
		return nil, fmt.Errorf("GetServers: %s: %+v, body=%s", err, servers, body)
	}

	return servers, nil
}

func (c client) CreateServer(serverData map[string]interface{}) (*Server, error) {
	serverStruct := Server{}

	if v, ok := serverData["name"]; ok {
		serverStruct.Name = v.(string)
	}
	if v, ok := serverData["protocol"]; ok {
		serverStruct.Protocol = v.(string)
	}
	if v, ok := serverData["cipher"]; ok {
		serverStruct.Cipher = v.(string)
	}
	if v, ok := serverData["network"]; ok {
		serverStruct.Network = v.(string)
	}
	if v, ok := serverData["hash"]; ok {
		serverStruct.Hash = v.(string)
	}
	if v, ok := serverData["port"]; ok {
		serverStruct.Port = v.(int)
	}
	if v, ok := serverData["bind_address"]; ok {
		serverStruct.BindAddress = v.(string)
	}
	if v, ok := serverData["groups"]; ok {
		groups := make([]string, 0)
		for _, group := range v.([]interface{}) {
			groups = append(groups, group.(string))
		}
		serverStruct.Groups = groups
	}
	if v, ok := serverData["dns_servers"]; ok {
		dnsServers := make([]string, 0)
		for _, dns := range v.([]interface{}) {
			dnsServers = append(dnsServers, dns.(string))
		}
		serverStruct.DnsServers = dnsServers
	}
	if v, ok := serverData["network_wg"]; ok {
		serverStruct.NetworkWG = v.(string)
	}
	if v, ok := serverData["port_wg"]; ok {
		serverStruct.PortWG = v.(int)
	}

	isWgEnabled := serverStruct.NetworkWG != "" && serverStruct.PortWG > 0
	serverStruct.WG = isWgEnabled

	if v, ok := serverData["otp_auth"]; ok {
		serverStruct.OtpAuth = v.(bool)
	}

	if v, ok := serverData["ipv6"]; ok {
		serverStruct.IPv6 = v.(bool)
	}

	if v, ok := serverData["dh_param_bits"]; ok {
		serverStruct.DhParamBits = v.(int)
	}

	if v, ok := serverData["ping_interval"]; ok {
		serverStruct.PingInterval = v.(int)
	}

	if v, ok := serverData["ping_timeout"]; ok {
		serverStruct.PingTimeout = v.(int)
	}

	if v, ok := serverData["link_ping_interval"]; ok {
		serverStruct.LinkPingInterval = v.(int)
	}

	if v, ok := serverData["link_ping_timeout"]; ok {
		serverStruct.LinkPingTimeout = v.(int)
	}

	if v, ok := serverData["inactive_timeout"]; ok {
		serverStruct.InactiveTimeout = v.(int)
	}

	if v, ok := serverData["max_clients"]; ok {
		serverStruct.MaxClients = v.(int)
	}

	if v, ok := serverData["network_mode"]; ok {
		serverStruct.NetworkMode = v.(string)
	}

	if v, ok := serverData["network_start"]; ok {
		serverStruct.NetworkStart = v.(string)
	}

	if v, ok := serverData["network_end"]; ok {
		serverStruct.NetworkEnd = v.(string)
	}

	if serverStruct.NetworkMode == ServerNetworkModeBridge && (serverStruct.NetworkStart == "" || serverStruct.NetworkEnd == "") {
		return nil, fmt.Errorf("the attribute network_mode = %s requires network_start and network_end attributes", ServerNetworkModeBridge)
	}

	if v, ok := serverData["mss_fix"]; ok {
		serverStruct.MssFix = v.(int)
	}

	if v, ok := serverData["max_devices"]; ok {
		serverStruct.MaxDevices = v.(int)
	}

	if v, ok := serverData["pre_connect_msg"]; ok {
		serverStruct.PreConnectMsg = v.(string)
	}

	if v, ok := serverData["allowed_devices"]; ok {
		serverStruct.AllowedDevices = v.(string)
	}

	if v, ok := serverData["search_domain"]; ok {
		serverStruct.SearchDomain = v.(string)
	}

	if v, ok := serverData["replica_count"]; ok {
		serverStruct.ReplicaCount = v.(int)
	}

	if v, ok := serverData["multi_device"]; ok {
		serverStruct.MultiDevice = v.(bool)
	}

	if v, ok := serverData["debug"]; ok {
		serverStruct.Debug = v.(bool)
	}

	if v, ok := serverData["restrict_routes"]; ok {
		serverStruct.RestrictRoutes = v.(bool)
	}

	if v, ok := serverData["block_outside_dns"]; ok {
		serverStruct.BlockOutsideDns = v.(bool)
	}

	if v, ok := serverData["dns_mapping"]; ok {
		serverStruct.DnsMapping = v.(bool)
	}

	if v, ok := serverData["inter_client"]; ok {
		serverStruct.InterClient = v.(bool)
	}

	if v, ok := serverData["vxlan"]; ok {
		serverStruct.VxLan = v.(bool)
	}

	jsonData, err := serverStruct.MarshalJSON()

	url := "/server"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CreateServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Non-200 response on creating the server\ncode=%d\nbody=%s", resp.StatusCode, body)
	}

	var server Server
	err = json.Unmarshal(body, &server)
	if err != nil {
		return nil, fmt.Errorf("CreateServer: Error on unmarshalling http response: %s", err)
	}

	return &server, nil
}

func (c client) UpdateServer(id string, server *Server) error {
	jsonData, err := server.MarshalJSON()
	if err != nil {
		return fmt.Errorf("UpdateServer: Error on marshalling data: %s", err)
	}

	url := fmt.Sprintf("/server/%s", id)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("UpdateServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response on updating the server\nbody=%s", body)
	}

	return nil
}

func (c client) DeleteServer(id string) error {
	url := fmt.Sprintf("/server/%s", id)
	req, err := http.NewRequest("DELETE", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("DeleteServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response on deleting the server\nbody=%s", body)
	}

	return nil
}

func (c client) GetOrganizationsByServer(serverId string) ([]Organization, error) {
	url := fmt.Sprintf("/server/%s/organization", serverId)
	req, err := http.NewRequest("GET", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GeteOrganizationsByServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Non-200 response on getting organizations on the server\nbody=%s", body)
	}

	var organizations []Organization
	json.Unmarshal(body, &organizations)

	return organizations, nil
}

func (c client) AttachOrganizationToServer(organizationId, serverId string) error {
	url := fmt.Sprintf("/server/%s/organization/%s", serverId, organizationId)
	req, err := http.NewRequest("PUT", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("AttachOrganizationToServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response on attaching an organization the server\nbody=%s", body)
	}

	return nil
}

func (c client) DetachOrganizationFromServer(organizationId, serverId string) error {
	url := fmt.Sprintf("/server/%s/organization/%s", serverId, organizationId)
	req, err := http.NewRequest("DELETE", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("DetachOrganizationFromServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response on detaching the organization from the server\nbody=%s", body)
	}

	return nil
}

func (c client) StartServer(serverId string) error {
	url := fmt.Sprintf("/server/%s/operation/start", serverId)
	req, err := http.NewRequest("PUT", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("StartServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response on starting the server\nbody=%s", body)
	}

	return nil
}

func (c client) StopServer(serverId string) error {
	url := fmt.Sprintf("/server/%s/operation/stop", serverId)
	req, err := http.NewRequest("PUT", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("StopServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response on stopping the server\nbody=%s", body)
	}

	return nil
}

func (c client) GetRoutesByServer(serverId string) ([]Route, error) {
	url := fmt.Sprintf("/server/%s/route", serverId)
	req, err := http.NewRequest("GET", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetRoutesByServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Non-200 response on getting routes on the server\nbody=%s", body)
	}

	var routes []Route
	json.Unmarshal(body, &routes)

	return routes, nil
}

func (c client) AddRouteToServer(serverId string, route Route) error {
	jsonData, err := json.Marshal(route)

	url := fmt.Sprintf("/server/%s/route", serverId)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("AddRouteToServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response on adding a route to the server\nbody=%s", body)
	}

	return nil
}

func (c client) AddRoutesToServer(serverId string, routes []Route) error {
	jsonData, err := json.Marshal(routes)

	url := fmt.Sprintf("/server/%s/routes", serverId)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("AddRoutesToServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response on adding routes to the server\nbody=%s", body)
	}

	return nil
}

func (c client) UpdateRouteOnServer(serverId string, route Route) error {
	jsonData, err := json.Marshal(route)

	url := fmt.Sprintf("/server/%s/route/%s", serverId, route.GetID())
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("UpdateRouteOnServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response on updating a route on the server\nbody=%s", body)
	}

	return nil
}

func (c client) DeleteRouteFromServer(serverId string, route Route) error {
	url := fmt.Sprintf("/server/%s/route/%s", serverId, route.GetID())
	req, err := http.NewRequest("DELETE", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("DeleteRouteFromServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response on deleting a route on the server\nbody=%s", body)
	}

	return nil
}

func (c client) GetUser(id string, orgId string) (*User, error) {
	url := fmt.Sprintf("/user/%s/%s", orgId, id)
	req, err := http.NewRequest("GET", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetUser: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Non-200 response on getting the user\nbody=%s", body)
	}

	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, fmt.Errorf("GetUser: %s: %+v, id=%s, body=%s", err, user, id, body)
	}

	return &user, nil
}

func (c client) CreateUser(newUser User) (*User, error) {
	jsonData, err := json.Marshal(newUser)
	if err != nil {
		return nil, fmt.Errorf("CreateUser: Error on marshalling data: %s", err)
	}

	url := fmt.Sprintf("/user/%s", newUser.Organization)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CreateUser: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Non-200 response on creating the user\ncode=%d\nbody=%s", resp.StatusCode, body)
	}

	var users []User
	err = json.Unmarshal(body, &users)
	if err != nil {
		return nil, fmt.Errorf("CreateUser: Error on unmarshalling API response %s (body=%+v)", err, string(body))
	}

	if len(users) > 0 {
		return &users[0], nil
	}

	return nil, fmt.Errorf("empty users response")
}

func (c client) UpdateUser(id string, user *User) error {
	jsonData, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("UpdateUser: Error on marshalling data: %s", err)
	}

	url := fmt.Sprintf("/user/%s/%s", user.Organization, id)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("UpdateUser: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response on updating the user\nbody=%s", body)
	}

	return nil
}

func (c client) DeleteUser(id string, orgId string) error {
	url := fmt.Sprintf("/user/%s/%s", orgId, id)
	req, err := http.NewRequest("DELETE", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("DeleteUser: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response on deleting the user\nbody=%s", body)
	}

	return nil
}

func (c client) GetHosts() ([]Host, error) {
	url := fmt.Sprintf("/host")
	req, err := http.NewRequest("GET", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetHosts: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Non-200 response on getting the hosts\nbody=%s", body)
	}

	var hosts []Host

	err = json.Unmarshal(body, &hosts)
	if err != nil {
		return nil, fmt.Errorf("GetHosts: %s: %+v, body=%s", err, hosts, body)
	}

	return hosts, nil
}

func (c client) GetHostsByServer(serverId string) ([]Host, error) {
	url := fmt.Sprintf("/server/%s/host", serverId)
	req, err := http.NewRequest("GET", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetHostsByServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Non-200 response on getting hosts by the server\nbody=%s", body)
	}

	var hosts []Host

	err = json.Unmarshal(body, &hosts)
	if err != nil {
		return nil, fmt.Errorf("GetHostsByServer: %s: %+v, body=%s", err, hosts, body)
	}

	return hosts, nil
}

func (c client) AttachHostToServer(hostId, serverId string) error {
	url := fmt.Sprintf("/server/%s/host/%s", serverId, hostId)
	req, err := http.NewRequest("PUT", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("AttachHostToServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response on attachhing the host the server\nbody=%s", body)
	}

	return nil
}

func (c client) DetachHostFromServer(hostId, serverId string) error {
	url := fmt.Sprintf("/server/%s/host/%s", serverId, hostId)
	req, err := http.NewRequest("DELETE", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("DetachHostFromServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response on detaching the host from the server\nbody=%s", body)
	}

	return nil
}

func (c client) GetSettings() (*Settings, error) {
	url := fmt.Sprintf("/settings")
	req, err := http.NewRequest("GET", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetSettings: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Non-200 response on getting settings\nbody=%s", body)
	}

	var settings Settings

	err = json.Unmarshal(body, &settings)
	if err != nil {
		return nil, fmt.Errorf("GetSettings: %s: body=%s", err, body)
	}

	return &settings, nil
}

func (c client) UpdateSettings(settings *Settings) error {
	jsonData, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("UpdateSettings: Error on marshalling data: %s", err)
	}

	url := fmt.Sprintf("/settings")
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("UpdateSettings: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Non-200 response on updating settings\nbody=%s", body)
	}

	return nil
}

func NewClient(baseUrl, apiToken, apiSecret string, insecure bool) Client {
	underlyingTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}
	httpClient := &http.Client{
		Transport: &transport{
			baseUrl:             baseUrl,
			apiToken:            apiToken,
			apiSecret:           apiSecret,
			underlyingTransport: underlyingTransport,
		},
	}

	return &client{httpClient: httpClient}
}
