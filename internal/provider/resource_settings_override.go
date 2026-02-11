package provider

import (
	"context"
	"strings"

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
			"email_from": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Email from address",
			},
			"email_server": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Email server address",
			},
			"email_username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Email server username",
			},
			"email_password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Email server password",
			},
			"influxdb_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "InfluxDB URL",
			},
			"influxdb_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "InfluxDB token",
			},
			"influxdb_org": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "InfluxDB organization",
			},
			"influxdb_bucket": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "InfluxDB bucket",
			},
			"server_cert": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Web console SSL certificate",
			},
			"server_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Description: "Web console SSL private key",
			},
			"public_address": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Public IPv4 address",
			},
			"public_address6": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Public IPv6 address",
			},
			"routed_subnet6": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Routed IPv6 subnet",
			},
			"routed_subnet6_wg": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Routed IPv6 subnet for WireGuard",
			},
			"ipv6": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Enable IPv6",
			},
			"drop_permissions": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Drop permissions after starting server",
			},
			"restrict_client": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Restrict client options",
			},
			"sso": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Single Sign-On Mode",
				ValidateFunc: validation.StringInSlice([]string{
					"saml_okta",
					"saml_okta_duo",
					"saml_okta_yubico",
					"saml_onelogin",
					"saml_onelogin_duo",
					"saml_onelogin_yubico",
					"authzero",
					"authzero_duo",
					"authzero_yubico",
					"slack",
					"slack_duo",
					"slack_yubico",
					"google",
					"google_duo",
					"google_yubico",
					"azure",
					"azure_duo",
					"azure_yubico",
					"saml",
					"saml_duo",
					"saml_yubico",
					"duo",
					"radius",
					"radius_duo",
					"jumpcloud",
					"jumpcloud_duo",
					"jumpcloud_yubico",
				}, false),
			},
			"sso_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: ssoSettingsSchema,
				},
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
				ValidateFunc: validation.StringInSlice([]string{"aws", "oracle", "pritunl"}, false),
			},
			"cloud_provider_aws_settings": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				RequiredWith:  []string{"cloud_provider"},
				ConflictsWith: []string{"cloud_provider_oracle_settings", "cloud_provider_pritunl_settings"},
				Elem: &schema.Resource{
					Schema: cloudProviderAwsSchema,
				},
			},
			"cloud_provider_oracle_settings": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				RequiredWith:  []string{"cloud_provider"},
				ConflictsWith: []string{"cloud_provider_aws_settings", "cloud_provider_pritunl_settings"},
				Elem: &schema.Resource{
					Schema: cloudProviderOracleSchema,
				},
			},
			"cloud_provider_pritunl_settings": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				RequiredWith:  []string{"cloud_provider"},
				ConflictsWith: []string{"cloud_provider_aws_settings", "cloud_provider_oracle_settings"},
				Elem: &schema.Resource{
					Schema: cloudProviderPritunlSchema,
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
	d.Set("restrict_import", settings.RestrictImport)
	d.Set("client_reconnect", settings.ClientReconnect)
	d.Set("cloud_provider", settings.CloudProvider)

	d.Set("email_from", settings.EmailFrom)
	d.Set("email_server", settings.EmailServer)
	d.Set("email_username", settings.EmailUsername)
	if settings.EmailPassword != "" {
		d.Set("email_password", settings.EmailPassword)
	}

	d.Set("influxdb_url", settings.InfluxdbUrl)
	if settings.InfluxdbToken != "" {
		d.Set("influxdb_token", settings.InfluxdbToken)
	}
	d.Set("influxdb_org", settings.InfluxdbOrg)
	d.Set("influxdb_bucket", settings.InfluxdbBucket)

	d.Set("server_cert", settings.ServerCert)
	if settings.ServerKey != "" {
		d.Set("server_key", settings.ServerKey)
	}

	d.Set("public_address", settings.PublicAddress)
	d.Set("public_address6", settings.PublicAddress6)
	d.Set("routed_subnet6", settings.RoutedSubnet6)
	d.Set("routed_subnet6_wg", settings.RoutedSubnet6Wg)
	d.Set("ipv6", settings.IPv6)

	d.Set("drop_permissions", settings.DropPermissions)
	d.Set("restrict_client", settings.RestrictClient)

	d.Set("sso", settings.SSO)

	sso := settings.SSO
	if sso != "" {
		ssoSettings := map[string]interface{}{
			"default_organization_id": settings.SSOOrg,
			"cache":                   settings.SSOCache,
			"client_cache":            settings.SSOClientCache,
			"server_sso_url":          settings.ServerSSOUrl,
		}

		if strings.Contains(sso, "okta") {
			ssoSettings["okta"] = []map[string]interface{}{
				{
					"app_id": settings.SSOOktaAppId,
					"mode":   settings.SSOOktaMode,
					"token":  settings.SSOOktaToken,
				},
			}
		}

		if strings.Contains(sso, "onelogin") {
			ssoSettings["onelogin"] = []map[string]interface{}{
				{
					"client_id":     settings.SSOOneloginId,
					"client_secret": settings.SSOOneloginSecret,
					"app_id":        settings.SSOOneloginAppId,
					"mode":          settings.SSOOneloginMode,
				},
			}
		}

		if strings.Contains(sso, "saml") && !strings.Contains(sso, "okta") && !strings.Contains(sso, "onelogin") || strings.Contains(sso, "okta") || strings.Contains(sso, "onelogin") {
			if settings.SSOSamlUrl != "" || settings.SSOSamlIssuerUrl != "" || settings.SSOSamlCert != "" {
				ssoSettings["saml"] = []map[string]interface{}{
					{
						"url":        settings.SSOSamlUrl,
						"issuer_url": settings.SSOSamlIssuerUrl,
						"cert":       settings.SSOSamlCert,
					},
				}
			}
		}

		if strings.Contains(sso, "authzero") {
			ssoSettings["authzero"] = []map[string]interface{}{
				{
					"subdomain":     settings.SSOAuthzeroDomain,
					"client_id":     settings.SSOAuthzeroAppId,
					"client_secret": settings.SSOAuthzeroAppSecret,
				},
			}
		}

		if strings.Contains(sso, "slack") {
			domain := ""
			if len(settings.SSOMatch) > 0 {
				domain = settings.SSOMatch[0]
			}
			ssoSettings["slack"] = []map[string]interface{}{
				{
					"domain": domain,
				},
			}
		}

		if strings.Contains(sso, "google") {
			domain := strings.Join(settings.SSOMatch, ",")
			ssoSettings["google"] = []map[string]interface{}{
				{
					"domain":      domain,
					"email":       settings.SSOGoogleEmail,
					"private_key": settings.SSOGoogleKey,
				},
			}
		}

		if strings.Contains(sso, "azure") {
			ssoSettings["azure"] = []map[string]interface{}{
				{
					"app_id":       settings.SSOAzureAppId,
					"app_secret":   settings.SSOAzureAppSecret,
					"directory_id": settings.SSOAzureDirectoryId,
					"region":       settings.SSOAzureRegion,
					"version":      settings.SSOAzureVersion,
				},
			}
		}

		if strings.Contains(sso, "radius") {
			ssoSettings["radius"] = []map[string]interface{}{
				{
					"host":   settings.SSORadiusHost,
					"secret": settings.SSORadiusSecret,
				},
			}
		}

		if strings.Contains(sso, "jumpcloud") {
			ssoSettings["jumpcloud"] = []map[string]interface{}{
				{
					"app_id": settings.SSOJumpcloudAppId,
					"secret": settings.SSOJumpcloudSecret,
				},
			}
		}

		if strings.Contains(sso, "duo") {
			ssoSettings["duo"] = []map[string]interface{}{
				{
					"token":  settings.SSODuoToken,
					"secret": settings.SSODuoSecret,
					"host":   settings.SSODuoHost,
					"mode":   settings.SSODuoMode,
				},
			}
		}

		if strings.Contains(sso, "yubico") {
			ssoSettings["yubico"] = []map[string]interface{}{
				{
					"client": settings.SSOYubicoClient,
					"secret": settings.SSOYubicoSecret,
				},
			}
		}

		d.Set("sso_settings", []map[string]interface{}{ssoSettings})
	}

	awsSettings := map[string]interface{}{
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
		"ap_southeast_3_access_key": settings.AwsApSouthEast3AccessKey,
		"ap_southeast_3_secret_key": settings.AwsApSouthEast3SecretKey,
		"ap_east_1_access_key":      settings.AwsApEast1AccessKey,
		"ap_east_1_secret_key":      settings.AwsApEast1SecretKey,
		"ap_south_1_access_key":     settings.AwsApSouth1AccessKey,
		"ap_south_1_secret_key":     settings.AwsApSouth1SecretKey,
		"sa_east_1_access_key":      settings.AwsSaEast1AccessKey,
		"sa_east_1_secret_key":      settings.AwsSaEast1SecretKey,
	}
	d.Set("cloud_provider_aws_settings", []map[string]interface{}{awsSettings})

	d.Set("cloud_provider_oracle_settings", []map[string]interface{}{
		{
			"oracle_user_ocid":  settings.OracleUserOcid,
			"oracle_public_key": settings.OraclePublicKey,
		},
	})

	d.Set("cloud_provider_pritunl_settings", []map[string]interface{}{
		{
			"host":   settings.PritunlCloudHost,
			"token":  settings.PritunlCloudToken,
			"secret": settings.PritunlCloudSecret,
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

	if v, ok := d.GetOk("email_from"); ok {
		settings.EmailFrom = v.(string)
	}

	if v, ok := d.GetOk("email_server"); ok {
		settings.EmailServer = v.(string)
	}

	if v, ok := d.GetOk("email_username"); ok {
		settings.EmailUsername = v.(string)
	}

	if v, ok := d.GetOk("email_password"); ok {
		settings.EmailPassword = v.(string)
	}

	if v, ok := d.GetOk("influxdb_url"); ok {
		settings.InfluxdbUrl = v.(string)
	}

	if v, ok := d.GetOk("influxdb_token"); ok {
		settings.InfluxdbToken = v.(string)
	}

	if v, ok := d.GetOk("influxdb_org"); ok {
		settings.InfluxdbOrg = v.(string)
	}

	if v, ok := d.GetOk("influxdb_bucket"); ok {
		settings.InfluxdbBucket = v.(string)
	}

	if v, ok := d.GetOk("server_cert"); ok {
		settings.ServerCert = v.(string)
	}

	if v, ok := d.GetOk("server_key"); ok {
		settings.ServerKey = v.(string)
	}

	if v, ok := d.GetOk("public_address"); ok {
		settings.PublicAddress = v.(string)
	}

	if v, ok := d.GetOk("public_address6"); ok {
		settings.PublicAddress6 = v.(string)
	}

	if v, ok := d.GetOk("routed_subnet6"); ok {
		settings.RoutedSubnet6 = v.(string)
	}

	if v, ok := d.GetOk("routed_subnet6_wg"); ok {
		settings.RoutedSubnet6Wg = v.(string)
	}

	if v, ok := d.GetOk("ipv6"); ok {
		settings.IPv6 = v.(bool)
	}

	if v, ok := d.GetOk("drop_permissions"); ok {
		settings.DropPermissions = v.(bool)
	}

	if v, ok := d.GetOk("restrict_client"); ok {
		settings.RestrictClient = v.(bool)
	}

	if v, ok := d.GetOk("restrict_import"); ok {
		settings.RestrictImport = v.(bool)
	}

	if v, ok := d.GetOk("client_reconnect"); ok {
		settings.ClientReconnect = v.(bool)
	}

	if v, ok := d.GetOk("sso"); ok {
		settings.SSO = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider"); ok {
		settings.CloudProvider = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.route53_region"); ok {
		settings.Route53Region = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.route53_zone"); ok {
		settings.Route53Zone = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.us_east_1_access_key"); ok {
		settings.AwsUsEast1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.us_east_1_secret_key"); ok {
		settings.AwsUsEast1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.us_east_2_access_key"); ok {
		settings.AwsUsEast2AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.us_east_2_secret_key"); ok {
		settings.AwsUsEast2SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.us_west_1_access_key"); ok {
		settings.AwsUsWest1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.us_west_1_secret_key"); ok {
		settings.AwsUsWest1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.us_west_2_access_key"); ok {
		settings.AwsUsWest2AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.us_west_2_secret_key"); ok {
		settings.AwsUsWest2SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.us_gov_east_1_access_key"); ok {
		settings.AwsUsGovEast1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.us_gov_east_1_secret_key"); ok {
		settings.AwsUsGovEast1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.us_gov_west_1_access_key"); ok {
		settings.AwsUsGovWest1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.us_gov_west_1_secret_key"); ok {
		settings.AwsUsGovWest1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.eu_north_1_access_key"); ok {
		settings.AwsEuNorth1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.eu_north_1_secret_key"); ok {
		settings.AwsEuNorth1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.eu_west_1_access_key"); ok {
		settings.AwsEuWest1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.eu_west_1_secret_key"); ok {
		settings.AwsEuWest1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.eu_west_2_access_key"); ok {
		settings.AwsEuWest2AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.eu_west_2_secret_key"); ok {
		settings.AwsEuWest2SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.eu_west_3_access_key"); ok {
		settings.AwsEuWest3AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.eu_west_3_secret_key"); ok {
		settings.AwsEuWest3SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.eu_central_1_access_key"); ok {
		settings.AwsEuCentral1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.eu_central_1_secret_key"); ok {
		settings.AwsEuCentral1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.ca_central_1_access_key"); ok {
		settings.AwsCaCentral1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.ca_central_1_secret_key"); ok {
		settings.AwsCaCentral1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.cn_north_1_access_key"); ok {
		settings.AwsCnNorth1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.cn_north_1_secret_key"); ok {
		settings.AwsCnNorth1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.cn_northwest_1_access_key"); ok {
		settings.AwsCnNorthWest1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.cn_northwest_1_secret_key"); ok {
		settings.AwsCnNorthWest1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.ap_northeast_1_access_key"); ok {
		settings.AwsApNorthEast1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.ap_northeast_1_secret_key"); ok {
		settings.AwsApNorthEast1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.ap_northeast_2_access_key"); ok {
		settings.AwsApNorthEast2AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.ap_northeast_2_secret_key"); ok {
		settings.AwsApNorthEast2SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.ap_southeast_1_access_key"); ok {
		settings.AwsApSouthEast1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.ap_southeast_1_secret_key"); ok {
		settings.AwsApSouthEast1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.ap_southeast_2_access_key"); ok {
		settings.AwsApSouthEast2AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.ap_southeast_2_secret_key"); ok {
		settings.AwsApSouthEast2SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.ap_southeast_3_access_key"); ok {
		settings.AwsApSouthEast3AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.ap_southeast_3_secret_key"); ok {
		settings.AwsApSouthEast3SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.ap_east_1_access_key"); ok {
		settings.AwsApEast1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.ap_east_1_secret_key"); ok {
		settings.AwsApEast1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.ap_south_1_access_key"); ok {
		settings.AwsApSouth1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.ap_south_1_secret_key"); ok {
		settings.AwsApSouth1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.sa_east_1_access_key"); ok {
		settings.AwsSaEast1AccessKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_aws_settings.0.sa_east_1_secret_key"); ok {
		settings.AwsSaEast1SecretKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_oracle_settings.0.oracle_user_ocid"); ok {
		settings.OracleUserOcid = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_oracle_settings.0.oracle_public_key"); ok {
		settings.OraclePublicKey = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_pritunl_settings.0.host"); ok {
		settings.PritunlCloudHost = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_pritunl_settings.0.token"); ok {
		settings.PritunlCloudToken = v.(string)
	}

	if v, ok := d.GetOk("cloud_provider_pritunl_settings.0.secret"); ok {
		settings.PritunlCloudSecret = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.default_organization_id"); ok {
		settings.SSOOrg = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.cache"); ok {
		settings.SSOCache = v.(bool)
	}

	if v, ok := d.GetOk("sso_settings.0.client_cache"); ok {
		settings.SSOClientCache = v.(bool)
	}

	if v, ok := d.GetOk("sso_settings.0.server_sso_url"); ok {
		settings.ServerSSOUrl = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.saml.0.url"); ok {
		settings.SSOSamlUrl = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.saml.0.issuer_url"); ok {
		settings.SSOSamlIssuerUrl = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.saml.0.cert"); ok {
		settings.SSOSamlCert = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.duo.0.token"); ok {
		settings.SSODuoToken = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.duo.0.secret"); ok {
		settings.SSODuoSecret = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.duo.0.host"); ok {
		settings.SSODuoHost = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.duo.0.mode"); ok {
		settings.SSODuoMode = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.yubico.0.client"); ok {
		settings.SSOYubicoClient = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.yubico.0.secret"); ok {
		settings.SSOYubicoSecret = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.google.0.domain"); ok {
		settings.SSOMatch = strings.Split(v.(string), ",")
	}

	if v, ok := d.GetOk("sso_settings.0.google.0.email"); ok {
		settings.SSOGoogleEmail = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.google.0.private_key"); ok {
		settings.SSOGoogleKey = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.okta.0.app_id"); ok {
		settings.SSOOktaAppId = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.okta.0.mode"); ok {
		settings.SSOOktaMode = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.okta.0.token"); ok {
		settings.SSOOktaToken = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.onelogin.0.client_id"); ok {
		settings.SSOOneloginId = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.onelogin.0.client_secret"); ok {
		settings.SSOOneloginSecret = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.onelogin.0.app_id"); ok {
		settings.SSOOneloginAppId = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.onelogin.0.mode"); ok {
		settings.SSOOneloginMode = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.authzero.0.subdomain"); ok {
		settings.SSOAuthzeroDomain = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.authzero.0.client_id"); ok {
		settings.SSOAuthzeroAppId = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.authzero.0.client_secret"); ok {
		settings.SSOAuthzeroAppSecret = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.slack.0.domain"); ok {
		settings.SSOMatch = []string{v.(string)}
	}

	if v, ok := d.GetOk("sso_settings.0.azure.0.directory_id"); ok {
		settings.SSOAzureDirectoryId = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.azure.0.app_id"); ok {
		settings.SSOAzureAppId = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.azure.0.app_secret"); ok {
		settings.SSOAzureAppSecret = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.azure.0.region"); ok {
		settings.SSOAzureRegion = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.azure.0.version"); ok {
		settings.SSOAzureVersion = v.(int)
	}

	if v, ok := d.GetOk("sso_settings.0.radius.0.host"); ok {
		settings.SSORadiusHost = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.radius.0.secret"); ok {
		settings.SSORadiusSecret = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.jumpcloud.0.app_id"); ok {
		settings.SSOJumpcloudAppId = v.(string)
	}

	if v, ok := d.GetOk("sso_settings.0.jumpcloud.0.secret"); ok {
		settings.SSOJumpcloudSecret = v.(string)
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

	if d.HasChange("email_from") {
		settings.EmailFrom = d.Get("email_from").(string)
	}

	if d.HasChange("email_server") {
		settings.EmailServer = d.Get("email_server").(string)
	}

	if d.HasChange("email_username") {
		settings.EmailUsername = d.Get("email_username").(string)
	}

	if d.HasChange("email_password") {
		settings.EmailPassword = d.Get("email_password").(string)
	}

	if d.HasChange("influxdb_url") {
		settings.InfluxdbUrl = d.Get("influxdb_url").(string)
	}

	if d.HasChange("influxdb_token") {
		settings.InfluxdbToken = d.Get("influxdb_token").(string)
	}

	if d.HasChange("influxdb_org") {
		settings.InfluxdbOrg = d.Get("influxdb_org").(string)
	}

	if d.HasChange("influxdb_bucket") {
		settings.InfluxdbBucket = d.Get("influxdb_bucket").(string)
	}

	if d.HasChange("server_cert") {
		settings.ServerCert = d.Get("server_cert").(string)
	}

	if d.HasChange("server_key") {
		settings.ServerKey = d.Get("server_key").(string)
	}

	if d.HasChange("public_address") {
		settings.PublicAddress = d.Get("public_address").(string)
	}

	if d.HasChange("public_address6") {
		settings.PublicAddress6 = d.Get("public_address6").(string)
	}

	if d.HasChange("routed_subnet6") {
		settings.RoutedSubnet6 = d.Get("routed_subnet6").(string)
	}

	if d.HasChange("routed_subnet6_wg") {
		settings.RoutedSubnet6Wg = d.Get("routed_subnet6_wg").(string)
	}

	if d.HasChange("ipv6") {
		settings.IPv6 = d.Get("ipv6").(bool)
	}

	if d.HasChange("drop_permissions") {
		settings.DropPermissions = d.Get("drop_permissions").(bool)
	}

	if d.HasChange("restrict_client") {
		settings.RestrictClient = d.Get("restrict_client").(bool)
	}

	if d.HasChange("restrict_import") {
		settings.RestrictImport = d.Get("restrict_import").(bool)
	}

	if d.HasChange("client_reconnect") {
		settings.ClientReconnect = d.Get("client_reconnect").(bool)
	}

	if d.HasChange("sso") {
		settings.SSO = d.Get("sso").(string)
	}

	if d.HasChange("cloud_provider") {
		settings.CloudProvider = d.Get("cloud_provider").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.route53_region") {
		settings.Route53Region = d.Get("cloud_provider_aws_settings.0.route53_region").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.route53_zone") {
		settings.Route53Zone = d.Get("cloud_provider_aws_settings.0.route53_zone").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.us_east_1_access_key") {
		settings.AwsUsEast1AccessKey = d.Get("cloud_provider_aws_settings.0.us_east_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.us_east_1_secret_key") {
		settings.AwsUsEast1SecretKey = d.Get("cloud_provider_aws_settings.0.us_east_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.us_east_2_access_key") {
		settings.AwsUsEast2AccessKey = d.Get("cloud_provider_aws_settings.0.us_east_2_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.us_east_2_secret_key") {
		settings.AwsUsEast2SecretKey = d.Get("cloud_provider_aws_settings.0.us_east_2_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.us_west_1_access_key") {
		settings.AwsUsWest1AccessKey = d.Get("cloud_provider_aws_settings.0.us_west_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.us_west_1_secret_key") {
		settings.AwsUsWest1SecretKey = d.Get("cloud_provider_aws_settings.0.us_west_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.us_west_2_access_key") {
		settings.AwsUsWest2AccessKey = d.Get("cloud_provider_aws_settings.0.us_west_2_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.us_west_2_secret_key") {
		settings.AwsUsWest2SecretKey = d.Get("cloud_provider_aws_settings.0.us_west_2_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.us_gov_east_1_access_key") {
		settings.AwsUsGovEast1AccessKey = d.Get("cloud_provider_aws_settings.0.us_gov_east_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.us_gov_east_1_secret_key") {
		settings.AwsUsGovEast1SecretKey = d.Get("cloud_provider_aws_settings.0.us_gov_east_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.us_gov_west_1_access_key") {
		settings.AwsUsGovWest1AccessKey = d.Get("cloud_provider_aws_settings.0.us_gov_west_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.us_gov_west_1_secret_key") {
		settings.AwsUsGovWest1SecretKey = d.Get("cloud_provider_aws_settings.0.us_gov_west_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.eu_north_1_access_key") {
		settings.AwsEuNorth1AccessKey = d.Get("cloud_provider_aws_settings.0.eu_north_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.eu_north_1_secret_key") {
		settings.AwsEuNorth1SecretKey = d.Get("cloud_provider_aws_settings.0.eu_north_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.eu_west_1_access_key") {
		settings.AwsEuWest1AccessKey = d.Get("cloud_provider_aws_settings.0.eu_west_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.eu_west_1_secret_key") {
		settings.AwsEuWest1SecretKey = d.Get("cloud_provider_aws_settings.0.eu_west_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.eu_west_2_access_key") {
		settings.AwsEuWest2AccessKey = d.Get("cloud_provider_aws_settings.0.eu_west_2_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.eu_west_2_secret_key") {
		settings.AwsEuWest2SecretKey = d.Get("cloud_provider_aws_settings.0.eu_west_2_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.eu_west_3_access_key") {
		settings.AwsEuWest3AccessKey = d.Get("cloud_provider_aws_settings.0.eu_west_3_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.eu_west_3_secret_key") {
		settings.AwsEuWest3SecretKey = d.Get("cloud_provider_aws_settings.0.eu_west_3_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.eu_central_1_access_key") {
		settings.AwsEuCentral1AccessKey = d.Get("cloud_provider_aws_settings.0.eu_central_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.eu_central_1_secret_key") {
		settings.AwsEuCentral1SecretKey = d.Get("cloud_provider_aws_settings.0.eu_central_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.ca_central_1_access_key") {
		settings.AwsCaCentral1AccessKey = d.Get("cloud_provider_aws_settings.0.ca_central_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.ca_central_1_secret_key") {
		settings.AwsCaCentral1SecretKey = d.Get("cloud_provider_aws_settings.0.ca_central_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.cn_north_1_access_key") {
		settings.AwsCnNorth1AccessKey = d.Get("cloud_provider_aws_settings.0.cn_north_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.cn_north_1_secret_key") {
		settings.AwsCnNorth1SecretKey = d.Get("cloud_provider_aws_settings.0.cn_north_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.cn_northwest_1_access_key") {
		settings.AwsCnNorthWest1AccessKey = d.Get("cloud_provider_aws_settings.0.cn_northwest_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.cn_northwest_1_secret_key") {
		settings.AwsCnNorthWest1SecretKey = d.Get("cloud_provider_aws_settings.0.cn_northwest_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.ap_northeast_1_access_key") {
		settings.AwsApNorthEast1AccessKey = d.Get("cloud_provider_aws_settings.0.ap_northeast_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.ap_northeast_1_secret_key") {
		settings.AwsApNorthEast1SecretKey = d.Get("cloud_provider_aws_settings.0.ap_northeast_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.ap_northeast_2_access_key") {
		settings.AwsApNorthEast2AccessKey = d.Get("cloud_provider_aws_settings.0.ap_northeast_2_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.ap_northeast_2_secret_key") {
		settings.AwsApNorthEast2SecretKey = d.Get("cloud_provider_aws_settings.0.ap_northeast_2_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.ap_southeast_1_access_key") {
		settings.AwsApSouthEast1AccessKey = d.Get("cloud_provider_aws_settings.0.ap_southeast_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.ap_southeast_1_secret_key") {
		settings.AwsApSouthEast1SecretKey = d.Get("cloud_provider_aws_settings.0.ap_southeast_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.ap_southeast_2_access_key") {
		settings.AwsApSouthEast2AccessKey = d.Get("cloud_provider_aws_settings.0.ap_southeast_2_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.ap_southeast_2_secret_key") {
		settings.AwsApSouthEast2SecretKey = d.Get("cloud_provider_aws_settings.0.ap_southeast_2_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.ap_southeast_3_access_key") {
		settings.AwsApSouthEast3AccessKey = d.Get("cloud_provider_aws_settings.0.ap_southeast_3_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.ap_southeast_3_secret_key") {
		settings.AwsApSouthEast3SecretKey = d.Get("cloud_provider_aws_settings.0.ap_southeast_3_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.ap_east_1_access_key") {
		settings.AwsApEast1AccessKey = d.Get("cloud_provider_aws_settings.0.ap_east_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.ap_east_1_secret_key") {
		settings.AwsApEast1SecretKey = d.Get("cloud_provider_aws_settings.0.ap_east_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.ap_south_1_access_key") {
		settings.AwsApSouth1AccessKey = d.Get("cloud_provider_aws_settings.0.ap_south_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.ap_south_1_secret_key") {
		settings.AwsApSouth1SecretKey = d.Get("cloud_provider_aws_settings.0.ap_south_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.sa_east_1_access_key") {
		settings.AwsSaEast1AccessKey = d.Get("cloud_provider_aws_settings.0.sa_east_1_access_key").(string)
	}

	if d.HasChange("cloud_provider_aws_settings.0.sa_east_1_secret_key") {
		settings.AwsSaEast1SecretKey = d.Get("cloud_provider_aws_settings.0.sa_east_1_secret_key").(string)
	}

	if d.HasChange("cloud_provider_oracle_settings.0.oracle_user_ocid") {
		settings.OracleUserOcid = d.Get("cloud_provider_oracle_settings.0.oracle_user_ocid").(string)
	}

	if d.HasChange("cloud_provider_oracle_settings.0.oracle_public_key") {
		settings.OraclePublicKey = d.Get("cloud_provider_oracle_settings.0.oracle_public_key").(string)
	}

	if d.HasChange("cloud_provider_pritunl_settings.0.host") {
		settings.PritunlCloudHost = d.Get("cloud_provider_pritunl_settings.0.host").(string)
	}

	if d.HasChange("cloud_provider_pritunl_settings.0.token") {
		settings.PritunlCloudToken = d.Get("cloud_provider_pritunl_settings.0.token").(string)
	}

	if d.HasChange("cloud_provider_pritunl_settings.0.secret") {
		settings.PritunlCloudSecret = d.Get("cloud_provider_pritunl_settings.0.secret").(string)
	}

	if d.HasChange("sso_settings.0.default_organization_id") {
		settings.SSOOrg = d.Get("sso_settings.0.default_organization_id").(string)
	}

	if d.HasChange("sso_settings.0.cache") {
		settings.SSOCache = d.Get("sso_settings.0.cache").(bool)
	}

	if d.HasChange("sso_settings.0.client_cache") {
		settings.SSOClientCache = d.Get("sso_settings.0.client_cache").(bool)
	}

	if d.HasChange("sso_settings.0.server_sso_url") {
		settings.ServerSSOUrl = d.Get("sso_settings.0.server_sso_url").(string)
	}

	if d.HasChange("sso_settings.0.saml.0.url") {
		settings.SSOSamlUrl = d.Get("sso_settings.0.saml.0.url").(string)
	}

	if d.HasChange("sso_settings.0.saml.0.issuer_url") {
		settings.SSOSamlIssuerUrl = d.Get("sso_settings.0.saml.0.issuer_url").(string)
	}

	if d.HasChange("sso_settings.0.saml.0.cert") {
		settings.SSOSamlCert = d.Get("sso_settings.0.saml.0.cert").(string)
	}

	if d.HasChange("sso_settings.0.duo.0.token") {
		settings.SSODuoToken = d.Get("sso_settings.0.duo.0.token").(string)
	}

	if d.HasChange("sso_settings.0.duo.0.secret") {
		settings.SSODuoSecret = d.Get("sso_settings.0.duo.0.secret").(string)
	}

	if d.HasChange("sso_settings.0.duo.0.host") {
		settings.SSODuoHost = d.Get("sso_settings.0.duo.0.host").(string)
	}

	if d.HasChange("sso_settings.0.duo.0.mode") {
		settings.SSODuoMode = d.Get("sso_settings.0.duo.0.mode").(string)
	}

	if d.HasChange("sso_settings.0.yubico.0.client") {
		settings.SSOYubicoClient = d.Get("sso_settings.0.yubico.0.client").(string)
	}

	if d.HasChange("sso_settings.0.yubico.0.secret") {
		settings.SSOYubicoSecret = d.Get("sso_settings.0.yubico.0.secret").(string)
	}

	if d.HasChange("sso_settings.0.google.0.domain") {
		settings.SSOMatch = strings.Split(d.Get("sso_settings.0.google.0.domain").(string), ",")
	}

	if d.HasChange("sso_settings.0.google.0.email") {
		settings.SSOGoogleEmail = d.Get("sso_settings.0.google.0.email").(string)
	}

	if d.HasChange("sso_settings.0.google.0.private_key") {
		settings.SSOGoogleKey = d.Get("sso_settings.0.google.0.private_key").(string)
	}

	if d.HasChange("sso_settings.0.okta.0.app_id") {
		settings.SSOOktaAppId = d.Get("sso_settings.0.okta.0.app_id").(string)
	}

	if d.HasChange("sso_settings.0.okta.0.mode") {
		settings.SSOOktaMode = d.Get("sso_settings.0.okta.0.mode").(string)
	}

	if d.HasChange("sso_settings.0.okta.0.token") {
		settings.SSOOktaToken = d.Get("sso_settings.0.okta.0.token").(string)
	}

	if d.HasChange("sso_settings.0.onelogin.0.client_id") {
		settings.SSOOneloginId = d.Get("sso_settings.0.onelogin.0.client_id").(string)
	}

	if d.HasChange("sso_settings.0.onelogin.0.client_secret") {
		settings.SSOOneloginSecret = d.Get("sso_settings.0.onelogin.0.client_secret").(string)
	}

	if d.HasChange("sso_settings.0.onelogin.0.app_id") {
		settings.SSOOneloginAppId = d.Get("sso_settings.0.onelogin.0.app_id").(string)
	}

	if d.HasChange("sso_settings.0.onelogin.0.mode") {
		settings.SSOOneloginMode = d.Get("sso_settings.0.onelogin.0.mode").(string)
	}

	if d.HasChange("sso_settings.0.authzero.0.subdomain") {
		settings.SSOAuthzeroDomain = d.Get("sso_settings.0.authzero.0.subdomain").(string)
	}

	if d.HasChange("sso_settings.0.authzero.0.client_id") {
		settings.SSOAuthzeroAppId = d.Get("sso_settings.0.authzero.0.client_id").(string)
	}

	if d.HasChange("sso_settings.0.authzero.0.client_secret") {
		settings.SSOAuthzeroAppSecret = d.Get("sso_settings.0.authzero.0.client_secret").(string)
	}

	if d.HasChange("sso_settings.0.slack.0.domain") {
		settings.SSOMatch = []string{d.Get("sso_settings.0.slack.0.domain").(string)}
	}

	if d.HasChange("sso_settings.0.azure.0.directory_id") {
		settings.SSOAzureDirectoryId = d.Get("sso_settings.0.azure.0.directory_id").(string)
	}

	if d.HasChange("sso_settings.0.azure.0.app_id") {
		settings.SSOAzureAppId = d.Get("sso_settings.0.azure.0.app_id").(string)
	}

	if d.HasChange("sso_settings.0.azure.0.app_secret") {
		settings.SSOAzureAppSecret = d.Get("sso_settings.0.azure.0.app_secret").(string)
	}

	if d.HasChange("sso_settings.0.azure.0.region") {
		settings.SSOAzureRegion = d.Get("sso_settings.0.azure.0.region").(string)
	}

	if d.HasChange("sso_settings.0.azure.0.version") {
		settings.SSOAzureVersion = d.Get("sso_settings.0.azure.0.version").(int)
	}

	if d.HasChange("sso_settings.0.radius.0.host") {
		settings.SSORadiusHost = d.Get("sso_settings.0.radius.0.host").(string)
	}

	if d.HasChange("sso_settings.0.radius.0.secret") {
		settings.SSORadiusSecret = d.Get("sso_settings.0.radius.0.secret").(string)
	}

	if d.HasChange("sso_settings.0.jumpcloud.0.app_id") {
		settings.SSOJumpcloudAppId = d.Get("sso_settings.0.jumpcloud.0.app_id").(string)
	}

	if d.HasChange("sso_settings.0.jumpcloud.0.secret") {
		settings.SSOJumpcloudSecret = d.Get("sso_settings.0.jumpcloud.0.secret").(string)
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

var cloudProviderPritunlSchema = map[string]*schema.Schema{
	"host": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Pritunl Cloud host",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"token": {
		Type:        schema.TypeString,
		Required:    true,
		Sensitive:   true,
		Description: "Pritunl Cloud token",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"secret": {
		Type:        schema.TypeString,
		Required:    true,
		Sensitive:   true,
		Description: "Pritunl Cloud secret",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
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
			"ap-southeast-3",
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
	"ap_southeast_3_access_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Asia Pacific (Jakarta) Access Key or 'role' to use the instance IAM role",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"ap_southeast_3_secret_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Asia Pacific (Jakarta) Secret key or 'role' to use the instance IAM role",
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

var ssoSettingsSchema = map[string]*schema.Schema{
	"default_organization_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Default Single Sign-On Organization",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"cache": {
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Enable an 8 hour secondary authentication cache using client ID, IP address and MAC address. This will allow clients to reconnect without secondary authentication. Works with Duo push, Okta push, OneLogin push, Duo passcodes and YubiKeys. Supported by all OpenVPN clients",
	},
	"client_cache": {
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Enable a two day secondary authentication cache using a token stored on the client. This will allow clients to reconnect without secondary authentication. Works with Duo push, Okta push, OneLogin push, Duo passcodes and YubiKeys. Only supported by Pritunl client",
	},
	"server_sso_url": {
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "Server SSO URL",
	},
	"okta": {
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: []string{"sso_settings.0.onelogin", "sso_settings.0.authzero", "sso_settings.0.slack", "sso_settings.0.google", "sso_settings.0.azure", "sso_settings.0.radius", "sso_settings.0.jumpcloud"},
		RequiredWith:  []string{"sso_settings.0.saml"},
		Elem: &schema.Resource{
			Schema: oktaSsoSettingsSchema,
		},
	},
	"onelogin": {
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: []string{"sso_settings.0.okta", "sso_settings.0.authzero", "sso_settings.0.slack", "sso_settings.0.google", "sso_settings.0.azure", "sso_settings.0.radius", "sso_settings.0.jumpcloud"},
		RequiredWith:  []string{"sso_settings.0.saml"},
		Elem: &schema.Resource{
			Schema: oneloginSsoSettingsSchema,
		},
	},
	"authzero": {
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: []string{"sso_settings.0.okta", "sso_settings.0.onelogin", "sso_settings.0.slack", "sso_settings.0.google", "sso_settings.0.azure", "sso_settings.0.saml", "sso_settings.0.radius", "sso_settings.0.jumpcloud"},
		Elem: &schema.Resource{
			Schema: authzeroSsoSettingsSchema,
		},
	},
	"slack": {
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: []string{"sso_settings.0.okta", "sso_settings.0.onelogin", "sso_settings.0.authzero", "sso_settings.0.google", "sso_settings.0.azure", "sso_settings.0.saml", "sso_settings.0.radius", "sso_settings.0.jumpcloud"},
		Elem: &schema.Resource{
			Schema: slackSsoSettingsSchema,
		},
	},
	"google": {
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: []string{"sso_settings.0.okta", "sso_settings.0.onelogin", "sso_settings.0.slack", "sso_settings.0.authzero", "sso_settings.0.azure", "sso_settings.0.saml", "sso_settings.0.radius", "sso_settings.0.jumpcloud"},
		Elem: &schema.Resource{
			Schema: googleSsoSettingsSchema,
		},
	},
	"azure": {
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: []string{"sso_settings.0.okta", "sso_settings.0.onelogin", "sso_settings.0.authzero", "sso_settings.0.slack", "sso_settings.0.google", "sso_settings.0.saml", "sso_settings.0.radius", "sso_settings.0.jumpcloud"},
		Elem: &schema.Resource{
			Schema: azureSsoSettingsSchema,
		},
	},
	"saml": {
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: []string{"sso_settings.0.authzero", "sso_settings.0.slack", "sso_settings.0.google", "sso_settings.0.azure", "sso_settings.0.radius", "sso_settings.0.jumpcloud"},
		Elem: &schema.Resource{
			Schema: samlSsoSettingsSchema,
		},
	},
	"radius": {
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: []string{"sso_settings.0.okta", "sso_settings.0.onelogin", "sso_settings.0.authzero", "sso_settings.0.slack", "sso_settings.0.google", "sso_settings.0.azure", "sso_settings.0.saml", "sso_settings.0.jumpcloud"},
		Elem: &schema.Resource{
			Schema: radiusSsoSettingsSchema,
		},
	},
	"jumpcloud": {
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: []string{"sso_settings.0.okta", "sso_settings.0.onelogin", "sso_settings.0.authzero", "sso_settings.0.slack", "sso_settings.0.google", "sso_settings.0.azure", "sso_settings.0.saml", "sso_settings.0.radius"},
		Elem: &schema.Resource{
			Schema: jumpcloudSsoSettingsSchema,
		},
	},
	"duo": {
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: []string{"sso_settings.0.yubico"},
		Elem: &schema.Resource{
			Schema: duoSsoSettingsSchema,
		},
	},
	"yubico": {
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: []string{"sso_settings.0.duo", "sso_settings.0.radius"},
		Elem: &schema.Resource{
			Schema: yubicoSsoSettingsSchema,
		},
	},
}

var oktaSsoSettingsSchema = map[string]*schema.Schema{
	"token": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Okta API token",
		Sensitive:   true,
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"app_id": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The ID on Okta Pritunl app. This can be found in the URL of the app settings page. Required to verify user is attached to Okta application on each VPN connection.",
	},
	"mode": {
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "Secondary factor for Okta users. Push when available will skip authentication for users who do not have push configured.",
		ValidateFunc: validation.StringInSlice([]string{"passcode", "push", "push_none"}, false),
	},
}

var oneloginSsoSettingsSchema = map[string]*schema.Schema{
	"client_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "OneLogin API client ID",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"client_secret": {
		Type:        schema.TypeString,
		Required:    true,
		Sensitive:   true,
		Description: "OneLogin API client secret",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"app_id": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The ID on OneLogin Pritunl app. This can be found in the URL of the app settings page. Required to verify user is attached to OneLogin application on each VPN connection.",
	},
	"mode": {
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "Secondary factor for OneLogin users. Push when available will skip authentication for users who do not have push configured.",
		ValidateFunc: validation.StringInSlice([]string{"passcode", "push", "push_none"}, false),
	},
}

var samlSsoSettingsSchema = map[string]*schema.Schema{
	"url": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The SAML identity provider single sign-on url. Also known as SAML 2.0 Endpoint",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"issuer_url": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The SAML identity provider issuer url",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"cert": {
		Type:        schema.TypeString,
		Required:    true,
		Sensitive:   true,
		Description: "The SAML X.509 Certificate",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
}

var authzeroSsoSettingsSchema = map[string]*schema.Schema{
	"subdomain": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Subdomain of Auth0 application. Enter subdomain portion only such as 'pritunl' for pritunl.auth0.com",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"client_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Auth0 application client ID",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"client_secret": {
		Type:        schema.TypeString,
		Required:    true,
		Sensitive:   true,
		Description: "Auth0 application client secret",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
}

var radiusSsoSettingsSchema = map[string]*schema.Schema{
	"host": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Radius host such as localhost:1645. If no port is specified default port 1645 will be used. Separate multiple hosts with a comma.",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"secret": {
		Type:        schema.TypeString,
		Required:    true,
		Sensitive:   true,
		Description: "Radius shared secret",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
}

var slackSsoSettingsSchema = map[string]*schema.Schema{
	"domain": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Slack team domain to match against users team. (example: pritunl.slack.com)",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
}

var azureSsoSettingsSchema = map[string]*schema.Schema{
	"app_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Enter Azure application ID",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"app_secret": {
		Type:        schema.TypeString,
		Required:    true,
		Sensitive:   true,
		Description: "Azure Application Secret",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"directory_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Azure Directory ID",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"region": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Azure region",
	},
	"version": {
		Type:         schema.TypeInt,
		Optional:     true,
		Description:  "Azure API version",
		ValidateFunc: validation.IntInSlice([]int{1, 2}),
	},
}

var googleSsoSettingsSchema = map[string]*schema.Schema{
	"domain": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Google apps domain to match against users email address. Multiple domains can be entered seperated by a comma. (example: pritunl.com)",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"email": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The email address of an administrator user in the Google G Suite to delegate API access to. This user will be used to get the groups of Google users. Only needed when providing the Google private key",
	},
	"private_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Sensitive:   true,
		Description: "The private key for service account in JSON format. This will allow a case sensitive match for any of the user groups to an existing organization. The group names will be matched to the first matching organization name in sorted order. If empty organization wont be matched to user groups. Also requires Google Admin Email",
	},
}

var jumpcloudSsoSettingsSchema = map[string]*schema.Schema{
	"app_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "JumpCloud application ID",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"secret": {
		Type:        schema.TypeString,
		Required:    true,
		Sensitive:   true,
		Description: "JumpCloud secret",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
}

var duoSsoSettingsSchema = map[string]*schema.Schema{
	"token": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Duo Integration Key",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"secret": {
		Type:        schema.TypeString,
		Required:    true,
		Sensitive:   true,
		Description: "Duo Secret Key",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"host": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Duo API Hostname",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"mode": {
		Type:         schema.TypeString,
		Required:     true,
		Description:  "Duo authentication mode",
		ValidateFunc: validation.StringInSlice([]string{"push", "phone", "push_phone", "passcode"}, false),
	},
}

var yubicoSsoSettingsSchema = map[string]*schema.Schema{
	"client": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Yubico Client ID",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
	"secret": {
		Type:        schema.TypeString,
		Required:    true,
		Sensitive:   true,
		Description: "Yubico Secret Key",
		ValidateFunc: func(i interface{}, s string) ([]string, []error) {
			return validation.StringIsNotEmpty(i, s)
		},
	},
}
