package provider

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"regexp"
	"strings"
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

	organizationName := d.Id()
	_, err := apiClient.GetOrganization(organizationName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func resourceReadOrganization(d *schema.ResourceData, meta interface{}) error {
	apiClient := meta.(pritunl.Client)

	organization, err := apiClient.GetOrganization(d.Get("name").(string))
	if err != nil {
		return err
	}

	name := ""
	if organization != nil {
		name = organization.Name
	}

	d.SetId(name)

	return nil
}

func resourceDeleteOrganization(d *schema.ResourceData, meta interface{}) error {
	apiClient := meta.(pritunl.Client)

	organization, err := apiClient.GetOrganization(d.Get("name").(string))
	if err != nil {
		return err
	}

	err = apiClient.DeleteOrganization(organization.ID)
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func resourceUpdateOrganization(d *schema.ResourceData, meta interface{}) error {
	apiClient := meta.(pritunl.Client)

	organization, err := apiClient.GetOrganization(d.Get("name").(string))
	if err != nil {
		return err
	}

	newName := d.Get("new_name").(string)
	err = apiClient.RenameOrganization(organization.ID, newName)
	if err != nil {
		return err
	}

	d.SetId(newName)

	return nil
}

func resourceCreateOrganization(d *schema.ResourceData, meta interface{}) error {
	apiClient := meta.(pritunl.Client)

	organization, err := apiClient.CreateOrganization(d.Get("name").(string))
	if err != nil {
		return err
	}

	d.SetId(organization.Name)

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
