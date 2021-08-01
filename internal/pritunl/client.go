package pritunl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client interface {
	GetOrganization(id string) (*Organization, error)
	CreateOrganization(name string) (*Organization, error)
	UpdateOrganization(id string, organization *Organization) error
	DeleteOrganization(name string) error

	GetServer(id string) (*Server, error)
	CreateServer(name, protocol, cipher, hash string, port *int) (*Server, error)
	UpdateServer(id string, server *Server) error
	DeleteServer(id string) error

	GetAttachedOrganizationsOnServer(serverId string) ([]Organization, error)
	AttachOrganizationToServer(organizationId, serverId string) error
	DetachOrganizationFromServer(organizationId, serverId string) error

	AddRouteToServer(serverId string, route Route) error
	AddRoutesToServer(serverId string, route []Route) error
	DeleteRouteFromServer(serverId string, route Route) error
	UpdateRouteOnServer(serverId string, route Route) error

	StartServer(serverId string) error
	StopServer(serverId string) error
	//RestartServer(serverId string) error
	//DeleteServer(serverId string) error
}

type client struct {
	httpClient *http.Client
	baseUrl    string
}

func (c client) GetOrganization(id string) (*Organization, error) {
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
		return nil, fmt.Errorf("GetOrganization: %s: %+v, id=%s, body=%s", err, organization, id, body)
	}

	return &organization, nil
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
		return err
	}
	defer resp.Body.Close()

	return nil
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

	var organization Organization
	err = json.Unmarshal(body, &organization)
	if err != nil {
		return fmt.Errorf("DeleteOrganization: %s: %+v, id=%s, body=%s", err, organization, id, body)
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

	var server Server
	err = json.Unmarshal(body, &server)

	if err != nil {
		return nil, fmt.Errorf("GetServer: %s: %+v, id=%s, body=%s", err, server, id, body)
	}

	return &server, nil
}

func (c client) CreateServer(name, protocol, cipher, hash string, port *int) (*Server, error) {
	serverStruct := Server{
		Name:     name,
		Protocol: protocol,
		Cipher:   cipher,
		Hash:     hash,
	}

	if port != nil {
		serverStruct.Port = *port
	}

	jsonData, err := json.Marshal(serverStruct)

	url := "/server"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CreateServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var server Server
	err = json.Unmarshal(body, &server)
	if err != nil {
		return nil, err
	}

	return &server, nil
}

func (c client) UpdateServer(id string, server *Server) error {
	jsonData, err := json.Marshal(server)
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

	var server Server
	err = json.Unmarshal(body, &server)
	if err != nil {
		return fmt.Errorf("DeleteServer: Error on parsing response: %s (id=%s, body=%s)", err, id, body)
	}

	return nil
}

func (c client) GetAttachedOrganizationsOnServer(serverId string) ([]Organization, error) {
	url := fmt.Sprintf("/server/%s/organization", serverId)
	req, err := http.NewRequest("GET", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var organizations []Organization
	json.Unmarshal(body, &organizations)

	return organizations, nil
}

func (c client) AttachOrganizationToServer(organizationId, serverId string) error {
	url := fmt.Sprintf("/server/%s/organization/%s", serverId, organizationId)
	req, err := http.NewRequest("PUT", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c client) DetachOrganizationFromServer(organizationId, serverId string) error {
	url := fmt.Sprintf("/server/%s/organization/%s", serverId, organizationId)
	req, err := http.NewRequest("DELETE", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
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
		return err
	}
	defer resp.Body.Close()

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

	return nil
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

	return nil
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
