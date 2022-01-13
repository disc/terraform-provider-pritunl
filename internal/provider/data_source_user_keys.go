package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/disc/terraform-provider-pritunl/internal/pritunl"
)

func dataSourceUserKeys() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get Pritunl user keys (profile).",
		ReadContext: dataSourceUserKeysRead,
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"configuration": {
				Description: "OpenVPN configuration files",
				Type:        schema.TypeMap,
				Elem:        schema.TypeString,
				Sensitive:   true,
				Computed:    true,
			},
		},
	}
}

func dataSourceUserKeysRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	userId := d.Get("user_id").(string)
	orgId := d.Get("organization_id").(string)

	keys, err := apiClient.GetUserKeys(userId, orgId)
	if err != nil {
		return diag.Errorf("could not fetch keys for user %s in organization %s. Previous error message: %v", userId, orgId, err)
	}
	err = d.Set("configuration", keys)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
