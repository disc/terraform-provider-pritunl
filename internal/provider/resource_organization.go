package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-pritunl/internal/pritunl"
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
		//Exists: resourceExistsOrganization,
		//Importer: &schema.ResourceImporter{
		//	State: schema.ImportStatePassthrough,
		//},
	}
}

func resourceExistsOrganization(d *schema.ResourceData, meta interface{}) (bool, error) {
	apiClient := meta.(pritunl.Client)

	organization, err := apiClient.GetOrganization(d.Id())
	if err != nil {
		return false, err
	}

	return organization != nil, nil
}

// Uses for importing
func resourceReadOrganization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	organization, err := apiClient.GetOrganization(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", organization.Name)

	return nil
}

func resourceDeleteOrganization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	err := apiClient.DeleteOrganization(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func resourceUpdateOrganization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

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
	apiClient := meta.(pritunl.Client)

	organization, err := apiClient.CreateOrganization(d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(organization.ID)

	return nil
}

//func validateName(v interface{}, k string) (ws []string, es []error) {
//	var errs []error
//	var warns []string
//	value, ok := v.(string)
//	if !ok {
//		errs = append(errs, fmt.Errorf("Expected name to be string"))
//		return warns, errs
//	}
//	whiteSpace := regexp.MustCompile(`\s+`)
//	if whiteSpace.Match([]byte(value)) {
//		errs = append(errs, fmt.Errorf("name cannot contain whitespace. Got %s", value))
//		return warns, errs
//	}
//	return warns, errs
//}
