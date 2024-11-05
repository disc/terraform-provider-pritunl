package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceServer() *schema.Resource {
	return &schema.Resource{
		// ... existing schema ...
		Schema: map[string]*schema.Schema{
			// ... existing fields ...

			"wg": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable WireGuard protocol",
			},
			"port_wg": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "WireGuard port number",
			},
			"network_wg": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "WireGuard network subnet",
			},

			// ... rest of the schema ...
		},
	}
}
