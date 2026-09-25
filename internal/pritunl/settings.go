package pritunl

import (
	"encoding/json"
	"fmt"
)

func (s *Settings) UnmarshalJSON(data []byte) error {
	// Use an alias to prevent infinite recursion.
	// Override SSO with interface{} because the API returns false (bool)
	// when SSO is disabled, but a string (e.g. "duo") when enabled.
	type Alias Settings
	aux := &struct {
		SSO interface{} `json:"sso"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	switch v := aux.SSO.(type) {
	case string:
		s.SSO = v
	case bool, nil:
		s.SSO = ""
	default:
		s.SSO = fmt.Sprintf("%v", v)
	}
	return nil
}

type Settings struct {
	Username                 string   `json:"username"`
	Auditing                 string   `json:"auditing,omitempty"`
	Monitoring               string   `json:"monitoring,omitempty"`
	InfluxdbUrl              string   `json:"influxdb_url,omitempty"`
	InfluxdbToken            string   `json:"influxdb_token,omitempty"`
	InfluxdbOrg              string   `json:"influxdb_org,omitempty"`
	InfluxdbBucket           string   `json:"influxdb_bucket,omitempty"`
	EmailFrom                string   `json:"email_from,omitempty"`
	EmailServer              string   `json:"email_server,omitempty"`
	EmailUsername            string   `json:"email_username,omitempty"`
	EmailPassword            string   `json:"email_password,omitempty"`
	PinMode                  string   `json:"pin_mode,omitempty"`
	SSO                      string   `json:"sso,omitempty"`
	SSOMatch                 []string `json:"sso_match,omitempty"`
	SSODuoToken              string   `json:"sso_duo_token,omitempty"`
	SSODuoSecret             string   `json:"sso_duo_secret,omitempty"`
	SSODuoHost               string   `json:"sso_duo_host,omitempty"`
	SSODuoMode               string   `json:"sso_duo_mode,omitempty"`
	SSOYubicoClient          string   `json:"sso_yubico_client,omitempty"`
	SSOYubicoSecret          string   `json:"sso_yubico_secret,omitempty"`
	SSOOrg                   string   `json:"sso_org,omitempty"`
	SSOAzureDirectoryId      string   `json:"sso_azure_directory_id,omitempty"`
	SSOAzureAppId            string   `json:"sso_azure_app_id,omitempty"`
	SSOAzureAppSecret        string   `json:"sso_azure_app_secret,omitempty"`
	SSOAuthzeroDomain        string   `json:"sso_authzero_domain,omitempty"`
	SSOAuthzeroAppId         string   `json:"sso_authzero_app_id,omitempty"`
	SSOAuthzeroAppSecret     string   `json:"sso_authzero_app_secret,omitempty"`
	SSOGoogleKey             string   `json:"sso_google_key,omitempty"`
	SSOGoogleEmail           string   `json:"sso_google_email,omitempty"`
	SSOSamlUrl               string   `json:"sso_saml_url,omitempty"`
	SSOSamlIssuerUrl         string   `json:"sso_saml_issuer_url,omitempty"`
	SSOSamlCert              string   `json:"sso_saml_cert,omitempty"`
	SSOOktaAppId             string   `json:"sso_okta_app_id,omitempty"`
	SSOOktaToken             string   `json:"sso_okta_token,omitempty"`
	SSOOktaMode              string   `json:"sso_okta_mode,omitempty"`
	SSOOneloginAppId         string   `json:"sso_onelogin_app_id,omitempty"`
	SSOOneloginId            string   `json:"sso_onelogin_id,omitempty"`
	SSOOneloginSecret        string   `json:"sso_onelogin_secret,omitempty"`
	SSOOneloginMode          string   `json:"sso_onelogin_mode,omitempty"`
	SSORadiusHost            string   `json:"sso_radius_host,omitempty"`
	SSORadiusSecret          string   `json:"sso_radius_secret,omitempty"`
	SSOJumpcloudAppId        string   `json:"sso_jumpcloud_app_id,omitempty"`
	SSOJumpcloudSecret       string   `json:"sso_jumpcloud_secret,omitempty"`
	SSOAzureRegion           string   `json:"sso_azure_region,omitempty"`
	SSOAzureVersion          int      `json:"sso_azure_version,omitempty"`
	ServerSSOUrl             string   `json:"server_sso_url,omitempty"`
	SSOCache                 bool     `json:"sso_cache,omitempty"`
	SSOClientCache           bool     `json:"sso_client_cache,omitempty"`
	IPv6                     bool     `json:"ipv6,omitempty"`
	DropPermissions          bool     `json:"drop_permissions,omitempty"`
	RestrictImport           bool     `json:"restrict_import,omitempty"`
	RestrictClient           bool     `json:"restrict_client,omitempty"`
	ClientReconnect          bool     `json:"client_reconnect,omitempty"`
	PublicAddress            string   `json:"public_address,omitempty"`
	PublicAddress6           string   `json:"public_address6,omitempty"`
	RoutedSubnet6            string   `json:"routed_subnet6,omitempty"`
	RoutedSubnet6Wg          string   `json:"routed_subnet6_wg,omitempty"`
	ReverseProxy             bool     `json:"reverse_proxy,omitempty"`
	Theme                    string   `json:"theme,omitempty"`
	ServerPort               int      `json:"server_port,omitempty"`
	ServerCert               string   `json:"server_cert,omitempty"`
	ServerKey                string   `json:"server_key,omitempty"`
	AcmeDomain               string   `json:"acme_domain,omitempty"`
	CloudProvider            string   `json:"cloud_provider,omitempty"`
	PritunlCloudHost         string   `json:"pritunl_cloud_host,omitempty"`
	PritunlCloudToken        string   `json:"pritunl_cloud_token,omitempty"`
	PritunlCloudSecret       string   `json:"pritunl_cloud_secret,omitempty"`
	Route53Region            string   `json:"route53_region,omitempty"`
	Route53Zone              string   `json:"route53_zone,omitempty"`
	OracleUserOcid           string   `json:"oracle_user_ocid,omitempty"`
	OraclePublicKey          string   `json:"oracle_public_key,omitempty"`
	AwsUsEast1AccessKey      string   `json:"us_east_1_access_key,omitempty"`
	AwsUsEast1SecretKey      string   `json:"us_east_1_secret_key,omitempty"`
	AwsUsEast2AccessKey      string   `json:"us_east_2_access_key,omitempty"`
	AwsUsEast2SecretKey      string   `json:"us_east_2_secret_key,omitempty"`
	AwsUsWest1AccessKey      string   `json:"us_west_1_access_key,omitempty"`
	AwsUsWest1SecretKey      string   `json:"us_west_1_secret_key,omitempty"`
	AwsUsWest2AccessKey      string   `json:"us_west_2_access_key,omitempty"`
	AwsUsWest2SecretKey      string   `json:"us_west_2_secret_key,omitempty"`
	AwsUsGovEast1AccessKey   string   `json:"us_gov_east_1_access_key,omitempty"`
	AwsUsGovEast1SecretKey   string   `json:"us_gov_east_1_secret_key,omitempty"`
	AwsUsGovWest1AccessKey   string   `json:"us_gov_west_1_access_key,omitempty"`
	AwsUsGovWest1SecretKey   string   `json:"us_gov_west_1_secret_key,omitempty"`
	AwsEuNorth1AccessKey     string   `json:"eu_north_1_access_key,omitempty"`
	AwsEuNorth1SecretKey     string   `json:"eu_north_1_secret_key,omitempty"`
	AwsEuWest1AccessKey      string   `json:"eu_west_1_access_key,omitempty"`
	AwsEuWest1SecretKey      string   `json:"eu_west_1_secret_key,omitempty"`
	AwsEuWest2AccessKey      string   `json:"eu_west_2_access_key,omitempty"`
	AwsEuWest2SecretKey      string   `json:"eu_west_2_secret_key,omitempty"`
	AwsEuWest3AccessKey      string   `json:"eu_west_3_access_key,omitempty"`
	AwsEuWest3SecretKey      string   `json:"eu_west_3_secret_key,omitempty"`
	AwsEuCentral1AccessKey   string   `json:"eu_central_1_access_key,omitempty"`
	AwsEuCentral1SecretKey   string   `json:"eu_central_1_secret_key,omitempty"`
	AwsCaCentral1AccessKey   string   `json:"ca_central_1_access_key,omitempty"`
	AwsCaCentral1SecretKey   string   `json:"ca_central_1_secret_key,omitempty"`
	AwsCnNorth1AccessKey     string   `json:"cn_north_1_access_key,omitempty"`
	AwsCnNorth1SecretKey     string   `json:"cn_north_1_secret_key,omitempty"`
	AwsCnNorthWest1AccessKey string   `json:"cn_northwest_1_access_key,omitempty"`
	AwsCnNorthWest1SecretKey string   `json:"cn_northwest_1_secret_key,omitempty"`
	AwsApNorthEast1AccessKey string   `json:"ap_northeast_1_access_key,omitempty"`
	AwsApNorthEast1SecretKey string   `json:"ap_northeast_1_secret_key,omitempty"`
	AwsApNorthEast2AccessKey string   `json:"ap_northeast_2_access_key,omitempty"`
	AwsApNorthEast2SecretKey string   `json:"ap_northeast_2_secret_key,omitempty"`
	AwsApSouthEast1AccessKey string   `json:"ap_southeast_1_access_key,omitempty"`
	AwsApSouthEast1SecretKey string   `json:"ap_southeast_1_secret_key,omitempty"`
	AwsApSouthEast2AccessKey string   `json:"ap_southeast_2_access_key,omitempty"`
	AwsApSouthEast2SecretKey string   `json:"ap_southeast_2_secret_key,omitempty"`
	AwsApSouthEast3AccessKey string   `json:"ap_southeast_3_access_key,omitempty"`
	AwsApSouthEast3SecretKey string   `json:"ap_southeast_3_secret_key,omitempty"`
	AwsApEast1AccessKey      string   `json:"ap_east_1_access_key,omitempty"`
	AwsApEast1SecretKey      string   `json:"ap_east_1_secret_key,omitempty"`
	AwsApSouth1AccessKey     string   `json:"ap_south_1_access_key,omitempty"`
	AwsApSouth1SecretKey     string   `json:"ap_south_1_secret_key,omitempty"`
	AwsSaEast1AccessKey      string   `json:"sa_east_1_access_key,omitempty"`
	AwsSaEast1SecretKey      string   `json:"sa_east_1_secret_key,omitempty"`
}
