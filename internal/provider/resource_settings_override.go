package provider

import (
	"context"
	"github.com/disc/terraform-provider-pritunl/internal/pritunl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceSettingsOverride() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Admin username",
			},
			"theme": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Pritunl color theme",
				Default:      "light",
				ValidateFunc: validation.StringInSlice([]string{"dark", "light"}, false),
			},
			// Auditing cannot be disabled from web console.
			//"auditing": {
			//	Type:         schema.TypeString,
			//	Optional:     true,
			//	Computed:     true,
			//	Description:  "Auditing mode. Enable to log user actions such as login attempts and profile downloads",
			//	ValidateFunc: validation.StringInSlice([]string{"all", "none"}, false),
			//},
			"monitoring": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "Enable to send performance and usage metrics to InfluxDB",
				ValidateFunc: validation.StringInSlice([]string{"influxdb", "none"}, false),
			},
			"pin_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "Pin mode",
				ValidateFunc: validation.StringInSlice([]string{"optional", "required", "disabled"}, false),
			},
			// If you change the port don't forget to update the port in the provider's url as well
			//provider "pritunl" {
			//	url    = var.pritunl_url // <--
			//	...
			//}
			"server_port": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Web console port",
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"acme_domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Web console domain",
				ValidateFunc: func(i interface{}, s string) ([]string, []error) {
					return validation.StringIsNotEmpty(i, s)
				},
			},
			"reverse_proxy": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Allow reading client IP address from reverse proxy header. Enable when using services such as CloudFlare or when using a load balancer",
			},
			"sso_yubico_client": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Yubico Client ID",
			},
			"sso_yubico_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Yubico Secret Key",
			},
		},
		CreateContext: resourceCreateSettingsOverride,
		ReadContext:   resourceReadSettingsOverride,
		UpdateContext: resourceUpdateSettingsOverride,
		DeleteContext: resourceDeleteSettingsOverride,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

// Uses for importing
func resourceReadSettingsOverride(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	settings, err := apiClient.GetSettings()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("settings")
	d.Set("username", settings.Username)
	d.Set("theme", settings.Theme)
	d.Set("monitoring", settings.Monitoring)
	d.Set("pin_mode", settings.PinMode)
	d.Set("server_port", settings.ServerPort)
	d.Set("acme_domain", settings.AcmeDomain)
	d.Set("reverse_proxy", settings.ReverseProxy)
	d.Set("sso_yubico_client", settings.SSOYubicoClient)
	d.Set("sso_yubico_secret", settings.SSOYubicoSecret)

	return nil
}

func resourceCreateSettingsOverride(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	settings, err := apiClient.GetSettings()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("settings")

	if v, ok := d.GetOk("username"); ok {
		settings.Username = v.(string)
	}

	if v, ok := d.GetOk("theme"); ok {
		settings.Theme = v.(string)
	}

	if v, ok := d.GetOk("monitoring"); ok {
		settings.Monitoring = v.(string)
	}

	if v, ok := d.GetOk("pin_mode"); ok {
		settings.PinMode = v.(string)
	}

	if v, ok := d.GetOk("server_port"); ok {
		settings.ServerPort = v.(int)
	}

	if v, ok := d.GetOk("acme_domain"); ok {
		settings.AcmeDomain = v.(string)
	}

	if v, ok := d.GetOk("reverse_proxy"); ok {
		settings.ReverseProxy = v.(bool)
	}

	if v, ok := d.GetOk("sso_yubico_client"); ok {
		settings.SSOYubicoClient = v.(string)
	}

	if v, ok := d.GetOk("sso_yubico_secret"); ok {
		settings.SSOYubicoSecret = v.(string)
	}

	err = apiClient.UpdateSettings(settings)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceReadSettingsOverride(ctx, d, meta)
}

func resourceUpdateSettingsOverride(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	settings, err := apiClient.GetSettings()
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("username") {
		settings.Username = d.Get("username").(string)
	}

	if d.HasChange("theme") {
		settings.Theme = d.Get("theme").(string)
	}

	if d.HasChange("monitoring") {
		settings.Monitoring = d.Get("monitoring").(string)
	}

	if d.HasChange("pin_mode") {
		settings.PinMode = d.Get("pin_mode").(string)
	}

	if d.HasChange("server_port") {
		settings.ServerPort = d.Get("server_port").(int)
	}

	if d.HasChange("acme_domain") {
		settings.AcmeDomain = d.Get("acme_domain").(string)
	}

	if d.HasChange("reverse_proxy") {
		settings.ReverseProxy = d.Get("reverse_proxy").(bool)
	}

	if d.HasChange("sso_yubico_client") {
		settings.SSOYubicoClient = d.Get("sso_yubico_client").(string)
	}

	if d.HasChange("sso_yubico_secret") {
		settings.SSOYubicoSecret = d.Get("sso_yubico_secret").(string)
	}

	err = apiClient.UpdateSettings(settings)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceReadSettingsOverride(ctx, d, meta)
}

func resourceDeleteSettingsOverride(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
