package provider

import (
	"context"
	"fmt"
	"github.com/disc/terraform-provider-pritunl/internal/pritunl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strings"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the user.",
			},
			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The organizations that user belongs to.",
				ValidateFunc: func(i interface{}, s string) ([]string, []error) {
					return validation.StringIsNotEmpty(i, s)
				},
			},
			"groups": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "Enter list of groups to allow connections from. Names are case sensitive. If empty all groups will able to connect.",
			},
			"email": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User email address.",
			},
			"disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Shows if user is disabled",
			},
			"port_forwarding": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
				Optional:    true,
				Description: "Comma seperated list of ports to forward using format source_port:dest_port/protocol or start_port-end_port/protocol. Such as 80, 80/tcp, 80:8000/tcp, 1000-2000/udp.",
			},
			"network_links": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "Network address with cidr subnet. This will provision access to a clients local network to the attached vpn servers and other clients. Multiple networks may be separated by a comma. Router must have a static route to VPN virtual network through client.",
			},
			"client_to_client": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Only allow this client to communicate with other clients. Access to routed networks will be blocked.",
			},
			"auth_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "User authentication type. This will determine how the user authenticates. This should be set automatically when the user authenticates with single sign-on.",
				ValidateFunc: validation.StringInSlice([]string{"local", "duo", "yubico", "azure", "azure_duo", "azure_yubico", "google", "google_duo", "google_yubico", "slack", "slack_duo", "slack_yubico", "saml", "saml_duo", "saml_yubico", "saml_okta", "saml_okta_duo", "saml_okta_yubico", "saml_onelogin", "saml_onelogin_duo", "saml_onelogin_yubico", "radius", "radius_duo", "plugin"}, false),
			},
			"mac_addresses": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "Comma separated list of MAC addresses client is allowed to connect from. The validity of the MAC address provided by the VPN client cannot be verified.",
			},
			"dns_servers": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "Dns server with port to forward sub-domain dns requests coming from this users domain. Multiple dns servers may be separated by a comma.",
			},
			"dns_suffix": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The suffix to use when forwarding dns requests. The full dns request will be the combination of the sub-domain of the users dns name suffixed by the dns suffix.",
			},
			"bypass_secondary": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Bypass secondary authentication such as the PIN and two-factor authentication. Use for server users that can't provide a two-factor code.",
			},
		},
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceUserImport,
		},
	}
}

func resourceUserRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	user, err := apiClient.GetUser(d.Id(), d.Get("organization_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", user.Name)
	d.Set("auth_type", user.AuthType)
	d.Set("dns_servers", user.DnsServers)
	d.Set("dns_suffix", user.DnsSuffix)
	d.Set("disabled", user.Disabled)
	d.Set("network_links", user.NetworkLinks)
	d.Set("port_forwarding", user.PortForwarding)
	d.Set("email", user.Email)
	d.Set("client_to_client", user.ClientToClient)
	d.Set("mac_addresses", user.MacAddresses)
	d.Set("bypass_secondary", user.BypassSecondary)
	d.Set("groups", user.Groups)
	d.Set("organization_id", user.Organization)

	return nil
}

func resourceUserDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	err := apiClient.DeleteUser(d.Id(), d.Get("organization_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	user, err := apiClient.GetUser(d.Id(), d.Get("organization_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	if v, ok := d.GetOk("name"); ok {
		user.Name = v.(string)
	}

	if v, ok := d.GetOk("organization_id"); ok {
		user.Organization = v.(string)
	}

	if d.HasChange("groups") {
		groups := make([]string, 0)
		for _, v := range d.Get("groups").([]interface{}) {
			groups = append(groups, v.(string))
		}
		user.Groups = groups
	}

	if v, ok := d.GetOk("email"); ok {
		user.Email = v.(string)
	}

	// TODO: Fixme
	if v, ok := d.GetOk("disabled"); ok {
		user.Disabled = v.(bool)
	}

	if d.HasChange("port_forwarding") {
		portForwarding := make([]map[string]interface{}, 0)
		for _, v := range d.Get("port_forwarding").([]interface{}) {
			portForwarding = append(portForwarding, v.(map[string]interface{}))
		}
		user.PortForwarding = portForwarding
	}

	if d.HasChange("network_links") {
		networkLinks := make([]string, 0)
		for _, v := range d.Get("network_links").([]interface{}) {
			networkLinks = append(networkLinks, v.(string))
		}
		user.NetworkLinks = networkLinks
	}

	if v, ok := d.GetOk("client_to_client"); ok {
		user.ClientToClient = v.(bool)
	}

	if v, ok := d.GetOk("auth_type"); ok {
		user.AuthType = v.(string)
	}

	if d.HasChange("mac_addresses") {
		macAddresses := make([]string, 0)
		for _, v := range d.Get("mac_addresses").([]interface{}) {
			macAddresses = append(macAddresses, v.(string))
		}
		user.MacAddresses = macAddresses
	}

	if d.HasChange("dns_servers") {
		dnsServers := make([]string, 0)
		for _, v := range d.Get("dns_servers").([]interface{}) {
			dnsServers = append(dnsServers, v.(string))
		}
		user.DnsServers = dnsServers
	}

	if v, ok := d.GetOk("dns_suffix"); ok {
		user.DnsSuffix = v.(string)
	}

	if v, ok := d.GetOk("bypass_secondary"); ok {
		user.BypassSecondary = v.(bool)
	}

	err = apiClient.UpdateUser(d.Id(), user)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceUserRead(ctx, d, meta)
}

func resourceUserCreate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	dnsServers := make([]string, 0)
	for _, v := range d.Get("dns_servers").([]interface{}) {
		dnsServers = append(dnsServers, v.(string))
	}

	macAddresses := make([]string, 0)
	for _, v := range d.Get("mac_addresses").([]interface{}) {
		macAddresses = append(macAddresses, v.(string))
	}

	networkLinks := make([]string, 0)
	for _, v := range d.Get("network_links").([]interface{}) {
		networkLinks = append(networkLinks, v.(string))
	}

	portForwarding := make([]map[string]interface{}, 0)
	for _, v := range d.Get("port_forwarding").([]interface{}) {
		portForwarding = append(portForwarding, v.(map[string]interface{}))
	}

	groups := make([]string, 0)
	for _, v := range d.Get("groups").([]interface{}) {
		groups = append(groups, v.(string))
	}

	userData := pritunl.User{
		Name:            d.Get("name").(string),
		Organization:    d.Get("organization_id").(string),
		AuthType:        d.Get("auth_type").(string),
		DnsServers:      dnsServers,
		DnsSuffix:       d.Get("dns_suffix").(string),
		Disabled:        d.Get("disabled").(bool),
		NetworkLinks:    networkLinks,
		PortForwarding:  portForwarding,
		Email:           d.Get("email").(string),
		ClientToClient:  d.Get("client_to_client").(bool),
		MacAddresses:    macAddresses,
		BypassSecondary: d.Get("bypass_secondary").(bool),
		Groups:          groups,
	}

	user, err := apiClient.CreateUser(userData)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(user.ID)

	return nil
}

func resourceUserImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(pritunl.Client)

	attributes := strings.Split(d.Id(), "-")
	if len(attributes) < 2 {
		return nil, fmt.Errorf("invalid format: expected ${organizationId}-${userId}, e.g. 60cd0be07723cf3c9114686c-60cd0be17723cf3c91146873, actual id is %s", d.Id())
	}

	orgId := attributes[0]
	userId := attributes[1]

	d.SetId(userId)
	d.Set("organization_id", orgId)

	_, err := apiClient.GetUser(userId, orgId)
	if err != nil {
		return nil, fmt.Errorf("error on getting user during import: %s", err)
	}

	return []*schema.ResourceData{d}, nil
}
