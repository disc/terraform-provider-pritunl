package provider

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"terraform-pritunl/internal/pritunl"
)

/*
 {"port_wg": null, "dns_servers": ["8.8.8.8"], "protocol": "tcp", "max_devices": 0, "max_clients": 2000, "link_ping_timeout": 5, "ping_timeout": 60, "ipv6": false, "vxlan": true, "network_mode": "tunnel", "bind_address": "", "block_outside_dns": false, "network_start": "", "name": "Alice-TCPnoTLS", "ping_interval": 10, "allowed_devices": null, "users_online": 1, "ipv6_firewall": true, "session_timeout": null, "otp_auth": false, "multi_device": false, "search_domain": null, "lzo_compression": "adaptive", "pre_connect_msg": null, "inactive_timeout": null, "link_ping_interval": 1, "id": "60d06624c36cc9d1d673304b", "ping_timeout_wg": 360, "uptime": 1295821, "network_end": "", "network": "192.168.249.0/24", "dh_param_bits": 2048, "wg": false, "port": 17490, "devices_online": 1, "network_wg": null, "status": "online", "dns_mapping": false, "hash": "sha1", "debug": false, "restrict_routes": true, "user_count": 1, "groups": [], "inter_client": true, "replica_count": 1, "cipher": "aes128", "mss_fix": null, "jumbo_frames": false}
*/
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
