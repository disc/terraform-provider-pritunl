package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strconv"
	"strings"
)

func resourceServer() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the server",
				ForceNew:    false,
			},
			"protocol": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The protocol for the server",
				Default:      "udp",
				ForceNew:     false,
				ValidateFunc: validation.StringInSlice([]string{"udp", "tcp"}, true),
			},
			"cipher": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The cipher for the server",
				Default:      "aes128",
				ForceNew:     false,
				ValidateFunc: validation.StringInSlice([]string{"none", "bf128", "bf256", "aes128", "aes192", "aes256"}, true),
			},
			"hash": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The hash for the server",
				Default:      "sha1",
				ForceNew:     false,
				ValidateFunc: validation.StringInSlice([]string{"none", "md5", "sha1", "sha256", "sha512"}, true),
			},
			"port": {
				Type:         schema.TypeInt,
				Required:     false,
				Optional:     true,
				Computed:     true,
				Description:  "The port for the server",
				ForceNew:     false,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"network": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Computed:    true,
				Description: "Network address for the private network that will be created for clients. This network cannot conflict with any existing local networks",
				ForceNew:    false,
				//ValidateFunc: validation.Any(
				//	// [10,172,192].[0-255,16-31,168].[0-255].0/[8-24]
				//	func(i interface{}, s string) ([]string, []error) {
				//		return validation.IsIPv4Address(i.(string), "10.0.0.0/8")
				//	},
				//	func(i interface{}, s string) ([]string, []error) {
				//		return validation.IsIPv4Address(i.(string), "172.16.0.0/11")
				//	},
				//	func(i interface{}, s string) ([]string, []error) {
				//		return validation.IsIPv4Address(i.(string), "192.168.0.0/16")
				//	},
				//),
			},
			"bind_address": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "Network address for the private network that will be created for clients. This network cannot conflict with any existing local networks",
				Computed:    true,
				ForceNew:    false,
			},
			"organizations": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
				Required:    false,
				Optional:    true,
				Description: "The list of attached organizations for the server",
				ForceNew:    false,
			},
			"route": {
				Type: schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Network address with subnet to route",
							ForceNew:    false,
						},
						"comment": {
							Type:        schema.TypeString,
							Required:    false,
							Optional:    true,
							Description: "Comment for route",
							ForceNew:    false,
						},
						"nat": {
							Type:        schema.TypeBool,
							Required:    false,
							Optional:    true,
							Description: "NAT vpn traffic destined to this network",
							Computed:    true,
							ForceNew:    false,
						},
					},
				},
				Required:    false,
				Optional:    true,
				Computed:    true,
				Description: "The list of attached routes for the server",
				ForceNew:    false,
			},
			"status": {
				Type:         schema.TypeString,
				Required:     false,
				Optional:     true,
				Description:  "The status of the server",
				ForceNew:     false,
				Computed:     true,
				RequiredWith: []string{"organizations"},
				ValidateDiagFunc: func(v interface{}, path cty.Path) diag.Diagnostics {
					allowedStatusesMap := map[string]struct{}{
						ServerStatusOffline: {},
						ServerStatusOnline:  {},
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
	apiClient := meta.(Client)

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

	d.Set("name", server.Name)
	d.Set("protocol", server.Protocol)
	d.Set("port", server.Port)
	d.Set("cipher", server.Cipher)
	d.Set("hash", server.Hash)
	d.Set("network", server.Network)
	d.Set("bind_address", server.BindAddress)

	if len(organizations) > 0 {
		d.Set("organizations", flattenOrganizationsData(organizations))
	}

	if len(routes) > 0 {
		d.Set("route", flattenRoutesData(routes))
	}

	return nil
}

func resourceCreateServer(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(Client)

	serverData := map[string]interface{}{
		"name":         d.Get("name"),
		"protocol":     d.Get("protocol"),
		"port":         d.Get("port"),
		"network":      d.Get("network"),
		"cipher":       d.Get("cipher"),
		"hash":         d.Get("hash"),
		"bind_address": d.Get("bind_address"),
	}

	server, err := apiClient.CreateServer(serverData)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(server.ID)

	if d.HasChange("organizations") {
		_, newOrgs := d.GetChange("organizations")
		for _, v := range newOrgs.([]interface{}) {
			org := ConvertMapToOrganization(v.(map[string]interface{}))

			err = apiClient.AttachOrganizationToServer(org.ID, d.Id())
			if err != nil {
				return diag.Errorf("Error on attaching server to the organization: %s", err)
			}
		}
	}

	// Delete default route
	defaultRoute := Route{
		Network: "0.0.0.0/0",
		Nat:     true,
	}
	err = apiClient.DeleteRouteFromServer(d.Id(), defaultRoute)
	if err != nil {
		return diag.Errorf("Error on attaching server to the organization: %s", err)
	}

	if d.HasChange("route") {
		_, newRoutes := d.GetChange("route")
		routes := make([]Route, 0)

		for _, v := range newRoutes.(*schema.Set).List() {
			routes = append(routes, ConvertMapToRoute(v.(map[string]interface{})))
		}

		err = apiClient.AddRoutesToServer(d.Id(), routes)
		if err != nil {
			return diag.Errorf("Error on attaching route from the server: %s", err)
		}
	}

	if d.Get("status").(string) == ServerStatusOnline {
		err = apiClient.StartServer(d.Id())
		if err != nil {
			return diag.Errorf("Error on starting server: %s", err)
		}
	}

	return resourceReadServer(ctx, d, meta)
}

func resourceUpdateServer(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(Client)

	server, err := apiClient.GetServer(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

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

	if d.HasChange("status") {
		newStatus := d.Get("status").(string)
		if newStatus == ServerStatusOnline {
			err = apiClient.StartServer(d.Id())
			if err != nil {
				return diag.Errorf("Error on starting server: %s", err)
			}
		} else {
			err = apiClient.StopServer(d.Id())
			if err != nil {
				return diag.Errorf("Error on stopping server: %s", err)
			}
		}
	}

	if d.Get("status").(string) == ServerStatusOnline {
		err = apiClient.StopServer(d.Id())
		if err != nil {
			return diag.Errorf("Error on stopping server: %s", err)
		}
	}

	if d.HasChange("organizations") {
		oldOrgs, newOrgs := d.GetChange("organizations")
		for _, v := range oldOrgs.([]interface{}) {
			organization := ConvertMapToOrganization(v.(map[string]interface{}))

			err = apiClient.DetachOrganizationFromServer(organization.ID, d.Id())
			if err != nil {
				return diag.Errorf("Error on detaching server to the organization: %s", err)
			}
		}
		for _, v := range newOrgs.([]interface{}) {
			org := ConvertMapToOrganization(v.(map[string]interface{}))

			err = apiClient.AttachOrganizationToServer(org.ID, d.Id())
			if err != nil {
				return diag.Errorf("Error on attaching server to the organization: %s", err)
			}
		}
	}

	if d.HasChange("route") {
		oldRoutes, newRoutes := d.GetChange("route")

		newRoutesMap := make(map[string]Route, 0)
		for _, v := range newRoutes.(*schema.Set).List() {
			route := ConvertMapToRoute(v.(map[string]interface{}))
			newRoutesMap[route.GetID()] = route
		}
		oldRoutesMap := make(map[string]Route, 0)
		for _, v := range oldRoutes.(*schema.Set).List() {
			route := ConvertMapToRoute(v.(map[string]interface{}))
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

	err = apiClient.UpdateServer(d.Id(), server)
	if err != nil {
		// start server in case of error?
		return diag.FromErr(err)
	}

	if d.Get("status").(string) == ServerStatusOnline {
		err = apiClient.StartServer(d.Id())
		if err != nil {
			return diag.Errorf("Error on starting server: %s", err)
		}
	}

	return resourceReadServer(ctx, d, meta)
}

func resourceDeleteServer(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(Client)

	err := apiClient.DeleteServer(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func flattenOrganizationsData(organizationsList []Organization) []interface{} {
	organizations := make([]interface{}, 0)

	if organizationsList != nil {
		for _, organization := range organizationsList {
			orgMap := make(map[string]interface{})

			orgMap["id"] = organization.ID
			orgMap["name"] = organization.Name

			organizations = append(organizations, orgMap)
		}
	}

	return organizations
}

func flattenRoutesData(routesList []Route) []interface{} {
	routes := make([]interface{}, 0)

	if routesList != nil {
		for _, route := range routesList {
			if route.VirtualNetwork {
				// skip virtual network route
				continue
			}

			routeMap := make(map[string]interface{})

			routeMap["id"] = route.GetID()
			routeMap["network"] = route.Network
			routeMap["nat"] = strconv.FormatBool(route.Nat)
			if route.Comment != "" {
				routeMap["comment"] = route.Comment
			}

			routes = append(routes, routeMap)
		}
	}

	return routes
}
