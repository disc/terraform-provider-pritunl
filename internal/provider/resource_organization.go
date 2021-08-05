package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOrganization() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
				ForceNew:    false,
			},
		},
		CreateContext: resourceCreateOrganization,
		ReadContext:   resourceReadOrganization,
		UpdateContext: resourceUpdateOrganization,
		DeleteContext: resourceDeleteOrganization,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

// Uses for importing
func resourceReadOrganization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(Client)

	organization, err := apiClient.GetOrganization(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", organization.Name)

	return nil
}

func resourceDeleteOrganization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(Client)

	err := apiClient.DeleteOrganization(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func resourceUpdateOrganization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(Client)

	organization, err := apiClient.GetOrganization(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("name") {
		organization.Name = d.Get("name").(string)

		err = apiClient.UpdateOrganization(d.Id(), organization)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceCreateOrganization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(Client)

	organization, err := apiClient.CreateOrganization(d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(organization.ID)

	return nil
}
