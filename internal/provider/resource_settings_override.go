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
			"sso_cache": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Enable an 8 hour secondary authentication cache using client ID, IP address and MAC address. This will allow clients to reconnect without secondary authentication. Works with Duo push, Okta push, OneLogin push, Duo passcodes and YubiKeys. Supported by all OpenVPN clients",
			},
			"sso_client_cache": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Enable a two day secondary authentication cache using a token stored on the client. This will allow clients to reconnect without secondary authentication. Works with Duo push, Okta push, OneLogin push, Duo passcodes and YubiKeys. Only supported by Pritunl client",
			},
			"restrict_import": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Require users to use Pritunl URI when importing profiles",
			},
			"client_reconnect": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Enable auto reconnecting on Pritunl client",
			},
			"cloud_provider": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Cloud Provider",
				ValidateFunc: validation.StringInSlice([]string{"aws", "oracle"}, false),
			},
			"cloud_provider_aws": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				MaxItems:      1,
				ConflictsWith: []string{"cloud_provider_oracle"},
				Elem: &schema.Resource{
					Schema: cloudProviderAwsSchema,
				},
			},
			"cloud_provider_oracle": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				MaxItems:      1,
				ConflictsWith: []string{"cloud_provider_aws"},
				Elem: &schema.Resource{
					Schema: cloudProviderOracleSchema,
				},
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
	d.Set("sso_cache", settings.SSOCache)
	d.Set("sso_client_cache", settings.SSOClientCache)
	d.Set("restrict_import", settings.RestrictImport)
	d.Set("client_reconnect", settings.ClientReconnect)
	d.Set("cloud_provider", settings.CloudProvider)
	d.Set("cloud_provider_aws", []map[string]interface{}{
		{
			"route53_region":            settings.Route53Region,
			"route53_zone":              settings.Route53Zone,
			"us_east_1_access_key":      settings.AwsUsEast1AccessKey,
			"us_east_1_secret_key":      settings.AwsUsEast1SecretKey,
			"us_east_2_access_key":      settings.AwsUsEast2AccessKey,
			"us_east_2_secret_key":      settings.AwsUsEast2SecretKey,
			"us_west_1_access_key":      settings.AwsUsWest1AccessKey,
			"us_west_1_secret_key":      settings.AwsUsWest1SecretKey,
			"us_west_2_access_key":      settings.AwsUsWest2AccessKey,
			"us_west_2_secret_key":      settings.AwsUsWest2SecretKey,
			"us_gov_east_1_access_key":  settings.AwsUsGovEast1AccessKey,
			"us_gov_east_1_secret_key":  settings.AwsUsGovEast1SecretKey,
			"us_gov_west_1_access_key":  settings.AwsUsGovWest1AccessKey,
			"us_gov_west_1_secret_key":  settings.AwsUsGovWest1SecretKey,
			"eu_north_1_access_key":     settings.AwsEuNorth1AccessKey,
			"eu_north_1_secret_key":     settings.AwsEuNorth1SecretKey,
			"eu_west_1_access_key":      settings.AwsEuWest1AccessKey,
			"eu_west_1_secret_key":      settings.AwsEuWest1SecretKey,
			"eu_west_2_access_key":      settings.AwsEuWest2AccessKey,
			"eu_west_2_secret_key":      settings.AwsEuWest2SecretKey,
			"eu_west_3_access_key":      settings.AwsEuWest3AccessKey,
			"eu_west_3_secret_key":      settings.AwsEuWest3SecretKey,
			"eu_central_1_access_key":   settings.AwsEuCentral1AccessKey,
			"eu_central_1_secret_key":   settings.AwsEuCentral1SecretKey,
			"ca_central_1_access_key":   settings.AwsCaCentral1AccessKey,
			"ca_central_1_secret_key":   settings.AwsCaCentral1SecretKey,
			"cn_north_1_access_key":     settings.AwsCnNorth1AccessKey,
			"cn_north_1_secret_key":     settings.AwsCnNorth1SecretKey,
			"cn_northwest_1_access_key": settings.AwsCnNorthWest1AccessKey,
			"cn_northwest_1_secret_key": settings.AwsCnNorthWest1SecretKey,
			"ap_northeast_1_access_key": settings.AwsApNorthEast1AccessKey,
			"ap_northeast_1_secret_key": settings.AwsApNorthEast1SecretKey,
			"ap_northeast_2_access_key": settings.AwsApNorthEast2AccessKey,
			"ap_northeast_2_secret_key": settings.AwsApNorthEast2SecretKey,
			"ap_southeast_1_access_key": settings.AwsApSouthEast1AccessKey,
			"ap_southeast_1_secret_key": settings.AwsApSouthEast1SecretKey,
			"ap_southeast_2_access_key": settings.AwsApSouthEast2AccessKey,
			"ap_southeast_2_secret_key": settings.AwsApSouthEast2SecretKey,
			"ap_east_1_access_key":      settings.AwsApEast1AccessKey,
			"ap_east_1_secret_key":      settings.AwsApEast1SecretKey,
			"ap_south_1_access_key":     settings.AwsApSouth1AccessKey,
			"ap_south_1_secret_key":     settings.AwsApSouth1SecretKey,
			"sa_east_1_access_key":      settings.AwsSaEast1AccessKey,
			"sa_east_1_secret_key":      settings.AwsSaEast1SecretKey,
		},
	})
	d.Set("cloud_provider_oracle", []map[string]interface{}{
		{
			"oracle_user_ocid":  settings.OracleUserOcid,
			"oracle_public_key": settings.OraclePublicKey,
		},
	})

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

	if v, ok := d.GetOk("sso_cache"); ok {
		settings.SSOCache = v.(bool)
	}

	if v, ok := d.GetOk("sso_client_cache"); ok {
		settings.SSOClientCache = v.(bool)
	}

	if v, ok := d.GetOk("restrict_import"); ok {
		settings.RestrictImport = v.(bool)
	}

	if v, ok := d.GetOk("client_reconnect"); ok {
		settings.ClientReconnect = v.(bool)
	}

	if v, ok := d.GetOk("cloud_provider"); ok {
		settings.CloudProvider = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.route53_region"); ok {
		settings.Route53Region = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.route53_zone"); ok {
		settings.Route53Zone = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.us_east_1_access_key"); ok {
		settings.AwsUsEast1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.us_east_1_secret_key"); ok {
		settings.AwsUsEast1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.us_east_2_access_key"); ok {
		settings.AwsUsEast2AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.us_east_2_secret_key"); ok {
		settings.AwsUsEast2SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.us_west_1_access_key"); ok {
		settings.AwsUsWest1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.us_west_1_secret_key"); ok {
		settings.AwsUsWest1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.us_west_2_access_key"); ok {
		settings.AwsUsWest2AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.us_west_2_secret_key"); ok {
		settings.AwsUsWest2SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.us_gov_east_1_access_key"); ok {
		settings.AwsUsGovEast1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.us_gov_east_1_secret_key"); ok {
		settings.AwsUsGovEast1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.us_gov_west_1_access_key"); ok {
		settings.AwsUsGovWest1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.us_gov_west_1_secret_key"); ok {
		settings.AwsUsGovWest1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.eu_north_1_access_key"); ok {
		settings.AwsEuNorth1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.eu_north_1_secret_key"); ok {
		settings.AwsEuNorth1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.eu_west_1_access_key"); ok {
		settings.AwsEuWest1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.eu_west_1_secret_key"); ok {
		settings.AwsEuWest1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.eu_west_2_access_key"); ok {
		settings.AwsEuWest2AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.eu_west_2_secret_key"); ok {
		settings.AwsEuWest2SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.eu_west_3_access_key"); ok {
		settings.AwsEuWest3AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.eu_west_3_secret_key"); ok {
		settings.AwsEuWest3SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.eu_central_1_access_key"); ok {
		settings.AwsEuCentral1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.eu_central_1_secret_key"); ok {
		settings.AwsEuCentral1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.ca_central_1_access_key"); ok {
		settings.AwsCaCentral1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.ca_central_1_secret_key"); ok {
		settings.AwsCaCentral1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.cn_north_1_access_key"); ok {
		settings.AwsCnNorth1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.cn_north_1_secret_key"); ok {
		settings.AwsCnNorth1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.cn_northwest_1_access_key"); ok {
		settings.AwsCnNorthWest1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.cn_northwest_1_secret_key"); ok {
		settings.AwsCnNorthWest1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.ap_northeast_1_access_key"); ok {
		settings.AwsApNorthEast1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.ap_northeast_1_secret_key"); ok {
		settings.AwsApNorthEast1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.ap_northeast_2_access_key"); ok {
		settings.AwsApNorthEast2AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.ap_northeast_2_secret_key"); ok {
		settings.AwsApNorthEast2SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.ap_southeast_1_access_key"); ok {
		settings.AwsApSouthEast1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.ap_southeast_1_secret_key"); ok {
		settings.AwsApSouthEast1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.ap_southeast_2_access_key"); ok {
		settings.AwsApSouthEast2AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.ap_southeast_2_secret_key"); ok {
		settings.AwsApSouthEast2SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.ap_east_1_access_key"); ok {
		settings.AwsApEast1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.ap_east_1_secret_key"); ok {
		settings.AwsApEast1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.ap_south_1_access_key"); ok {
		settings.AwsApSouth1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.ap_south_1_secret_key"); ok {
		settings.AwsApSouth1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.sa_east_1_access_key"); ok {
		settings.AwsSaEast1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws.0.sa_east_1_secret_key"); ok {
		settings.AwsSaEast1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_oracle.0.oracle_user_ocid"); ok {
		settings.OracleUserOcid = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_oracle.0.oracle_public_key"); ok {
		settings.OraclePublicKey = v.(string)
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

	if d.HasChange("sso_cache") {
		settings.SSOCache = d.Get("sso_cache").(bool)
	}

	if d.HasChange("sso_client_cache") {
		settings.SSOClientCache = d.Get("sso_client_cache").(bool)
	}

	if d.HasChange("restrict_import") {
		settings.RestrictImport = d.Get("restrict_import").(bool)
	}

	if d.HasChange("client_reconnect") {
		settings.ClientReconnect = d.Get("client_reconnect").(bool)
	}

	if d.HasChange("cloud_provider") {
		settings.CloudProvider = d.Get("cloud_provider").(string)
	}

	if d.HasChange("cloud_provider_aws.0.route53_region") {
		settings.Route53Region = d.Get("cloud_provider_aws.0.route53_region").(string)
	}

	if d.HasChange("cloud_provider_aws.0.route53_zone") {
		settings.Route53Zone = d.Get("cloud_provider_aws.0.route53_zone").(string)
	}

	if d.HasChange("cloud_provider_aws.0.us_east_1_access_key") {
		settings.AwsUsEast1AccessKey = d.Get("cloud_provider_aws.0.us_east_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.us_east_1_secret_key") {
		settings.AwsUsEast1SecretKey = d.Get("cloud_provider_aws.0.us_east_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.us_east_2_access_key") {
		settings.AwsUsEast2AccessKey = d.Get("cloud_provider_aws.0.us_east_2_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.us_east_2_secret_key") {
		settings.AwsUsEast2SecretKey = d.Get("cloud_provider_aws.0.us_east_2_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.us_west_1_access_key") {
		settings.AwsUsWest1AccessKey = d.Get("cloud_provider_aws.0.us_west_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.us_west_1_secret_key") {
		settings.AwsUsWest1SecretKey = d.Get("cloud_provider_aws.0.us_west_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.us_west_2_access_key") {
		settings.AwsUsWest2AccessKey = d.Get("cloud_provider_aws.0.us_west_2_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.us_west_2_secret_key") {
		settings.AwsUsWest2SecretKey = d.Get("cloud_provider_aws.0.us_west_2_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.us_gov_east_1_access_key") {
		settings.AwsUsGovEast1AccessKey = d.Get("cloud_provider_aws.0.us_gov_east_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.us_gov_east_1_secret_key") {
		settings.AwsUsGovEast1SecretKey = d.Get("cloud_provider_aws.0.us_gov_east_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.us_gov_west_1_access_key") {
		settings.AwsUsGovWest1AccessKey = d.Get("cloud_provider_aws.0.us_gov_west_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.us_gov_west_1_secret_key") {
		settings.AwsUsGovWest1SecretKey = d.Get("cloud_provider_aws.0.us_gov_west_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.eu_north_1_access_key") {
		settings.AwsEuNorth1AccessKey = d.Get("cloud_provider_aws.0.eu_north_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.eu_north_1_secret_key") {
		settings.AwsEuNorth1SecretKey = d.Get("cloud_provider_aws.0.eu_north_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.eu_west_1_access_key") {
		settings.AwsEuWest1AccessKey = d.Get("cloud_provider_aws.0.eu_west_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.eu_west_1_secret_key") {
		settings.AwsEuWest1SecretKey = d.Get("cloud_provider_aws.0.eu_west_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.eu_west_2_access_key") {
		settings.AwsEuWest2AccessKey = d.Get("cloud_provider_aws.0.eu_west_2_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.eu_west_2_secret_key") {
		settings.AwsEuWest2SecretKey = d.Get("cloud_provider_aws.0.eu_west_2_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.eu_west_3_access_key") {
		settings.AwsEuWest3AccessKey = d.Get("cloud_provider_aws.0.eu_west_3_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.eu_west_3_secret_key") {
		settings.AwsEuWest3SecretKey = d.Get("cloud_provider_aws.0.eu_west_3_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.eu_central_1_access_key") {
		settings.AwsEuCentral1AccessKey = d.Get("cloud_provider_aws.0.eu_central_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.eu_central_1_secret_key") {
		settings.AwsEuCentral1SecretKey = d.Get("cloud_provider_aws.0.eu_central_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.ca_central_1_access_key") {
		settings.AwsCaCentral1AccessKey = d.Get("cloud_provider_aws.0.ca_central_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.ca_central_1_secret_key") {
		settings.AwsCaCentral1SecretKey = d.Get("cloud_provider_aws.0.ca_central_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.cn_north_1_access_key") {
		settings.AwsCnNorth1AccessKey = d.Get("cloud_provider_aws.0.cn_north_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.cn_north_1_secret_key") {
		settings.AwsCnNorth1SecretKey = d.Get("cloud_provider_aws.0.cn_north_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.cn_northwest_1_access_key") {
		settings.AwsCnNorthWest1AccessKey = d.Get("cloud_provider_aws.0.cn_northwest_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.cn_northwest_1_secret_key") {
		settings.AwsCnNorthWest1SecretKey = d.Get("cloud_provider_aws.0.cn_northwest_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.ap_northeast_1_access_key") {
		settings.AwsApNorthEast1AccessKey = d.Get("cloud_provider_aws.0.ap_northeast_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.ap_northeast_1_secret_key") {
		settings.AwsApNorthEast1SecretKey = d.Get("cloud_provider_aws.0.ap_northeast_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.ap_northeast_2_access_key") {
		settings.AwsApNorthEast2AccessKey = d.Get("cloud_provider_aws.0.ap_northeast_2_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.ap_northeast_2_secret_key") {
		settings.AwsApNorthEast2SecretKey = d.Get("cloud_provider_aws.0.ap_northeast_2_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.ap_southeast_1_access_key") {
		settings.AwsApSouthEast1AccessKey = d.Get("cloud_provider_aws.0.ap_southeast_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.ap_southeast_1_secret_key") {
		settings.AwsApSouthEast1SecretKey = d.Get("cloud_provider_aws.0.ap_southeast_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.ap_southeast_2_access_key") {
		settings.AwsApSouthEast2AccessKey = d.Get("cloud_provider_aws.0.ap_southeast_2_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.ap_southeast_2_secret_key") {
		settings.AwsApSouthEast2SecretKey = d.Get("cloud_provider_aws.0.ap_southeast_2_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.ap_east_1_access_key") {
		settings.AwsApEast1AccessKey = d.Get("cloud_provider_aws.0.ap_east_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.ap_east_1_secret_key") {
		settings.AwsApEast1SecretKey = d.Get("cloud_provider_aws.0.ap_east_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.ap_south_1_access_key") {
		settings.AwsApSouth1AccessKey = d.Get("cloud_provider_aws.0.ap_south_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.ap_south_1_secret_key") {
		settings.AwsApSouth1SecretKey = d.Get("cloud_provider_aws.0.ap_south_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.sa_east_1_access_key") {
		settings.AwsSaEast1AccessKey = d.Get("cloud_provider_aws.0.sa_east_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws.0.sa_east_1_secret_key") {
		settings.AwsSaEast1SecretKey = d.Get("cloud_provider_aws.0.sa_east_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_oracle.0.oracle_user_ocid") {
		settings.OracleUserOcid = d.Get("cloud_provider_oracle.0.oracle_user_ocid").(string)
	}

	if d.HasChange("cloud_provider_oracle.0.oracle_public_key") {
		settings.OraclePublicKey = d.Get("cloud_provider_oracle.0.oracle_public_key").(string)
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

var cloudProviderOracleSchema = map[string]*schema.Schema{
	"oracle_user_ocid": {
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "The OCID of the Oracle Cloud user that is associated with the public key",
	},
	"oracle_public_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "Generated public key for Oracle API. Copy this public key to the user API keys in the Oracle Cloud console. Clear the input field to generate a new key.",
	},
}

var cloudProviderAwsSchema = map[string]*schema.Schema{
	"route53_region": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Route 53 region: This will automatically create and update DNS records for each host in the selected region",
		ValidateFunc: validation.StringInSlice([]string{
			"us-east-1",
			"us-east-2",
			"us-west-1",
			"us-west-2",
			"us-gov-east-1",
			"us-gov-west-1",
			"eu-north-1",
			"eu-west-1",
			"eu-west-2",
			"eu-west-3",
			"eu-central-1",
			"cn-north-1",
			"cn-northwest-1",
			"ca-central-1",
			"ap-northeast-1",
			"ap-northeast-2",
			"ap-southeast-1",
			"ap-southeast-2",
			"ap-east-1",
			"ap-south-1",
			"sa-east-1",
		}, false),
	},
	"route53_zone": {
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "Route 53 zone: This will automatically create and update DNS records for each host in the selected zone. The name of the host will be used for the subdomain in the zone. The AWS keys below must be saved before a list of available zones will be shown.",
		ValidateFunc: validation.StringInSlice([]string{"aws", "oracle"}, false),
	},
	"us_east_1_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "US East (N. Virginia) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"us_east_1_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "US East (N. Virginia) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"us_east_2_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "US East (Ohio) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"us_east_2_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "US East (Ohio) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"us_west_1_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "US West (N. California) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"us_west_1_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "US West (N. California) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"us_west_2_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "US West (Oregon) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"us_west_2_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "US West (Oregon) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"us_gov_east_1_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "US GovCloud (East) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"us_gov_east_1_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "US GovCloud (East) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"us_gov_west_1_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "US GovCloud (West) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"us_gov_west_1_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "US GovCloud (West) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"eu_north_1_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "EU (Stockholm) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"eu_north_1_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "EU (Stockholm) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"eu_west_1_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "EU (Ireland) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"eu_west_1_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "EU (Ireland) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"eu_west_2_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "EU (London) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"eu_west_2_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "EU (London) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"eu_west_3_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "EU (Paris) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"eu_west_3_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "EU (Paris) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"eu_central_1_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "EU (Frankfurt) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"eu_central_1_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "EU (Frankfurt) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"ca_central_1_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Canada (Central) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"ca_central_1_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Canada (Central) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"cn_north_1_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "China (Beijing) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"cn_north_1_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "China (Beijing) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"cn_northwest_1_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "China (Ningxia) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"cn_northwest_1_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "China (Ningxia) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"ap_northeast_1_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Asia Pacific (Tokyo) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"ap_northeast_1_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Asia Pacific (Tokyo) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"ap_northeast_2_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Asia Pacific (Seoul) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"ap_northeast_2_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Asia Pacific (Seoul) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"ap_southeast_1_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Asia Pacific (Singapore) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"ap_southeast_1_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Asia Pacific (Singapore) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"ap_southeast_2_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Asia Pacific (Sydney) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"ap_southeast_2_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Asia Pacific (Sydney) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"ap_east_1_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Asia Pacific (Hong Kong) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"ap_east_1_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Asia Pacific (Hong Kong) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"ap_south_1_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Asia Pacific (Mumbai) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"ap_south_1_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Asia Pacific (Mumbai) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"sa_east_1_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "South America (Sao Paulo) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"sa_east_1_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "South America (Sao Paulo) Secret key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
}
