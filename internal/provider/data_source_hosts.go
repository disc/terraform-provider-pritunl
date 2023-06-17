package provider

import (
	"context"
	"github.com/disc/terraform-provider-pritunl/internal/pritunl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceHosts() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get a list of the Pritunl hosts.",
		ReadContext: dataSourceHostsRead,
		Schema: map[string]*schema.Schema{
			"hosts": {
				Description: "A list of the Pritunl hosts resources.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: dataSourceHost().Schema,
				},
			},
		},
	}
}

func dataSourceHostsRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	hosts, err := apiClient.GetHosts()
	if err != nil {
		return diag.Errorf("could not find any host. Previous error message: %v", err)
	}

	var resultHosts []interface{}

	for _, host := range hosts {
		resultHosts = append(resultHosts, flattenHost(&host))
	}

	if err = d.Set("hosts", resultHosts); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("hosts")

	return nil
}

func flattenHost(host *pritunl.Host) interface{} {
	result := map[string]interface{}{}

	result["id"] = host.ID
	result["hostname"] = host.Hostname

	return result
}
