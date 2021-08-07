package provider

import (
	"context"
	"fmt"
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
			"organization": {
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
			"pin": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User pin, required when user connects to vpn. When using with two-factor authentication the pin and two-factor authentication code should both be placed in the password field.",
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
	apiClient := meta.(Client)

	user, err := apiClient.GetUser(d.Id(), d.Get("organization").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", user.Name)
	d.Set("auth_type", user.AuthType)
	d.Set("dns_servers", user.DnsServers)
	d.Set("pin", user.Pin)
	d.Set("dns_suffix", user.DnsSuffix)
	d.Set("disabled", user.Disabled)
	d.Set("network_links", user.NetworkLinks)
	d.Set("port_forwarding", user.PortForwarding)
	d.Set("email", user.Email)
	d.Set("client_to_client", user.ClientToClient)
	d.Set("mac_addresses", user.MacAddresses)
	d.Set("yubico_id", user.YubicoID)
	d.Set("sso", user.SSO)
	d.Set("bypass_secondary", user.BypassSecondary)
	d.Set("groups", user.Groups)
	d.Set("organization", user.Organization)

	return nil
}

func resourceUserDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(Client)

	err := apiClient.DeleteUser(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func resourceUserUpdate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(Client)

	organization, err := apiClient.GetOrganization(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("name") {
		organization.Name = d.Get("name").(string)

		err = apiClient.UpdateOrganization(d.Id(), organization)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceUserCreate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(Client)

	organization, err := apiClient.CreateOrganization(d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(organization.ID)

	return nil
}

func resourceUserImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(Client)

	attributes := strings.Split(d.Id(), "-")
	if len(attributes) < 2 {
		return nil, fmt.Errorf("invalid format: expected ${organizationId}-${userId}, e.g. 60cd0be07723cf3c9114686c-60cd0be17723cf3c91146873")
	}

	orgId := attributes[0]
	userId := attributes[1]

	d.SetId(userId)
	d.Set("organization", orgId)

	_, err := apiClient.GetUser(userId, orgId)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
