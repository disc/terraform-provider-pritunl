package provider

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"terraform-pritunl/internal/pritunl"
)

func resourceServer() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The name of the server",
				ForceNew:     false,
				ValidateFunc: validateName,
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
					Type: schema.TypeString,
				},
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
		Create: resourceCreateServer,
		Read:   resourceReadServer,
		Update: resourceUpdateServer,
		Delete: resourceDeleteServer,
		Exists: resourceExistsServer,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceExistsServer(d *schema.ResourceData, meta interface{}) (bool, error) {
	apiClient := meta.(pritunl.Client)

	server, err := apiClient.GetServer(d.Id())
	if err != nil {
		return false, err
	}

	return server != nil, nil
}

func resourceReadServer(d *schema.ResourceData, meta interface{}) error {
	apiClient := meta.(pritunl.Client)

	_, err := apiClient.GetServer(d.Id())
	if err != nil {
		return err
	}

	d.Set("name", d.Get("name").(string))
	d.Set("protocol", d.Get("protocol").(string))
	d.Set("cipher", d.Get("cipher").(string))
	d.Set("hash", d.Get("hash").(string))

	return nil
}

func resourceCreateServer(d *schema.ResourceData, meta interface{}) error {
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
		return err
	}

	d.SetId(server.ID)
	d.Set("port", server.Port)

	return nil
}

func resourceUpdateServer(d *schema.ResourceData, meta interface{}) error {
	apiClient := meta.(pritunl.Client)

	server, err := apiClient.GetServer(d.Id())
	if err != nil {
		return err
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
			err = apiClient.DetachOrganizationFromServer(v.(string), d.Id())
			if err != nil {
				return fmt.Errorf("Error on detaching server to the organization: %s", err)
			}
		}
		for _, v := range newOrgs.([]interface{}) {
			err = apiClient.AttachOrganizationToServer(v.(string), d.Id())
			if err != nil {
				return fmt.Errorf("Error on attaching server to the organization: %s", err)
			}
		}
	}

	if d.HasChange("routes") {
		err = apiClient.StopServer(d.Id())
		if err != nil {
			return fmt.Errorf("Error on stopping server: %s", err)
		}

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
					return fmt.Errorf("Error on updating route on the server: %s", err)
				}
			} else {
				// add route
				err = apiClient.AddRouteToServer(d.Id(), route)
				if err != nil {
					return fmt.Errorf("Error on attaching route from the server: %s", err)
				}
			}
		}

		for _, route := range oldRoutesMap {
			if _, found := newRoutesMap[route.GetID()]; !found {
				// delete route
				err = apiClient.DeleteRouteFromServer(d.Id(), route)
				if err != nil {
					return fmt.Errorf("Error on detaching route from the server: %s", err)
				}
			}
		}

		err = apiClient.StartServer(d.Id())
		if err != nil {
			return fmt.Errorf("Error on starting server: %s", err)
		}
	}

	err = apiClient.UpdateServer(d.Id(), server)
	if err != nil {
		return err
	}

	return nil
}

func resourceDeleteServer(d *schema.ResourceData, meta interface{}) error {
	apiClient := meta.(pritunl.Client)

	err := apiClient.DeleteServer(d.Id())
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}
