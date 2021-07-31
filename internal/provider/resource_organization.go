package provider

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"regexp"
	"terraform-pritunl/internal/pritunl"
)

func resourceOrganization() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The name of the resource, also acts as it's unique ID",
				ForceNew:     false,
				ValidateFunc: validateName,
			},
		},
		Create: resourceCreateOrganization,
		Read:   resourceReadOrganization,
		Update: resourceUpdateOrganization,
		Delete: resourceDeleteOrganization,
		Exists: resourceExistsOrganization,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
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

func resourceReadOrganization(d *schema.ResourceData, meta interface{}) error {
	apiClient := meta.(pritunl.Client)

	organization, err := apiClient.GetOrganization(d.Id())
	if err != nil {
		return err
	}

	id := ""
	if organization != nil {
		id = organization.ID
	}

	d.SetId(id)

	return nil
}

func resourceDeleteOrganization(d *schema.ResourceData, meta interface{}) error {
	apiClient := meta.(pritunl.Client)

	err := apiClient.DeleteOrganization(d.Id())
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func resourceUpdateOrganization(d *schema.ResourceData, meta interface{}) error {
	apiClient := meta.(pritunl.Client)

	organization, err := apiClient.GetOrganization(d.Id())
	if err != nil {
		return err
	}

	if d.HasChange("name") {
		organization.Name = d.Get("name").(string)

		err = apiClient.UpdateOrganization(d.Id(), organization)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceCreateOrganization(d *schema.ResourceData, meta interface{}) error {
	apiClient := meta.(pritunl.Client)

	organization, err := apiClient.CreateOrganization(d.Get("name").(string))
	if err != nil {
		return err
	}

	d.SetId(organization.ID)

	return nil
}

func validateName(v interface{}, k string) (ws []string, es []error) {
	var errs []error
	var warns []string
	value, ok := v.(string)
	if !ok {
		errs = append(errs, fmt.Errorf("Expected name to be string"))
		return warns, errs
	}
	whiteSpace := regexp.MustCompile(`\s+`)
	if whiteSpace.Match([]byte(value)) {
		errs = append(errs, fmt.Errorf("name cannot contain whitespace. Got %s", value))
		return warns, errs
	}
	return warns, errs
}
