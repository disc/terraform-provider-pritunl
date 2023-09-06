package provider

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/disc/terraform-provider-pritunl/internal/pritunl"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceServer() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the server",
			},
			"protocol": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The protocol for the server",
				Default:      "udp",
				ValidateFunc: validation.StringInSlice([]string{"udp", "tcp"}, false),
			},
			"cipher": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The cipher for the server",
				Default:      "aes128",
				ValidateFunc: validation.StringInSlice([]string{"none", "bf128", "bf256", "aes128", "aes192", "aes256"}, false),
			},
			"hash": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The hash for the server",
				Default:      "sha1",
				ValidateFunc: validation.StringInSlice([]string{"none", "md5", "sha1", "sha256", "sha512"}, false),
			},
			"port": {
				Type:         schema.TypeInt,
				Required:     false,
				Optional:     true,
				Computed:     true,
				Description:  "The port for the server",
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"network": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Computed:    true,
				Description: "Network address for the private network that will be created for clients. This network cannot conflict with any existing local networks",

				ValidateFunc: func(i interface{}, s string) ([]string, []error) {
					// [10,172,192].[0-255,16-31,168].[0-255].0/[8-24]
					// 10.0.0.0/8
					// 172.16.0.0/12
					// 192.168.0.0/16
					warnings := make([]string, 0)
					errors := make([]error, 0)

					_, actualIpNet, err := net.ParseCIDR(i.(string))
					if err != nil {
						errors = append(errors, err)

						return warnings, errors
					}

					expectedIpNets := []string{
						"10.0.0.0/8",
						"172.16.0.0/12",
						"192.168.0.0/16",
					}

					found := false
					for _, v := range expectedIpNets {
						_, expectedIpNet, _ := net.ParseCIDR(v)
						if actualIpNet.Contains(expectedIpNet.IP) || expectedIpNet.Contains(actualIpNet.IP) {
							found = true
							break
						}
					}

					if !found {
						errors = append(errors, fmt.Errorf("provided subnet %s does not belong to expected subnets %s", actualIpNet.String(), strings.Join(expectedIpNets, ", ")))
					}

					return warnings, errors
				},
			},
			"bind_address": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "Network address for the private network that will be created for clients. This network cannot conflict with any existing local networks",
				ValidateFunc: func(i interface{}, s string) ([]string, []error) {
					return validation.IsIPAddress(i, s)
				},
			},
			"network_wg": {
				Type:         schema.TypeString,
				Required:     false,
				Optional:     true,
				Description:  "Network address for the private network that will be created for clients. This network cannot conflict with any existing local networks",
				RequiredWith: []string{"port_wg"},
				ValidateFunc: func(i interface{}, s string) ([]string, []error) {
					// [10,172,192].[0-255,16-31,168].[0-255].0/[8-24]
					// 10.0.0.0/8
					// 172.16.0.0/12
					// 192.168.0.0/16
					warnings := make([]string, 0)
					errors := make([]error, 0)

					_, actualIpNet, err := net.ParseCIDR(i.(string))
					if err != nil {
						errors = append(errors, err)

						return warnings, errors
					}

					expectedIpNets := []string{
						"10.0.0.0/8",
						"172.16.0.0/12",
						"192.168.0.0/16",
					}

					found := false
					for _, v := range expectedIpNets {
						_, expectedIpNet, _ := net.ParseCIDR(v)
						if actualIpNet.Contains(expectedIpNet.IP) || expectedIpNet.Contains(actualIpNet.IP) {
							found = true
							break
						}
					}

					if !found {
						errors = append(errors, fmt.Errorf("provided subnet %s does not belong to expected subnets %s", actualIpNet.String(), strings.Join(expectedIpNets, ", ")))
					}

					return warnings, errors
				},
			},
			"port_wg": {
				Type:         schema.TypeInt,
				Required:     false,
				Optional:     true,
				Description:  "Network address for the private network that will be created for clients. This network cannot conflict with any existing local networks",
				RequiredWith: []string{"network_wg"},
				ValidateFunc: validation.IntBetween(1, 65535),
				// TODO: Add validation
			},
			"groups": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    false,
				Optional:    true,
				Description: "Enter list of groups to allow connections from. Names are case sensitive. If empty all groups will able to connect",
			},
			"dns_servers": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: func(i interface{}, s string) ([]string, []error) {
						return validation.IsIPAddress(i, s)
					},
				},
				Required:    false,
				Optional:    true,
				Description: "Enter list of DNS servers applied on the client",
			},
			"otp_auth": {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				Description: "Enables two-step authentication using Google Authenticator. Verification code is entered as the user password when connecting",
			},
			"ipv6": {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				Description: "Enables IPv6 on server, requires IPv6 network interface",
			},
			"dh_param_bits": {
				Type:         schema.TypeInt,
				Required:     false,
				Optional:     true,
				Computed:     true,
				Description:  "Size of DH parameters",
				ValidateFunc: validation.IntInSlice([]int{1024, 1536, 2048, 2048, 3072, 4096}),
				// TODO: Cover the case " Generating DH parameters, please wait..." before start the server
			},
			"ping_interval": {
				Type:         schema.TypeInt,
				Required:     false,
				Optional:     true,
				Computed:     true,
				Description:  "Interval to ping client",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"ping_timeout": {
				Type:        schema.TypeInt,
				Required:    false,
				Optional:    true,
				Computed:    true,
				Description: "Timeout for client ping. Must be greater then ping interval",
				ValidateFunc: validation.All(
					validation.IntAtLeast(1),
					//func(i interface{}, s string) ([]string, []error) {
					//	TODO: Implement "Must be greater then ping interval" rule
					//},
				),
			},
			"link_ping_interval": {
				Type:         schema.TypeInt,
				Required:     false,
				Optional:     true,
				Computed:     true,
				Description:  "Time in between pings used when multiple users have the same network link to failover to another user when one network link fails.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"link_ping_timeout": {
				Type:         schema.TypeInt,
				Required:     false,
				Optional:     true,
				Computed:     true,
				Description:  "Optional, ping timeout used when multiple users have the same network link to failover to another user when one network link fails..",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"session_timeout": {
				Type:         schema.TypeInt,
				Required:     false,
				Optional:     true,
				Description:  "Disconnect users after the specified number of seconds.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"inactive_timeout": {
				Type:         schema.TypeInt,
				Required:     false,
				Optional:     true,
				Description:  "Disconnects users after the specified number of seconds of inactivity.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"max_clients": {
				Type:         schema.TypeInt,
				Required:     false,
				Optional:     true,
				Computed:     true,
				Description:  "Maximum number of clients connected to a server or to each server replica.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"network_mode": {
				Type:         schema.TypeString,
				Required:     false,
				Optional:     true,
				Description:  "Sets network mode. Bridged mode is not recommended using it will impact performance and client support will be limited.",
				ValidateFunc: validation.StringInSlice([]string{"tunnel", "bridge"}, false),
			},
			"network_start": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "Starting network address for the bridged VPN client IP addresses. Must be in the subnet of the server network.",
				ValidateFunc: func(i interface{}, s string) ([]string, []error) {
					return validation.IsIPAddress(i, s)
				},
				RequiredWith: []string{"network_mode", "network_end"},
			},
			"network_end": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "Ending network address for the bridged VPN client IP addresses. Must be in the subnet of the server network.",
				ValidateFunc: func(i interface{}, s string) ([]string, []error) {
					return validation.IsIPAddress(i, s)
				},
				RequiredWith: []string{"network_mode", "network_start"},
			},
			"mss_fix": {
				Type:        schema.TypeInt,
				Required:    false,
				Optional:    true,
				Description: "MSS fix value",
			},
			"max_devices": {
				Type:         schema.TypeInt,
				Required:     false,
				Optional:     true,
				Description:  "Maximum number of devices per client connected to a server.",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"pre_connect_msg": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "Messages that will be shown after connect to the server",
			},
			"allowed_devices": {
				Type:         schema.TypeString,
				Required:     false,
				Optional:     true,
				Description:  "Device types permitted to connect to server.",
				ValidateFunc: validation.StringInSlice([]string{"mobile", "desktop"}, false),
			},
			"search_domain": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "DNS search domain for clients. Separate multiple search domains by a comma.",
				// TODO: Add validation
			},
			"replica_count": {
				Type:         schema.TypeInt,
				Required:     false,
				Optional:     true,
				Computed:     true,
				Description:  "Replicate server across multiple hosts.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"multi_device": {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				Description: "Allow users to connect with multiple devices concurrently.",
			},
			"debug": {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				Description: "Show server debugging information in output.",
			},
			"sso_auth": {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				Description: "Require client to authenticate with single sign-on provider on each connection using web browser.",
			},
			"restrict_routes": {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				Description: "Prevent traffic from networks not specified in the servers routes from being tunneled over the vpn.",
			},
			"block_outside_dns": {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				Description: "Block outside DNS on Windows clients.",
			},
			"dns_mapping": {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				Description: "Map the vpn clients ip address to the .vpn domain such as example_user.example_org.vpn This will conflict with the DNS port if systemd-resolve is running.",
			},
			"inter_client": {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				Description: "Enable inter-client routing across hosts.",
			},
			"vxlan": {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				Description: "Use VXLan for routing client-to-client traffic with replicated servers.",
			},
			"organization_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    false,
				Optional:    true,
				Description: "The list of attached organizations to the server.",
			},
			"host_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    false,
				Optional:    true,
				Computed:    true,
				Description: "The list of attached hosts to the server",
			},
			"route": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Network address with subnet to route",
							ValidateFunc: func(i interface{}, s string) ([]string, []error) {
								return validation.IsCIDR(i, s)
							},
						},
						"comment": {
							Type:        schema.TypeString,
							Required:    false,
							Optional:    true,
							Description: "Comment for route",
						},
						"nat": {
							Type:        schema.TypeBool,
							Required:    false,
							Optional:    true,
							Description: "NAT vpn traffic destined to this network",
							Computed:    true,
						},
						"net_gateway": {
							Type:        schema.TypeBool,
							Required:    false,
							Optional:    true,
							Description: "Net Gateway vpn traffic destined to this network",
							Computed:    true,
						},
					},
				},
				Required:    false,
				Optional:    true,
				Description: "The list of attached routes to the server",
			},
			"status": {
				Type:         schema.TypeString,
				Required:     false,
				Optional:     true,
				Computed:     true,
				Description:  "The status of the server",
				RequiredWith: []string{"organization_ids"},
				ValidateDiagFunc: func(v interface{}, path cty.Path) diag.Diagnostics {
					allowedStatusesMap := map[string]struct{}{
						pritunl.ServerStatusOffline: {},
						pritunl.ServerStatusOnline:  {},
					}

					allowedStatusesList := make([]string, 0)
					for status := range allowedStatusesMap {
						allowedStatusesList = append(allowedStatusesList, status)
					}

					if _, ok := allowedStatusesMap[strings.ToLower(v.(string))]; !ok {
						return diag.Diagnostics{
							{
								Severity:      diag.Error,
								Summary:       "Unsupported value for the `status` attribute",
								Detail:        fmt.Sprintf("Supported values are: %s", strings.Join(allowedStatusesList, ", ")),
								AttributePath: cty.Path{cty.GetAttrStep{Name: "status"}},
							},
						}
					}

					return nil
				},
			},
		},
		CreateContext: resourceCreateServer,
		ReadContext:   resourceReadServer,
		UpdateContext: resourceUpdateServer,
		DeleteContext: resourceDeleteServer,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

// Uses for importing
func resourceReadServer(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	server, err := apiClient.GetServer(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// get organizations
	organizations, err := apiClient.GetOrganizationsByServer(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// get routes
	routes, err := apiClient.GetRoutesByServer(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// get hosts
	hosts, err := apiClient.GetHostsByServer(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", server.Name)
	d.Set("protocol", server.Protocol)
	d.Set("port", server.Port)
	d.Set("cipher", server.Cipher)
	d.Set("hash", server.Hash)
	d.Set("network", server.Network)
	d.Set("bind_address", server.BindAddress)
	d.Set("dns_servers", server.DnsServers)
	d.Set("network_wg", server.NetworkWG)
	d.Set("port_wg", server.PortWG)
	d.Set("otp_auth", server.OtpAuth)
	d.Set("ipv6", server.IPv6)
	d.Set("dh_param_bits", server.DhParamBits)
	d.Set("ping_interval", server.PingInterval)
	d.Set("ping_timeout", server.PingTimeout)
	d.Set("link_ping_interval", server.LinkPingInterval)
	d.Set("link_ping_timeout", server.LinkPingTimeout)
	d.Set("session_timeout", server.SessionTimeout)
	d.Set("inactive_timeout", server.InactiveTimeout)
	d.Set("max_clients", server.MaxClients)
	d.Set("network_mode", server.NetworkMode)
	d.Set("network_start", server.NetworkStart)
	d.Set("network_end", server.NetworkEnd)
	d.Set("mss_fix", server.MssFix)
	d.Set("max_devices", server.MaxDevices)
	d.Set("pre_connect_msg", server.PreConnectMsg)
	d.Set("allowed_devices", server.AllowedDevices)
	d.Set("search_domain", server.SearchDomain)
	d.Set("replica_count", server.ReplicaCount)
	d.Set("multi_device", server.MultiDevice)
	d.Set("debug", server.Debug)
	d.Set("sso_auth", server.SsoAuth)
	d.Set("restrict_routes", server.RestrictRoutes)
	d.Set("block_outside_dns", server.BlockOutsideDns)
	d.Set("dns_mapping", server.DnsMapping)
	d.Set("inter_client", server.InterClient)
	d.Set("vxlan", server.VxLan)
	d.Set("status", server.Status)

	if len(organizations) > 0 {
		organizationsList := make([]string, 0)

		if organizations != nil {
			for _, organization := range organizations {
				organizationsList = append(organizationsList, organization.ID)
			}
		}

		declaredOrganizations, ok := d.Get("organization_ids").([]interface{})
		if !ok {
			return diag.Errorf("failed to parse organization_ids for the server: %s", server.Name)
		}

		if len(declaredOrganizations) > 0 {
			organizationsList = matchStringEntitiesWithSchema(organizationsList, declaredOrganizations)
		}

		d.Set("organization_ids", organizationsList)
	}

	if len(server.Groups) > 0 {
		groupsList := make([]string, 0)

		for _, group := range server.Groups {
			groupsList = append(groupsList, group)
		}

		declaredGroups, ok := d.Get("groups").([]interface{})
		if !ok {
			return diag.Errorf("failed to parse groups for the server: %s", server.Name)
		}

		if len(declaredGroups) > 0 {
			groupsList = matchStringEntitiesWithSchema(groupsList, declaredGroups)
		}

		d.Set("groups", groupsList)
	}

	if len(routes) > 0 {
		declaredRoutes, ok := d.Get("route").([]interface{})
		if !ok {
			return diag.Errorf("failed to parse routes for the server: %s", server.Name)
		}

		if len(declaredRoutes) > 0 {
			routes = matchRoutesWithSchema(routes, declaredRoutes)
		}

		d.Set("route", flattenRoutesData(routes))
	}

	if len(hosts) > 0 {
		hostsList := make([]string, 0)

		if hosts != nil {
			for _, host := range hosts {
				hostsList = append(hostsList, host.ID)
			}
		}

		declaredHosts, ok := d.Get("host_ids").([]interface{})
		if !ok {
			return diag.Errorf("failed to parse host_ids for the server: %s", server.Name)
		}

		if len(declaredHosts) > 0 {
			hostsList = matchStringEntitiesWithSchema(hostsList, declaredHosts)
		}

		d.Set("host_ids", hostsList)
	}

	return nil
}

func resourceCreateServer(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	serverData := map[string]interface{}{
		"name":               d.Get("name"),
		"protocol":           d.Get("protocol"),
		"port":               d.Get("port"),
		"network":            d.Get("network"),
		"cipher":             d.Get("cipher"),
		"hash":               d.Get("hash"),
		"bind_address":       d.Get("bind_address"),
		"groups":             d.Get("groups"),
		"dns_servers":        d.Get("dns_servers"),
		"network_wg":         d.Get("network_wg"),
		"port_wg":            d.Get("port_wg"),
		"otp_auth":           d.Get("otp_auth"),
		"ipv6":               d.Get("ipv6"),
		"dh_param_bits":      d.Get("dh_param_bits"),
		"ping_interval":      d.Get("ping_interval"),
		"ping_timeout":       d.Get("ping_timeout"),
		"link_ping_interval": d.Get("link_ping_interval"),
		"link_ping_timeout":  d.Get("link_ping_timeout"),
		"session_timeout":    d.Get("session_timeout"),
		"inactive_timeout":   d.Get("inactive_timeout"),
		"max_clients":        d.Get("max_clients"),
		"network_mode":       d.Get("network_mode"),
		"network_start":      d.Get("network_start"),
		"network_end":        d.Get("network_end"),
		"mss_fix":            d.Get("mss_fix"),
		"max_devices":        d.Get("max_devices"),
		"pre_connect_msg":    d.Get("pre_connect_msg"),
		"allowed_devices":    d.Get("allowed_devices"),
		"search_domain":      d.Get("search_domain"),
		"replica_count":      d.Get("replica_count"),
		"multi_device":       d.Get("multi_device"),
		"debug":              d.Get("debug"),
		"sso_auth":           d.Get("sso_auth"),
		"restrict_routes":    d.Get("restrict_routes"),
		"block_outside_dns":  d.Get("block_outside_dns"),
		"dns_mapping":        d.Get("dns_mapping"),
		"inter_client":       d.Get("inter_client"),
		"vxlan":              d.Get("vxlan"),
	}

	server, err := apiClient.CreateServer(serverData)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(server.ID)

	if d.HasChange("organization_ids") {
		_, newOrgs := d.GetChange("organization_ids")
		for _, v := range newOrgs.([]interface{}) {
			err = apiClient.AttachOrganizationToServer(v.(string), d.Id())
			if err != nil {
				return diag.Errorf("Error on attaching server to the organization: %s", err)
			}
		}
	}

	// Delete default route
	defaultRoute := pritunl.Route{
		Network: "0.0.0.0/0",
		Nat:     true,
	}
	err = apiClient.DeleteRouteFromServer(d.Id(), defaultRoute)
	if err != nil {
		return diag.Errorf("Error on attaching server to the organization: %s", err)
	}

	if d.HasChange("route") {
		_, newRoutes := d.GetChange("route")
		routes := make([]pritunl.Route, 0)

		for _, v := range newRoutes.([]interface{}) {
			routes = append(routes, pritunl.ConvertMapToRoute(v.(map[string]interface{})))
		}

		err = apiClient.AddRoutesToServer(d.Id(), routes)
		if err != nil {
			return diag.Errorf("Error on attaching route from the server: %s", err)
		}
	}

	if d.HasChange("host_ids") {
		// delete default host(s) only when host_ids aren't empty

		hosts, err := apiClient.GetHostsByServer(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}
		for _, host := range hosts {
			err = apiClient.DetachHostFromServer(host.ID, d.Id())
			if err != nil {
				return diag.Errorf("Error on detaching a host from the server: %s", err)
			}
		}

		_, newHosts := d.GetChange("host_ids")
		for _, v := range newHosts.([]interface{}) {
			err = apiClient.AttachHostToServer(v.(string), d.Id())
			if err != nil {
				return diag.Errorf("Error on attaching a host to the server: %s", err)
			}
		}
	}

	if d.Get("status").(string) == pritunl.ServerStatusOnline {
		err = apiClient.StartServer(d.Id())
		if err != nil {
			return diag.Errorf("Error on starting server: %s", err)
		}
	}

	return resourceReadServer(ctx, d, meta)
}

func resourceUpdateServer(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	server, err := apiClient.GetServer(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	prevServerStatus := server.Status

	if v, ok := d.GetOk("name"); ok {
		server.Name = v.(string)
	}

	if v, ok := d.GetOk("protocol"); ok {
		server.Protocol = v.(string)
	}

	if v, ok := d.GetOk("cipher"); ok {
		server.Cipher = v.(string)
	}

	if v, ok := d.GetOk("hash"); ok {
		server.Hash = v.(string)
	}

	if v, ok := d.GetOk("port"); ok {
		server.Port = v.(int)
	}

	if v, ok := d.GetOk("network"); ok {
		server.Network = v.(string)
	}

	if d.HasChange("bind_address") {
		server.BindAddress = d.Get("bind_address").(string)
	}

	if d.HasChange("network_wg") {
		server.NetworkWG = d.Get("network_wg").(string)
	}

	if d.HasChange("port_wg") {
		server.PortWG = d.Get("port_wg").(int)
	}

	isWgEnabled := server.NetworkWG != "" && server.PortWG > 0
	server.WG = isWgEnabled

	if d.HasChange("otp_auth") {
		server.OtpAuth = d.Get("otp_auth").(bool)
	}

	if d.HasChange("ipv6") {
		server.IPv6 = d.Get("ipv6").(bool)
	}

	if d.HasChange("dh_param_bits") {
		server.DhParamBits = d.Get("dh_param_bits").(int)
	}

	if d.HasChange("ping_interval") {
		server.PingInterval = d.Get("ping_interval").(int)
	}

	if d.HasChange("ping_timeout") {
		server.PingTimeout = d.Get("ping_timeout").(int)
	}

	if d.HasChange("link_ping_interval") {
		server.LinkPingInterval = d.Get("link_ping_interval").(int)
	}

	if d.HasChange("link_ping_timeout") {
		server.LinkPingTimeout = d.Get("link_ping_timeout").(int)
	}

	if d.HasChange("session_timeout") {
		server.SessionTimeout = d.Get("session_timeout").(int)
	}

	if d.HasChange("inactive_timeout") {
		server.InactiveTimeout = d.Get("inactive_timeout").(int)
	}

	if d.HasChange("max_clients") {
		server.MaxClients = d.Get("max_clients").(int)
	}

	if d.HasChange("network_mode") {
		server.NetworkMode = d.Get("network_mode").(string)
	}

	if d.HasChange("network_start") {
		server.NetworkStart = d.Get("network_start").(string)
	}

	if d.HasChange("network_end") {
		server.NetworkEnd = d.Get("network_end").(string)
	}

	if server.NetworkMode == pritunl.ServerNetworkModeBridge && (server.NetworkStart == "" || server.NetworkEnd == "") {
		return diag.Errorf("the attribute network_mode = %s requires network_start and network_end attributes", pritunl.ServerNetworkModeBridge)
	}

	if d.HasChange("mss_fix") {
		server.MssFix = d.Get("mss_fix").(int)
	}

	if d.HasChange("max_devices") {
		server.MaxDevices = d.Get("max_devices").(int)
	}

	if d.HasChange("pre_connect_msg") {
		server.PreConnectMsg = d.Get("pre_connect_msg").(string)
	}

	if d.HasChange("allowed_devices") {
		server.AllowedDevices = d.Get("allowed_devices").(string)
	}

	if d.HasChange("search_domain") {
		server.SearchDomain = d.Get("search_domain").(string)
	}

	if d.HasChange("replica_count") {
		server.ReplicaCount = d.Get("replica_count").(int)
	}

	if d.HasChange("multi_device") {
		server.MultiDevice = d.Get("multi_device").(bool)
	}

	if d.HasChange("debug") {
		server.Debug = d.Get("debug").(bool)
	}

	if d.HasChange("sso_auth") {
		server.Debug = d.Get("sso_auth").(bool)
	}

	if d.HasChange("restrict_routes") {
		server.RestrictRoutes = d.Get("restrict_routes").(bool)
	}

	if d.HasChange("block_outside_dns") {
		server.BlockOutsideDns = d.Get("block_outside_dns").(bool)
	}

	if d.HasChange("dns_mapping") {
		server.DnsMapping = d.Get("dns_mapping").(bool)
	}

	if d.HasChange("vxlan") {
		server.VxLan = d.Get("vxlan").(bool)
	}

	if d.HasChange("groups") {
		groups := make([]string, 0)
		for _, v := range d.Get("groups").([]interface{}) {
			groups = append(groups, v.(string))
		}
		server.Groups = groups
	}

	if d.HasChange("dns_servers") {
		dnsServers := make([]string, 0)
		for _, v := range d.Get("dns_servers").([]interface{}) {
			dnsServers = append(dnsServers, v.(string))
		}
		server.DnsServers = dnsServers
	}

	// Stop server before applying any change
	err = apiClient.StopServer(d.Id())
	if err != nil {
		return diag.Errorf("Error on stopping server: %s", err)
	}

	if d.HasChange("organization_ids") {
		oldOrgs, newOrgs := d.GetChange("organization_ids")

		oldOrgsOnly := diffStringLists(oldOrgs.([]interface{}), newOrgs.([]interface{}))
		for _, v := range oldOrgsOnly {
			err = apiClient.DetachOrganizationFromServer(v, d.Id())
			if err != nil {
				return diag.Errorf("Error on detaching server to the organization: %s", err)
			}
		}

		newOrgsOnly := diffStringLists(newOrgs.([]interface{}), oldOrgs.([]interface{}))
		for _, v := range newOrgsOnly {
			err = apiClient.AttachOrganizationToServer(v, d.Id())
			if err != nil {
				return diag.Errorf("Error on attaching server to the organization: %s", err)
			}
		}
	}

	if d.HasChange("route") {
		oldRoutes, newRoutes := d.GetChange("route")

		newRoutesMap := make(map[string]pritunl.Route, 0)
		for _, v := range newRoutes.([]interface{}) {
			route := pritunl.ConvertMapToRoute(v.(map[string]interface{}))
			newRoutesMap[route.GetID()] = route
		}
		oldRoutesMap := make(map[string]pritunl.Route, 0)
		for _, v := range oldRoutes.([]interface{}) {
			route := pritunl.ConvertMapToRoute(v.(map[string]interface{}))
			oldRoutesMap[route.GetID()] = route
		}

		for _, route := range newRoutesMap {
			if _, found := oldRoutesMap[route.GetID()]; found {
				// update or skip
				err = apiClient.UpdateRouteOnServer(d.Id(), route)
				if err != nil {
					return diag.Errorf("Error on updating route on the server: %s", err)
				}
			} else {
				// add route
				err = apiClient.AddRouteToServer(d.Id(), route)
				if err != nil {
					return diag.Errorf("Error on attaching route from the server: %s", err)
				}
			}
		}

		for _, route := range oldRoutesMap {
			if _, found := newRoutesMap[route.GetID()]; !found {
				// delete route
				err = apiClient.DeleteRouteFromServer(d.Id(), route)
				if err != nil {
					return diag.Errorf("Error on detaching route from the server: %s", err)
				}
			}
		}
	}

	if d.HasChange("host_ids") {
		oldHosts, newHosts := d.GetChange("host_ids")
		for _, v := range oldHosts.([]interface{}) {
			err = apiClient.DetachHostFromServer(v.(string), d.Id())
			if err != nil {
				return diag.Errorf("Error on detaching server to the organization: %s", err)
			}
		}
		for _, v := range newHosts.([]interface{}) {
			err = apiClient.AttachHostToServer(v.(string), d.Id())
			if err != nil {
				return diag.Errorf("Error on attaching server to the organization: %s", err)
			}
		}
	}

	// Start server if it was ONLINE before and status wasn't changed OR status was changed to ONLINE
	shouldServerBeStarted := (prevServerStatus == pritunl.ServerStatusOnline && !d.HasChange("status")) || (d.HasChange("status") && d.Get("status").(string) != pritunl.ServerStatusOffline)

	err = apiClient.UpdateServer(d.Id(), server)
	if err != nil {
		// start server in case of error?
		return diag.FromErr(err)
	}

	if shouldServerBeStarted {
		err = apiClient.StartServer(d.Id())
		if err != nil {
			return diag.Errorf("Error on starting server: %s", err)
		}
	}

	return resourceReadServer(ctx, d, meta)
}

func resourceDeleteServer(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	err := apiClient.DeleteServer(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func diffStringLists(mainList []interface{}, otherList []interface{}) []string {
	result := make([]string, 0)
	var found bool

	for _, i := range mainList {
		found = false
		for _, j := range otherList {
			if i.(string) == j.(string) {
				found = true
				break
			}
		}
		if !found {
			result = append(result, i.(string))
		}
	}

	return result
}

func flattenRoutesData(routesList []pritunl.Route) []interface{} {
	routes := make([]interface{}, 0)

	if routesList != nil {
		for _, route := range routesList {
			if route.VirtualNetwork {
				// skip virtual network route
				continue
			}

			routeMap := make(map[string]interface{})

			routeMap["network"] = route.Network
			routeMap["nat"] = route.Nat
			routeMap["net_gateway"] = route.NetGateway
			if route.Comment != "" {
				routeMap["comment"] = route.Comment
			}

			routes = append(routes, routeMap)
		}
	}

	return routes
}

// This cannot currently be handled efficiently by a DiffSuppressFunc
// See: https://github.com/hashicorp/terraform-plugin-sdk/issues/477
func matchRoutesWithSchema(routes []pritunl.Route, declaredRoutes []interface{}) []pritunl.Route {
	result := make([]pritunl.Route, len(declaredRoutes))

	routesMap := make(map[string]pritunl.Route, len(declaredRoutes))
	for _, route := range routes {
		routesMap[route.GetID()] = route
	}

	for i, declaredRoute := range declaredRoutes {
		declaredRouteMap := declaredRoute.(map[string]interface{})

		for key, route := range routesMap {
			if route.Network != declaredRouteMap["network"] || route.Nat != declaredRouteMap["nat"] || route.NetGateway != declaredRouteMap["net_gateway"] {
				continue
			}

			result[i] = route
			delete(routesMap, key)
			break
		}
	}

	for _, route := range routesMap {
		result = append(result, route)
	}

	return result
}

// This cannot currently be handled efficiently by a DiffSuppressFunc
// See: https://github.com/hashicorp/terraform-plugin-sdk/issues/477
func matchStringEntitiesWithSchema(entities []string, declaredEntities []interface{}) []string {
	if len(declaredEntities) == 0 {
		return entities
	}

	result := make([]string, len(declaredEntities))

	for i, declaredEntity := range declaredEntities {
		for _, entity := range entities {
			if entity != declaredEntity.(string) {
				continue
			}

			result[i] = entity
			break
		}
	}

	return result
}
