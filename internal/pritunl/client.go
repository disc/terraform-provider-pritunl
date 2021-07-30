package pritunl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Client interface {
	GetOrganizationByID(id string) (*Organization, error)
	GetOrganization(name string) (*Organization, error)
	CreateOrganization(name string) (*Organization, error)
	RenameOrganization(id string, name string) error
	DeleteOrganization(name string) error

	CreateServer(name, protocol, cipher, hash string) (*Server, error)
	AttachOrganizationToServer(organizationId, serverId string) error

	StartServer(serverId string) error
	StopServer(serverId string) error
	//RestartServer(serverId string) error
	//DeleteServer(serverId string) error

	AddRouteToServer(serverId string, network string) error
}

type client struct {
	httpClient *http.Client
	baseUrl    string
}

func (c client) GetOrganizationByID(id string) (*Organization, error) {
	url := fmt.Sprintf("/organization/%s", id)
	req, err := http.NewRequest("GET", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	// iterate over all pages
	var organization Organization

	err = json.Unmarshal(body, &organization)
	if err != nil {
		return nil, fmt.Errorf("GetOrganizationByID: %s: %+v, id=%s, body=%s", err, organization, id, body)
	}

	return &organization, nil
}

func (c client) GetOrganization(name string) (*Organization, error) {
	url := "/organization"
	req, err := http.NewRequest("GET", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	type GetOrganizationsApiResponse struct {
		Organizations []Organization
	}

	var organizations []Organization

	err = json.Unmarshal(body, &organizations)
	if err != nil {
		return nil, fmt.Errorf("GetOrganization: %s: %+v, name=%s, body=%s", err, organizations, name, body)
	}

	for _, organization := range organizations {
		if strings.ToLower(organization.Name) == strings.ToLower(name) {
			return &organization, nil
		}
	}

	return nil, nil
}

func (c client) CreateOrganization(name string) (*Organization, error) {
	var jsonStr = []byte(`{"name": "` + name + `"}`)

	url := "/organization"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	var organization Organization
	err = json.Unmarshal(body, &organization)
	if err != nil {
		return nil, fmt.Errorf("CreateOrganization: %s: %+v, name=%s, body=%s", err, organization, name, body)
	}

	return &organization, nil
}

func (c client) RenameOrganization(id string, name string) error {
	panic("implement me")
}

func (c client) DeleteOrganization(id string) error {
	url := fmt.Sprintf("/organization/%s", id)
	req, err := http.NewRequest("DELETE", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	var organization Organization
	err = json.Unmarshal(body, &organization)
	if err != nil {
		return fmt.Errorf("DeleteOrganization: %s: %+v, id=%s, body=%s", err, organization, id, body)
	}

	return nil
}

/*
	{"name":"test-server","network":"192.168.226.0/24","port":15760,"protocol":"udp","dh_param_bits":2048,"ipv6_firewall":true,"dns_servers":["8.8.8.8"],"cipher":"aes128","hash":"sha1","inter_client":true,"restrict_routes":true,"vxlan":true,"id":null,"status":null,"uptime":null,"users_online":null,"devices_online":null,"user_count":null,"network_wg":"","groups":[],"bind_address":null,"port_wg":null,"ipv6":false,"network_mode":"tunnel","network_start":"","network_end":"","wg":false,"multi_device":false,"search_domain":null,"otp_auth":false,"block_outside_dns":false,"jumbo_frames":null,"lzo_compression":null,"ping_interval":null,"ping_timeout":null,"link_ping_interval":null,"link_ping_timeout":null,"inactive_timeout":null,"session_timeout":null,"allowed_devices":null,"max_clients":null,"max_devices":null,"replica_count":1,"dns_mapping":false,"debug":false,"pre_connect_msg":null,"mss_fix":null}
*/
func (c client) CreateServer(name, protocol, cipher, hash string) (*Server, error) {
	var jsonStr = []byte(`{"name": "` + name + `", "protocol": "` + protocol + `", "cipher": "` + cipher + `", "hash": "` + hash + `"}`)

	url := "/server"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	var server Server
	err = json.Unmarshal(body, &server)
	if err != nil {
		return nil, err
	}

	return &server, nil
}
func (c client) AttachOrganizationToServer(organizationId, serverId string) error {
	// /server/61032df34bce2ca96a7571ed/organization/6102ffef1332c1d92cf35cb5

	url := fmt.Sprintf("/server/%s/organization/%s", serverId, organizationId)
	req, err := http.NewRequest("PUT", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	return nil
}

func (c client) StartServer(serverId string) error {
	url := fmt.Sprintf("/server/%s/operation/start", serverId)
	req, err := http.NewRequest("PUT", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	return nil
}

func (c client) StopServer(serverId string) error {
	url := fmt.Sprintf("/server/%s/operation/stop", serverId)
	req, err := http.NewRequest("PUT", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	return nil
}

// POST /server/610332d44bce2ca96a757523/route
// {"comment": null, "vpc_region": null, "metric": null, "advertise": false, "nat_interface": null, "id": "382e382e382e322f3332", "nat_netmap": null, "network": "8.8.8.2/32", "server": "610332d44bce2ca96a757523", "nat": true, "vpc_id": null, "net_gateway": false}
func (c client) AddRouteToServer(serverId string, network string) error {
	err := c.StopServer(serverId)
	if err != nil {
		return err
	}

	var jsonStr = []byte(`{"server": "` + serverId + `", "network": "` + network + `"}`)

	url := fmt.Sprintf("/server/%s/route", serverId)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	err = c.StartServer(serverId)
	if err != nil {
		return err
	}

	return nil
}

func NewHttpClient(baseUrl, apiToken, apiSecret string) *http.Client {
	return &http.Client{
		Transport: &transport{
			baseUrl:             baseUrl,
			apiToken:            apiToken,
			apiSecret:           apiSecret,
			underlyingTransport: http.DefaultTransport,
		},
	}
}

func NewClient(baseUrl, apiToken, apiSecret string) Client {
	httpClient := &http.Client{
		Transport: &transport{
			baseUrl:             baseUrl,
			apiToken:            apiToken,
			apiSecret:           apiSecret,
			underlyingTransport: http.DefaultTransport,
		},
	}

	return &client{httpClient: httpClient}
}
