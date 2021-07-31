package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-pritunl/internal/pritunl"
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
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "The protocol for the server",
				Default:     "udp",
				ForceNew:    false,
			},
			"cipher": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "The cipher for the server",
				Default:     "aes128",
				ForceNew:    false,
			},
			"hash": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Description: "The hash for the server",
				Default:     "sha1",
				ForceNew:    false,
			},
			"port": {
				Type:        schema.TypeInt,
				Required:    false,
				Optional:    true,
				Computed:    true,
				Description: "The port for the server",
				ForceNew:    false,
			},
			"organizations": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
				//Elem: &schema.Resource{
				//	Schema: resourceOrganization().Schema,
				//},
				Required:    false,
				Optional:    true,
				Description: "The list of attached organizations for the server",
				ForceNew:    false,
			},
			"routes": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
				Required:    false,
				Optional:    true,
				Description: "The list of attached routes for the server",
				ForceNew:    false,
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

	d.Set("name", server.Name)
	d.Set("protocol", server.Protocol)
	d.Set("cipher", server.Cipher)
	d.Set("hash", server.Hash)

	return nil
}

func resourceCreateServer(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	var port int
	if v, ok := d.GetOk("port"); ok {
		port = v.(int)
	}

	server, err := apiClient.CreateServer(
		d.Get("name").(string),
		d.Get("protocol").(string),
		d.Get("cipher").(string),
		d.Get("hash").(string),
		&port,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(server.ID)
	d.Set("port", server.Port)

	if d.HasChange("organizations") {
		_, newOrgs := d.GetChange("organizations")
		for _, v := range newOrgs.([]interface{}) {
			org := pritunl.ConvertMapToOrganization(v.(map[string]interface{}))

			err = apiClient.AttachOrganizationToServer(org.ID, d.Id())
			if err != nil {
				return diag.Errorf("Error on attaching server to the organization: %s", err)
			}
		}
	}

	if d.HasChange("routes") {
		_, newRoutes := d.GetChange("routes")
		routes := make([]pritunl.Route, 0)
		for _, v := range newRoutes.([]interface{}) {
			routes = append(routes, pritunl.ConvertMapToRoute(v.(map[string]interface{})))
		}

		err = apiClient.AddRoutesToServer(d.Id(), routes)
		if err != nil {
			return diag.Errorf("Error on attaching route from the server: %s", err)
		}
	}

	// Need to start server after a successful creation?

	return nil
}

func resourceUpdateServer(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

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

	if d.HasChange("organizations") {
		oldOrgs, newOrgs := d.GetChange("organizations")
		for _, v := range oldOrgs.([]interface{}) {
			organization := pritunl.ConvertMapToOrganization(v.(map[string]interface{}))

			err = apiClient.DetachOrganizationFromServer(organization.ID, d.Id())
			if err != nil {
				return diag.Errorf("Error on detaching server to the organization: %s", err)
			}
		}
		for _, v := range newOrgs.([]interface{}) {
			org := pritunl.ConvertMapToOrganization(v.(map[string]interface{}))

			err = apiClient.AttachOrganizationToServer(org.ID, d.Id())
			if err != nil {
				return diag.Errorf("Error on attaching server to the organization: %s", err)
			}
		}
	}

	if d.HasChange("routes") {
		oldRoutes, newRoutes := d.GetChange("routes")

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

	// Check if changes require stopping of the server
	// Check if server is running
	err = apiClient.StopServer(d.Id())
	if err != nil {
		return diag.Errorf("Error on stopping server: %s", err)
	}

	err = apiClient.UpdateServer(d.Id(), server)
	if err != nil {
		// start server in case of error?
		return diag.FromErr(err)
	}

	// Check if server is stopped
	err = apiClient.StartServer(d.Id())
	if err != nil {
		return diag.Errorf("Error on starting server: %s", err)
	}

	return nil
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
