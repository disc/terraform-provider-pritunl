package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-pritunl/internal/pritunl"
)

func datasourceRoute() *schema.Resource {
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
		ReadContext: datasourceReadRoute,
	}
}

func datasourceReadRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	route := &pritunl.Route{}

	// required
	if v, ok := d.GetOk("network"); ok {
		route.Network = v.(string)
	}
	// optional
	if v, ok := d.GetOk("port"); ok {
		route.Network = v.(string)
	}
	if v, ok := d.GetOk("Nat"); ok {
		route.Nat = v.(bool)
	}
	if v, ok := d.GetOk("comment"); ok {
		route.Comment = v.(string)
	}

	d.SetId(route.GetID())

	return nil
}
