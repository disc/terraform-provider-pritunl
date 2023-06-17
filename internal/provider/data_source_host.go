package provider

import (
	"context"
	"errors"
	"github.com/disc/terraform-provider-pritunl/internal/pritunl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceHost() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to get information about the Pritunl hosts.",
		ReadContext: dataSourceHostRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hostname": {
				Description: "Hostname",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceHostRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	hostname := d.Get("hostname")
	filterFunction := func(host pritunl.Host) bool {
		return host.Hostname == hostname
	}

	host, err := filterHosts(meta, filterFunction)
	if err != nil {
		return diag.Errorf("could not find host with a hostname %s. Previous error message: %v", hostname, err)
	}

	d.SetId(host.ID)
	d.Set("hostname", host.Hostname)

	return nil
}

func filterHosts(meta interface{}, test func(host pritunl.Host) bool) (pritunl.Host, error) {
	apiClient := meta.(pritunl.Client)

	hosts, err := apiClient.GetHosts()

	if err != nil {
		return pritunl.Host{}, err
	}

	for _, dir := range hosts {
		if test(dir) {
			return dir, nil
		}
	}

	return pritunl.Host{}, errors.New("could not find a host with specified parameters")
}
