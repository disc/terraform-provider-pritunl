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
			"name": {
				Description: "Name of host",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"public_addr": {
				Description: "Public IP address or domain name of the host",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"public_addr6": {
				Description: "Public IPv6 address or domain name of the host",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"routed_subnet6": {
				Description: "IPv6 subnet that is routed to the host",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"routed_subnet6_wg": {
				Description: "IPv6 WG subnet that is routed to the host",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"local_addr": {
				Description: "Local network address for server",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"local_addr6": {
				Description: "Local IPv6 network address for server",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"availability_group": {
				Description: "Availability group for host. Replicated servers will only be replicated to a group of hosts in the same availability group\"",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"link_addr": {
				Description: "IP address or domain used when linked servers connect to a linked server on this host",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"sync_address": {
				Description: "IP address or domain used by users when syncing configuration. This is needed when using a load balancer.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"status": {
				Description: "Status of host",
				Type:        schema.TypeString,
				Computed:    true,
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
	d.Set("name", host.Name)
	d.Set("hostname", host.Hostname)
	d.Set("public_addr", host.PublicAddr)
	d.Set("public_addr6", host.PublicAddr6)
	d.Set("routed_subnet6", host.RoutedSubnet6)
	d.Set("routed_subnet6_wg", host.RoutedSubnet6WG)
	d.Set("local_addr", host.LocalAddr)
	d.Set("local_addr6", host.LocalAddr6)
	d.Set("link_addr", host.LinkAddr)
	d.Set("sync_address", host.SyncAddress)
	d.Set("availability_group", host.AvailabilityGroup)
	d.Set("status", host.Status)

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
