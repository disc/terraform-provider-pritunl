package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

func resourceRoute() *schema.Resource {
	return &schema.Resource{
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
				Default:     true,
				ForceNew:    false,
			},
		},
		CreateContext: resourceCreateRoute,
		ReadContext:   resourceReadRoute,
		UpdateContext: resourceUpdateRoute,
		DeleteContext: resourceDeleteRoute,
		//Exists: resourceExistsRoute,
		//Importer: &schema.ResourceImporter{
		//	State: schema.ImportStatePassthrough,
		//},
	}
}

func resourceExistsRoute(d *schema.ResourceData, meta interface{}) (bool, error) {
	return d.Id() != "", nil
}

// Uses for importing
func resourceReadRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//TODO

	return nil
}

func resourceCreateRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	network := d.Get("network").(string)
	d.Set("network", network)
	d.Set("comment", d.Get("comment").(string))
	d.Set("nat", d.Get("nat").(bool))

	id := fmt.Sprintf("pritunl-route-%d", time.Now().Unix())
	d.SetId(id)

	return nil
}

func resourceUpdateRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.Set("network", d.Get("network").(string))
	d.Set("comment", d.Get("comment").(string))
	d.Set("nat", d.Get("nat").(bool))

	return nil
}

func resourceDeleteRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")

	return nil
}
