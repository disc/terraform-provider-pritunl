package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/lfventura/terraform-provider-pritunl/internal/pritunl"
)

// A Pritunl instance exposes a single, instance wide settings object, so the
// resource is a singleton and uses a fixed identifier instead of an API
// provided one. It is also the import id of the resource.
const settingsResourceId = "settings"

const (
	// Pritunl schedules the web server restart shortly after answering the
	// request, so the endpoint is given a head start before being polled and
	// has to answer a few times in a row to be considered back
	settingsRestartDelay     = 5 * time.Second
	settingsRestartPoll      = 2 * time.Second
	settingsRestartSettle    = time.Second
	settingsRestartSuccesses = 3
	settingsRestartTimeout   = 2 * time.Minute
)

func resourceSettings() *schema.Resource {
	return &schema.Resource{
		Description: "The settings resource allows managing the global settings of a Pritunl instance: the TLS certificate served by its own web console and API, the single sign-on configuration and the instance wide user and VPN toggles. It is a singleton resource: a single instance of it maps to the whole Pritunl instance and its import id is always `settings`. Every write is a read-modify-write of the complete settings object, the same way the Pritunl web console itself works: `PUT /settings` is a full replace and not a partial update, so the settings are read back from the instance immediately before each write and handed over again with the managed attributes overlaid on top of them. That is what keeps unmanaged settings such as the SMTP or monitoring configuration untouched, and only the managed attributes are ever stored in the Terraform state. Leaving an attribute out of the configuration never resets the setting behind it: the value the instance already holds is read back and handed over again, so a setting only ever changes by being configured.",
		Schema: map[string]*schema.Schema{
			// the certificate fields are computed so that leaving them out of
			// the configuration keeps the certificate of the instance as it is
			// instead of asking to remove it, it is only handed back to Pritunl
			// when the resource itself is destroyed
			"server_cert": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				RequiredWith: []string{"server_key"},
				StateFunc:    trimSpaceStateFunc,
				Description:  "The certificate served by the web console and the API, as pure concatenated PEM blocks: the leaf certificate followed by the intermediate certificate(s), in that order. The root CA certificate must be left out, clients already trust it through their own trust store and only need the intermediates to build the chain. Nothing but `-----BEGIN CERTIFICATE-----` blocks may be present: the `Bag Attributes`, `subject=` and `issuer=` lines that `openssl pkcs12 -nokeys` writes in front of every block are OpenSSL specific text, not part of the PEM format, and Pritunl rejects a certificate that still carries them. Destroying the resource hands the certificate back to Pritunl, which regenerates its self-signed default.",
			},
			"server_key": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Sensitive:    true,
				RequiredWith: []string{"server_cert"},
				StateFunc:    trimSpaceStateFunc,
				Description:  "The PEM encoded private key matching `server_cert`. It is write-only: the value is never read back from the Pritunl API, so it is neither refreshed nor populated on import.",
			},
			"server_port": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
				Description:  "The port the web console and the API listen on. Defaults to the port already configured on the instance. Changing it also requires updating the `url` of the provider.",
			},
			"pin_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"optional", "required", "disabled"}, false),
				Description:  "How the PIN of a user is treated during the secondary authentication, one of `optional`, `required` or `disabled`, the three modes the web console offers. `required` refuses the users that have no PIN set, `disabled` ignores the PINs that are already set. Defaults to the mode already configured on the instance.",
			},
			// the single sign-on attributes are always written together:
			// Pritunl clears every sso_* setting of the instance as soon as the
			// provider it receives is falsy, and rejects a provider that comes
			// without an organization or a domain with a 400
			"sso": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"saml_okta"}, false),
				RequiredWith: []string{"sso_org", "server_sso_url"},
				Description:  "The single sign-on provider the instance authenticates its users against. Only `saml_okta` is accepted, the value the `Okta` entry of the web console stands for: Okta is a SAML integration underneath, which is why an Okta configuration is made of both the SAML attributes (`sso_saml_url`, `sso_saml_issuer_url`, `sso_saml_cert`) and the Okta ones (`sso_okta_app_id`, `sso_okta_token`, `sso_okta_mode`). The console offers `saml_okta_duo` and `saml_okta_yubico` next to it, the same integration with a Duo or a Yubico second factor bolted on, and neither is supported here: they need credentials this resource does not manage. `sso_org` and `server_sso_url` are required along with it, Pritunl answers a single sign-on configuration missing the organization or the domain with a `400`, and a working Okta integration also needs the three SAML attributes, which are left to the instance when they are not configured. Leaving the attribute out of the configuration keeps the single sign-on of the instance exactly as it is: this resource never turns it off, because Pritunl clears every single sign-on credential it holds, for every provider, as soon as it is handed a falsy one. Single sign-on is disabled from the web console instead, and the next plan then reports it as a drift from the configuration.",
			},
			"sso_saml_url": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				RequiredWith: []string{"sso"},
				StateFunc:    trimSpaceStateFunc,
				Description:  "The single sign-on URL of the identity provider, the `SAML 2.0 Endpoint` of the Okta application. Pritunl normalises it the same way as `server_sso_url`, adding an `https://` scheme when the value carries none and lowercasing the host, so it is best configured already normalised: a value Pritunl rewrites reads back differently from the configured one and leaves a plan that never settles.",
			},
			"sso_saml_issuer_url": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				RequiredWith: []string{"sso"},
				StateFunc:    trimSpaceStateFunc,
				Description:  "The issuer URL of the identity provider, the `Identity Provider Issuer` of the Okta application. Normalised by Pritunl just like `sso_saml_url`.",
			},
			"sso_saml_cert": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				RequiredWith: []string{"sso"},
				StateFunc:    normalizeSAMLCertStateFunc,
				Description:  "The X.509 certificate the identity provider signs its SAML assertions with. It is a single certificate of the identity provider and not a chain, so unlike `server_cert` there are no intermediates to concatenate and no leaf to put first. Pritunl accepts and stores whatever string it is handed but its SAML service only works with an armored certificate, so the configured value is canonicalised into PEM on the way in: naked base64, the form Okta's SAML metadata carries in its `X509Certificate` element, is stripped of whitespace, wrapped at 64 columns and armored, while an already armored value only loses its surrounding whitespace. The read reports the value the instance really holds, so an instance still carrying a naked, unusable certificate shows as a drift and converges on the next apply.",
			},
			"sso_okta_app_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				RequiredWith: []string{"sso"},
				StateFunc:    trimSpaceStateFunc,
				Description:  "The id of the Okta application, the last segment of the URL of that application in the Okta administration console. Optional even with single sign-on enabled, as the web console itself points out, but required for Pritunl to check that a user is still attached to the Okta application on every VPN connection.",
			},
			"sso_okta_token": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Sensitive:    true,
				RequiredWith: []string{"sso"},
				StateFunc:    trimSpaceStateFunc,
				Description:  "The Okta API token the instance looks its users up with, a credential with read access to the users and the groups of the Okta organization. `GET /settings` does return it in plain text, but it is deliberately treated the same way as `server_key` and never read back into the state, so it is neither refreshed nor populated on import and the plan compares the configured value against the previously configured one rather than against the token the instance actually holds. It is still handed back to Pritunl untouched by every write that does not configure it, so a token set outside of Terraform is preserved instead of being cleared.",
			},
			// the only managed setting whose empty value means something on its
			// own, which is why the overlay reads it from the raw configuration
			// rather than from the state
			"sso_okta_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				RequiredWith: []string{"sso"},
				ValidateFunc: validation.StringInSlice([]string{"", "passcode", "push", "push_none"}, false),
				Description:  "The secondary factor Okta asks the users for, one of `passcode`, `push`, `push_none` for a push notification whenever the user has a device that takes one, and the empty string for no secondary factor at all, the four entries the web console offers. Pritunl only takes it over while `sso` is exactly `saml_okta`, it silently drops the setting for every other provider, which is why it is required with `sso` here. Leaving it out of the configuration keeps the mode the instance already runs with, and configuring it as the empty string is what turns the secondary factor off.",
			},
			"sso_org": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				RequiredWith: []string{"sso"},
				StateFunc:    trimSpaceStateFunc,
				Description:  "The organization the users authenticated through single sign-on are created in. It is an organization id and not a name, so it is meant to be taken from an organization managed elsewhere in the same configuration, as in `pritunl_organization.sso.id`. Pritunl parses it as a Mongo object id and rejects anything that is not one, and it refuses a single sign-on configuration that comes without an organization with a `400 sso_org_null`.",
			},
			"server_sso_url": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				RequiredWith: []string{"sso"},
				StateFunc:    trimSpaceStateFunc,
				Description:  "The URL the users are sent to for the single sign-on exchange, which for most configurations is the URL the web console itself is reached at. Pritunl requires it as soon as `sso` is set and answers a configuration without it with a `400 sso_url_missing`. Recent versions normalise it before storing it, adding an `https://` scheme when the value carries none and lowercasing the host, so it is best configured already normalised, as in `https://vpn.example.com`: a value Pritunl rewrites reads back differently from the configured one and leaves a plan that never settles.",
			},
			"ipv6": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether the VPN servers of the instance accept IPv6 connections next to IPv4 ones. Defaults to the value already configured on the instance.",
			},
			"sso_cache": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "The `OpenVPN Authentication Cache` of the web console: an 8 hour secondary authentication cache keyed on the client id, the IP address and the MAC address of the client, which lets a client reconnect without going through the secondary authentication again. It works with Duo push, Okta push, OneLogin push, Duo passcodes and YubiKeys, and is supported by every OpenVPN client. Not to be confused with `sso_client_cache`, the cache of the Pritunl client, a different setting with a deceptively similar name.",
			},
			"sso_client_cache": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "The `Pritunl Authentication Cache` of the web console: a 7 day secondary authentication cache kept as a token on the client itself, which lets a client reconnect without going through the secondary authentication again. It covers the same secondary factors as `sso_cache` but is only supported by the Pritunl client. Not to be confused with `sso_cache`, the cache of every OpenVPN client, a different setting with a deceptively similar name.",
			},
			"restrict_import": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether the users have to import their profile through a Pritunl URI, which keeps them from downloading the profile directly and using it with another OpenVPN client. Defaults to the value already configured on the instance.",
			},
			"client_reconnect": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether the Pritunl client reconnects on its own once a connection drops. The web console only shows the toggle while single sign-on is enabled, but that is a detail of the console: `PUT /settings` takes the setting either way, and so does this resource, which never ties it to `sso`.",
			},
		},
		CreateContext: resourceCreateSettings,
		ReadContext:   resourceReadSettings,
		UpdateContext: resourceUpdateSettings,
		DeleteContext: resourceDeleteSettings,
		CustomizeDiff: customizeSettingsDiff,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

// customizeSettingsDiff puts the change of sso_okta_mode back on the plan when
// it is being turned off.
//
// The empty mode is the Okta secondary factor disabled, a value of its own, but
// the SDK reads the empty string of a computed attribute as "not configured"
// and drops that change before it ever reaches the plan, "to align with legacy
// behavior" as it puts it. Setting the value by hand is the way out the SDK
// itself leaves open: a value set here is marked as customized, which is the
// one thing that keeps it from being dropped a second time.
func customizeSettingsDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	config := d.GetRawConfig()
	if config.IsNull() || !config.IsKnown() {
		return nil
	}

	mode := config.GetAttr("sso_okta_mode")
	if mode.IsNull() || !mode.IsKnown() || mode.AsString() != "" {
		return nil
	}

	// the state is what the change is against: d.Get merges the configuration
	// on top of it and would answer with the very empty mode being planned
	current, _ := d.GetChange("sso_okta_mode")
	if current.(string) == "" {
		return nil
	}

	return d.SetNew("sso_okta_mode", "")
}

// Pritunl strips the surrounding whitespace of the PEM values before storing
// them, so the configured value is normalised the same way to keep it in sync
// with what the API returns.
func trimSpaceStateFunc(value interface{}) string {
	return strings.TrimSpace(value.(string))
}

// normalizeSAMLCert canonicalises the identity provider's signing certificate
// into armored PEM. Okta's SAML metadata carries the naked base64 of its
// X509Certificate element while Pritunl's SAML service only works with an
// armored certificate, so a value without an armor block is stripped of all
// whitespace, wrapped at 64 columns and armored; a value that already carries
// one only has its surrounding whitespace trimmed, the way Pritunl stores it,
// and the empty string stays empty so the attribute keeps its keep-as-is
// semantics. Only the configured value goes through here (StateFunc): the
// read keeps the instance's own value, so an instance holding a naked,
// unusable certificate surfaces as drift and converges instead of being
// masked by the normalisation.
func normalizeSAMLCert(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || strings.Contains(trimmed, "-----BEGIN") {
		return trimmed
	}

	compact := strings.Join(strings.Fields(trimmed), "")

	var pem strings.Builder
	pem.WriteString("-----BEGIN CERTIFICATE-----\n")
	for len(compact) > 64 {
		pem.WriteString(compact[:64])
		pem.WriteByte('\n')
		compact = compact[64:]
	}
	pem.WriteString(compact)
	pem.WriteString("\n-----END CERTIFICATE-----")

	return pem.String()
}

func normalizeSAMLCertStateFunc(value interface{}) string {
	return normalizeSAMLCert(value.(string))
}

// The boolean settings the resource manages. They share the overlay and the
// read because they all behave the same way, and because a boolean is the one
// kind of value the overlay cannot decide on with d.Get alone, see
// settingConfigured.
var settingsBoolAttributes = []string{
	"client_reconnect",
	"ipv6",
	"restrict_import",
	"sso_cache",
	"sso_client_cache",
}

// The single sign-on settings the resource manages, which Pritunl clears
// together as soon as it is handed a falsy sso, and which it therefore only
// ever receives as a whole.
var settingsSsoAttributes = []string{
	"server_sso_url",
	"sso_okta_app_id",
	"sso_okta_token",
	"sso_org",
	"sso_saml_cert",
	"sso_saml_issuer_url",
	"sso_saml_url",
}

// Every attribute the resource manages, used by the update to tell whether the
// configuration has anything to write at all.
var settingsAttributes = []string{
	"client_reconnect",
	"ipv6",
	"pin_mode",
	"restrict_import",
	"server_cert",
	"server_key",
	"server_port",
	"server_sso_url",
	"sso",
	"sso_cache",
	"sso_client_cache",
	"sso_okta_app_id",
	"sso_okta_mode",
	"sso_okta_token",
	"sso_org",
	"sso_saml_cert",
	"sso_saml_issuer_url",
	"sso_saml_url",
}

// settingConfigured reports whether an attribute is written down in the
// configuration, as opposed to being left to the value the instance already
// holds.
//
// It exists for the settings whose empty value is a value of its own: false is
// both a boolean an operator can ask for and the zero value d.Get falls back
// on, and the same goes for the empty string of sso_okta_mode, so neither
// d.Get nor d.GetOk can tell "turn this off" from "do not manage this". d.Get
// is not even stable across the lifecycle: an optional and computed attribute
// missing from the configuration is still unknown while the resource is being
// created, and reads back as the zero value, which would turn a setting off on
// the very first apply of a configuration that never mentions it. The raw
// configuration is the only place where the difference survives, so it is what
// decides.
func settingConfigured(d *schema.ResourceData, key string) bool {
	_, configured := settingConfiguredString(d, key)

	return configured
}

// settingConfiguredString returns the value an attribute carries in the raw
// configuration, and whether the configuration mentions it at all.
//
// It is the only reliable way to read back a setting whose empty value means
// something: the SDK turns the plan of a computed attribute set to the empty
// string into an unknown value, so d.Get answers with the value the instance
// had before instead of the one it is being given.
func settingConfiguredString(d *schema.ResourceData, key string) (string, bool) {
	config := d.GetRawConfig()
	if config.IsNull() || !config.IsKnown() {
		return "", false
	}

	value := config.GetAttr(key)
	if value.IsNull() || !value.IsKnown() {
		return "", false
	}

	if value.Type() != cty.String {
		return "", true
	}

	return value.AsString(), true
}

func settingsHaveChanges(d *schema.ResourceData) bool {
	for _, attribute := range settingsAttributes {
		if d.HasChange(attribute) {
			return true
		}
	}

	// Turning the Okta secondary factor off is the one change HasChange cannot
	// see: the SDK plans the empty value of a computed attribute as unknown,
	// and an unknown value reads back as the one the instance had, so both
	// sides of the comparison end up being the mode that is on its way out.
	// The raw configuration is where the difference is left, here as well.
	if mode, configured := settingConfiguredString(d, "sso_okta_mode"); configured {
		return mode != d.Get("sso_okta_mode").(string)
	}

	return false
}

func resourceCreateSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	// the settings object always exists, there is nothing to create: the whole
	// object is read, the managed attributes are overlaid on it and all of it
	// is written back
	settings, err := apiClient.GetSettings()
	if err != nil {
		return diag.FromErr(err)
	}

	portChanged := overlaySettings(d, settings)

	if err = apiClient.UpdateSettings(settings); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(settingsResourceId)

	return finishSettingsWrite(ctx, d, meta, portChanged)
}

// Uses for importing
func resourceReadSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	settings, err := apiClient.GetSettings()
	if err != nil {
		return diag.FromErr(err)
	}

	// only the managed attributes make it into the state, the rest of the
	// settings object is a write time detail and never leaves this package
	d.Set("server_cert", strings.TrimSpace(settings.String("server_cert")))
	d.Set("server_port", settings.Int("server_port"))
	d.Set("pin_mode", settings.String("pin_mode"))

	// sso reads as the boolean false on an instance where single sign-on has
	// never been configured, which String turns into the empty string, the very
	// same thing it means to Pritunl
	d.Set("sso", settings.String("sso"))
	d.Set("sso_org", settings.String("sso_org"))
	d.Set("sso_saml_url", settings.String("sso_saml_url"))
	d.Set("sso_saml_issuer_url", settings.String("sso_saml_issuer_url"))
	// Deliberately NOT normalised: the read reports what the instance really
	// holds, so an instance still carrying a naked certificate — which its
	// SAML service cannot use — shows as a drift from the canonical
	// configuration and converges on the next apply. Normalising here would
	// make the broken value compare equal and never be repaired.
	d.Set("sso_saml_cert", strings.TrimSpace(settings.String("sso_saml_cert")))
	d.Set("sso_okta_app_id", settings.String("sso_okta_app_id"))
	d.Set("sso_okta_mode", settings.String("sso_okta_mode"))
	d.Set("server_sso_url", settings.String("server_sso_url"))

	for _, attribute := range settingsBoolAttributes {
		d.Set(attribute, settings.Bool(attribute))
	}

	// GET /settings does return server_key and sso_okta_token in plain text,
	// but they are deliberately not stored: they would copy the private key of
	// the instance and an Okta credential into the state of every configuration
	// that only manages server_port, and the absence of server_key from the
	// state is what tells the delete that this resource never pushed a
	// certificate of its own.

	return nil
}

func resourceUpdateSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	if !settingsHaveChanges(d) {
		return resourceReadSettings(ctx, d, meta)
	}

	settings, err := apiClient.GetSettings()
	if err != nil {
		return diag.FromErr(err)
	}

	portChanged := overlaySettings(d, settings)

	if err = apiClient.UpdateSettings(settings); err != nil {
		return diag.FromErr(err)
	}

	return finishSettingsWrite(ctx, d, meta, portChanged)
}

// overlaySettings writes the attributes managed by this resource on top of the
// settings object that was just read from the instance, and reports whether the
// web server is about to move to another port.
//
// Everything it does not touch is left in the object exactly as the API
// returned it, which is what the surrounding full object round trip hands back.
// That includes the settings the API computes rather than stores, such as
// public_address, which falls back to the address Pritunl detected on its own:
// echoing back a value that was just read is a no-op, echoing back a stale one
// would pin it as a manual override, so the settings are always read as part of
// the write and never cached between runs.
func overlaySettings(d *schema.ResourceData, settings pritunl.Settings) bool {
	cert := strings.TrimSpace(d.Get("server_cert").(string))
	key := strings.TrimSpace(d.Get("server_key").(string))

	// the pair is only pushed when this resource manages one: the certificate
	// already installed on the instance is otherwise handed back untouched.
	// Both always travel together, a key without its certificate would leave
	// Pritunl unable to serve HTTPS at all.
	if cert != "" && key != "" {
		settings["server_cert"] = cert
		settings["server_key"] = key
	}

	overlaySettingsString(d, settings, "pin_mode")

	// Single sign-on is written as a whole or not at all. Pritunl clears
	// sso_org, server_sso_url and the credentials of every provider it knows as
	// soon as the sso it receives is falsy, so the provider has to be in the
	// object whenever this resource manages one, and it has to come along with
	// the organization and the domain, which Pritunl refuses to do without. A
	// configuration that manages no single sign-on hands no provider over at
	// all, blank or otherwise, and the round trip gives the instance its own
	// one back along with the credentials that belong to it.
	if sso := strings.TrimSpace(d.Get("sso").(string)); sso != "" {
		settings["sso"] = sso

		for _, attribute := range settingsSsoAttributes {
			overlaySettingsString(d, settings, attribute)
		}

		// the empty mode is the secondary factor turned off, a value of its
		// own that has to be told apart from an unmanaged attribute, and one
		// the plan hands over as unknown rather than as the empty string
		if mode, configured := settingConfiguredString(d, "sso_okta_mode"); configured {
			settings["sso_okta_mode"] = mode
		}
	}

	for _, attribute := range settingsBoolAttributes {
		if settingConfigured(d, attribute) {
			settings[attribute] = d.Get(attribute).(bool)
		}
	}

	portChanged := false

	if port, ok := d.GetOk("server_port"); ok {
		portChanged = port.(int) != settings.Int("server_port")
		settings["server_port"] = port.(int)
	}

	return portChanged
}

// overlaySettingsString hands a string attribute over only when it holds a
// value, which is what leaves the setting of the instance in place when the
// attribute is not managed. An empty string is never written: none of the
// string settings managed here can be cleared through this resource, and the
// write only ones read back as empty by design, so writing them empty would
// clear on the instance what the state simply does not know.
func overlaySettingsString(d *schema.ResourceData, settings pritunl.Settings, key string) {
	if value := strings.TrimSpace(d.Get(key).(string)); value != "" {
		settings[key] = value
	}
}

func resourceDeleteSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	// server_key is never refreshed from the api, so it only holds a value when
	// this resource pushed a certificate itself. Without it the certificate in
	// the state merely comes from a read and must not be touched.
	if d.Get("server_key").(string) == "" {
		d.SetId("")

		return nil
	}

	settings, err := apiClient.GetSettings()
	if err != nil {
		return diag.FromErr(err)
	}

	// the settings object cannot be deleted, destroying the resource resets the
	// managed certificate instead. An empty string is enough to clear it:
	// Pritunl turns any falsy certificate into null and then regenerates its
	// self-signed default while restarting the web server.
	//
	// The certificate is the only managed setting with a meaningful "unset":
	// none of the others has a value that would mean "the one from before", and
	// turning off single sign-on would discard every credential Pritunl holds
	// for it. They are all handed back untouched by the round trip, server_port
	// included.
	settings["server_cert"] = ""
	settings["server_key"] = ""

	if err = apiClient.UpdateSettings(settings); err != nil {
		return diag.FromErr(err)
	}

	if err = waitForWebServer(ctx, apiClient); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func finishSettingsWrite(ctx context.Context, d *schema.ResourceData, meta interface{}, portChanged bool) diag.Diagnostics {
	if portChanged {
		// the web console and the API are served on another port from now on,
		// waiting for them on the previous one or reading the settings back
		// would only time out
		return diag.Diagnostics{{
			Severity: diag.Warning,
			Summary:  "The Pritunl web server port has been changed",
			Detail:   "Pritunl restarted its web server on the new port, the `url` of the pritunl provider has to be updated accordingly before the next Terraform run.",
		}}
	}

	if err := waitForWebServer(ctx, meta.(pritunl.Client)); err != nil {
		return diag.FromErr(err)
	}

	return resourceReadSettings(ctx, d, meta)
}

// Applying a certificate, a key or a port makes Pritunl restart its web server
// shortly after it answered the request, which leaves the API unreachable for a
// moment. Reading the settings back right away would fail, and the previous
// server is still up for a short while, so the endpoint has to answer a few
// times in a row before it is considered restarted.
func waitForWebServer(ctx context.Context, apiClient pritunl.Client) error {
	if err := sleepContext(ctx, settingsRestartDelay); err != nil {
		return err
	}

	deadline := time.Now().Add(settingsRestartTimeout)
	successes := 0

	for {
		err := apiClient.TestApiCall()
		if err == nil {
			successes++
			if successes >= settingsRestartSuccesses {
				return nil
			}

			if err = sleepContext(ctx, settingsRestartSettle); err != nil {
				return err
			}

			continue
		}

		successes = 0

		if time.Now().After(deadline) {
			return fmt.Errorf("timeout while waiting for the pritunl web server to restart with the new settings: %s", err)
		}

		if err = sleepContext(ctx, settingsRestartPoll); err != nil {
			return err
		}
	}
}

func sleepContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
