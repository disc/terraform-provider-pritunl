package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/disc/terraform-provider-pritunl/internal/pritunl"
)

func dataSourceUserKeyUrls() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get Pritunl user keys (profile) temporary download URLs.",
		ReadContext: dataSourceUserKeyUrlsRead,
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key_url": {
				Description: "Keys (profile) tarball temporary download URL",
				Type:        schema.TypeString,
				Sensitive:   true,
				Computed:    true,
			},
			"zip_url": {
				Description: "ZIP archived profile temporary download URL",
				Type:        schema.TypeString,
				Sensitive:   true,
				Computed:    true,
			},
			"onc_url": {
				Description: "Chromebook profile temporary download URL",
				Type:        schema.TypeString,
				Sensitive:   true,
				Computed:    true,
			},
			"view_url": {
				Description: "Temporary URL to view profile links",
				Type:        schema.TypeString,
				Sensitive:   true,
				Computed:    true,
			},
			"uri_url": {
				Description: "Temporary URI path for Pritunl Client",
				Type:        schema.TypeString,
				Sensitive:   true,
				Computed:    true,
			},
		},
	}
}

func dataSourceUserKeyUrlsRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	userId := d.Get("user_id").(string)
	orgId := d.Get("organization_id").(string)

	key, err := apiClient.GetUserKeyUrls(userId, orgId)
	if err != nil {
		return diag.Errorf("could not fetch keys for user %s in organization %s. Previous error message: %v", userId, orgId, err)
	}

	d.SetId(key.ID)

	err = d.Set("key_url", key.KeyUrl)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("zip_url", key.KeyZipUrl)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("onc_url", key.KeyOncURL)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("view_url", key.ViewUrl)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("uri_url", key.UriUrl)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
