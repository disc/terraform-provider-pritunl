package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/lfventura/terraform-provider-pritunl/internal/pritunl"
)

// The characters Pritunl keeps in a username, everything else is stripped
// before the account is stored. It filters the value down to letters, digits
// and a fixed set of punctuation, then lowercases it.
var administratorUsernameSafeChars = map[rune]bool{
	'-': true, '=': true, '_': true, '@': true, '.': true, ':': true, '/': true,
	'!': true, '#': true, '$': true, '&': true, '*': true, '+': true,
	'?': true, '^': true, '|': true, '~': true,
}

// Pritunl only keeps the first twelve characters of a YubiKey OTP, which is
// the public id of the key, and discards the rest.
const administratorYubikeyIdLength = 12

// An administrator is addressed by the object id Pritunl gave it. The import
// takes a username as well, and this is what tells the two apart.
var administratorObjectIdPattern = regexp.MustCompile(`^[0-9a-fA-F]{24}$`)

// The flags of an administrator. They share the overlay, the create body and
// the read because they all behave the same way, and because a boolean is the
// one kind of value the overlay cannot decide on with d.Get alone, see
// administratorConfigured.
var administratorBoolAttributes = []string{
	"auth_api",
	"disabled",
	"local_otp_auth",
	"otp_auth",
	"super_user",
}

func resourceAdministrator() *schema.Resource {
	return &schema.Resource{
		Description: "The administrator resource allows managing the administrator accounts of a Pritunl instance, the logins of its web console and of its API. They have nothing to do with `pritunl_user`, which manages the VPN profiles of the end users inside an organization: an administrator is not a member of an organization and cannot connect to a VPN server, it operates Pritunl itself.\n\n" +
			"**Every request behind this resource requires the credentials configured on the `pritunl` provider to belong to a super user.** `GET`, `POST`, `PUT` and `DELETE` on `/admin` are all refused with a `400 requires_super_user` otherwise, so the resource is unusable with an administrator that only holds API access.\n\n" +
			"Every write is a read-modify-write of the complete administrator object, the same way the Pritunl web console itself works: `PUT /admin/<id>` is a full replace and not a partial update, and a request that leaves a field out has that field reset rather than kept. A rename alone is enough to strip the super user flag off an account, turn its API access off and rotate the API credentials that come with it, so the account is read back from the instance immediately before each write and handed over again with the managed attributes overlaid on top of it.\n\n" +
			"Take care when the account managed here is the very one the `pritunl` provider authenticates as: setting `auth_api = false` or `disabled = true` on it revokes the credentials Terraform is holding, and every later request of the same run, the refresh of this resource included, fails with a `401`.",
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				StateFunc:   administratorUsernameStateFunc,
				Description: "The login of the administrator. Pritunl normalises it before storing it, dropping every character that is neither a letter, a digit nor one of `-=_@.:/!#$&*+?^|~` and lowercasing the rest, and the configured value is normalised the same way so that it keeps matching what the API reports back. Usernames are unique, Pritunl answers a duplicate with a `400 admin_username_exists`.",
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				Description: "The password of the administrator. It is required to create one and optional afterwards: `GET /admin/<id>` never returns a password, not even a hash, so the value is write-only and is treated exactly the way `server_key` and `sso_okta_token` are treated by `pritunl_settings`. It is never read back, which means it is neither refreshed nor populated on import, and the plan compares the configured value against the previously configured one rather than against the password the account actually holds.\n\n" +
					"Leaving it out of an existing configuration keeps the password as it is instead of clearing it: the backend only takes a password over when the request carries a non-empty one.",
			},
			"yubikey_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				StateFunc:   administratorYubikeyIdStateFunc,
				Description: "The public id of the YubiKey the administrator authenticates with, the first twelve characters of an OTP the key emits. A longer value is truncated to those twelve characters, by Pritunl and by this resource alike, so that pasting a whole OTP works and still matches what the API reports back. Configuring it as the empty string removes the YubiKey from the account, leaving the attribute out keeps whatever the account already carries.",
			},
			"super_user": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether the administrator may administer the instance itself, which among other things is what every request of this resource requires. Defaults to `false` on a new account and to the value the account already holds otherwise. Pritunl refuses to take the flag off the last super user of the instance with a `400 no_super_users`.",
			},
			"auth_api": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				Description: "Whether the administrator may authenticate against the API with `token` and `secret` rather than through the web console. Defaults to `false` on a new account and to the value the account already holds otherwise.\n\n" +
					"Turning it off does not merely revoke the credentials, it discards them: Pritunl clears the token and the secret of the account and generates a fresh pair the next time the account is written, so turning it back on later hands out different credentials than before.",
			},
			"token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
				Description: "The API token of the administrator, as `PRITUNL_TOKEN` and the `token` of this provider take it. It is read-only: Pritunl generates it and regenerates it as soon as a request carries anything in that field at all, so there is no value to configure, and it is only reported while `auth_api` is `true`, the credentials of an account without API access being refused with a `401`.\n\n" +
					"There is no attribute to rotate it with. Rotating is replacing the account, `terraform apply -replace=pritunl_administrator.example`, which creates it again with a fresh pair.",
			},
			"secret": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The API secret of the administrator, as `PRITUNL_SECRET` and the `secret` of this provider take it. Read-only and rotated the same way as `token`.",
			},
			"disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether the account is locked out of the web console and of the API. Defaults to `false` on a new account and to the value the account already holds otherwise. Pritunl refuses to disable the last enabled super user of the instance with a `400 no_admins_enabled`.",
			},
			"otp_auth": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				Description: "Whether the administrator is asked for a two-step authentication code from an authenticator application. Defaults to `false` on a new account and to the value the account already holds otherwise. Mutually exclusive with `local_otp_auth`, Pritunl refuses an account carrying both with a `400 admin_invalid_otp`, and a configuration enabling both is refused at plan time.\n\n" +
					"Turning it on makes Pritunl generate the shared secret of the account, turning it off discards it, so an account that goes through both ends up with a different secret than it started with and has to be enrolled again. The secret itself is never exposed by this resource: it is enrolled from the web console, which is the only place that shows the QR code.",
			},
			"local_otp_auth": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				Description: "Whether the administrator is asked for a two-step authentication code issued by the Pritunl instance itself rather than by an authenticator application. Defaults to `false` on a new account and to the value the account already holds otherwise. Mutually exclusive with `otp_auth`.\n\n" +
					"It only exists on the Pritunl versions that ship it, older ones leave the field out of their responses and ignore it in a request. Asking for it on such an instance is answered with an error rather than silently dropped.",
			},
			"default": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this is the account the instance was seeded with, the one Pritunl creates on the first start and prints the generated password of. It is read-only and Pritunl clears it on its own the first time that account changes its password. It is reported so that a configuration adopting the seeded account through an import can tell it apart from the accounts it created itself.",
			},
		},
		CreateContext: resourceCreateAdministrator,
		ReadContext:   resourceReadAdministrator,
		UpdateContext: resourceUpdateAdministrator,
		DeleteContext: resourceDeleteAdministrator,
		CustomizeDiff: customizeAdministratorDiff,
		Importer: &schema.ResourceImporter{
			StateContext: importAdministrator,
		},
	}
}

// importAdministrator adopts an existing administrator, by the object id
// Pritunl addresses it with or by its username.
//
// The id is what the API itself uses and what the state ends up holding either
// way, but it is not something an operator has at hand: adopting the account
// the instance was seeded with means looking its id up through the API first,
// which is exactly the kind of detour an import is meant to avoid. A value
// that is not an object id is therefore looked up among the administrators of
// the instance instead. Usernames are unique, so the lookup is unambiguous,
// and a username that happens to be twenty-four hexadecimal characters long is
// read as an id, which is the one case this cannot tell apart.
func importAdministrator(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if administratorObjectIdPattern.MatchString(d.Id()) {
		return []*schema.ResourceData{d}, nil
	}

	apiClient := meta.(pritunl.Client)

	administrators, err := apiClient.GetAdministrators()
	if err != nil {
		return nil, err
	}

	username := administratorUsername(d.Id())

	for _, administrator := range administrators {
		if administrator.String("username") == username {
			d.SetId(administrator.String("id"))

			return []*schema.ResourceData{d}, nil
		}
	}

	return nil, fmt.Errorf("no administrator named %q exists on this pritunl instance, "+
		"an administrator is imported by its username or by the object id the api addresses it with", username)
}

// customizeAdministratorDiff refuses the pair of two-step authentication modes
// Pritunl does not accept, and puts the change of yubikey_id back on the plan
// when the YubiKey is being removed.
//
// The two modes are mutually exclusive, an account carrying both is answered
// with a 400 admin_invalid_otp, and catching that at plan time beats letting
// the apply run into it. It is deliberately not a ConflictsWith: that one
// refuses the two attributes being written down together, whatever they are
// set to, and only enabling both is a problem here. Turning both off in the
// same configuration is an ordinary thing to ask for, and so is turning one on
// while turning the other off, which is in fact the only way to move an
// account from one mode to the other.
func customizeAdministratorDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	config := d.GetRawConfig()
	if config.IsNull() || !config.IsKnown() {
		return nil
	}

	otpAuth, otpAuthConfigured := administratorConfiguredBool(config, "otp_auth")
	localOtpAuth, localOtpAuthConfigured := administratorConfiguredBool(config, "local_otp_auth")

	if otpAuthConfigured && localOtpAuthConfigured && otpAuth && localOtpAuth {
		return errors.New("otp_auth and local_otp_auth cannot both be enabled: an administrator authenticates " +
			"either with an authenticator application or with a code issued by the pritunl instance, and pritunl " +
			"refuses an account carrying both with a 400 admin_invalid_otp")
	}

	// The empty id is the YubiKey taken off the account, a value of its own,
	// but the SDK reads the empty string of a computed attribute as "not
	// configured" and drops that change before it ever reaches the plan, "to
	// align with legacy behavior" as it puts it. Setting the value by hand is
	// the way out the SDK itself leaves open: a value set here is marked as
	// customized, which is the one thing that keeps it from being dropped a
	// second time. The same treatment is what makes sso_okta_mode removable in
	// pritunl_settings.
	yubikeyId := config.GetAttr("yubikey_id")
	if yubikeyId.IsNull() || !yubikeyId.IsKnown() || yubikeyId.AsString() != "" {
		return nil
	}

	// the state is what the change is against: d.Get merges the configuration
	// on top of it and would answer with the very empty id being planned
	current, _ := d.GetChange("yubikey_id")
	if current.(string) == "" {
		return nil
	}

	return d.SetNew("yubikey_id", "")
}

// administratorConfiguredBool returns the value a boolean attribute carries in
// the raw configuration, and whether the configuration mentions it at all.
// False being both a flag an operator can ask for and the zero value of the
// attribute, the raw configuration is the only place the two are told apart.
func administratorConfiguredBool(config cty.Value, key string) (bool, bool) {
	value := config.GetAttr(key)
	if value.IsNull() || !value.IsKnown() {
		return false, false
	}

	return value.True(), true
}

// administratorUsername normalises a username the way Pritunl does before
// storing it, so that the configured value keeps matching the one the API
// reports back instead of leaving a plan that never settles.
func administratorUsername(value string) string {
	var normalized strings.Builder

	for _, character := range value {
		if unicode.IsLetter(character) || unicode.IsDigit(character) ||
			administratorUsernameSafeChars[character] {
			normalized.WriteRune(character)
		}
	}

	return strings.ToLower(normalized.String())
}

func administratorUsernameStateFunc(value interface{}) string {
	return administratorUsername(value.(string))
}

// administratorYubikeyId keeps the public id of a YubiKey OTP the way Pritunl
// does, so that pasting a whole OTP into the configuration does not leave a
// plan that never settles either.
func administratorYubikeyId(value string) string {
	trimmed := []rune(strings.TrimSpace(value))
	if len(trimmed) > administratorYubikeyIdLength {
		trimmed = trimmed[:administratorYubikeyIdLength]
	}

	return string(trimmed)
}

func administratorYubikeyIdStateFunc(value interface{}) string {
	return administratorYubikeyId(value.(string))
}

// administratorConfigured reports whether an attribute is written down in the
// configuration, as opposed to being left to the value the account already
// holds.
//
// It exists for the attributes whose empty value is a value of its own: false
// is both a flag an operator can ask for and the zero value d.Get falls back
// on, and the same goes for the empty yubikey_id, so neither d.Get nor d.GetOk
// can tell "turn this off" from "do not manage this". d.Get is not even stable
// across the lifecycle: an optional and computed attribute missing from the
// configuration is still unknown while the resource is being created, and
// reads back as the zero value, which would take a flag off an account on the
// very first apply of a configuration that never mentions it. The raw
// configuration is the only place where the difference survives, so it is what
// decides.
//
// pritunl_settings solves the same problem the same way. The helpers are
// deliberately not shared between the two resources: they are a handful of
// lines each and keeping them apart lets either resource change the way it
// reads its configuration without dragging the other one along.
func administratorConfigured(d *schema.ResourceData, key string) bool {
	_, configured := administratorConfiguredString(d, key)

	return configured
}

// administratorConfiguredString returns the value an attribute carries in the
// raw configuration, and whether the configuration mentions it at all.
func administratorConfiguredString(d *schema.ResourceData, key string) (string, bool) {
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

func resourceCreateAdministrator(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	password := d.Get("password").(string)
	if password == "" {
		return diag.Diagnostics{{
			Severity:      diag.Error,
			Summary:       "A password is required to create an administrator",
			Detail:        "Pritunl refuses an administrator without a password. The attribute is optional because the password is write-only and is never read back from the API, which is what makes it optional on every later apply, but creating the account needs one.",
			AttributePath: cty.GetAttrPath("password"),
		}}
	}

	// nothing to merge, the account does not exist yet: unlike the update, the
	// create takes only the attributes the configuration carries, and Pritunl
	// defaults every flag it does not receive to false
	administrator := pritunl.Administrator{
		"username": administratorUsername(d.Get("username").(string)),
		"password": password,
	}

	if yubikeyId, configured := administratorConfiguredString(d, "yubikey_id"); configured {
		administrator["yubikey_id"] = administratorYubikeyId(yubikeyId)
	}

	for _, attribute := range administratorBoolAttributes {
		if administratorConfigured(d, attribute) {
			administrator[attribute] = d.Get(attribute).(bool)
		}
	}

	created, err := apiClient.CreateAdministrator(administrator)
	if err != nil {
		return administratorDiagnostics(err)
	}

	d.SetId(created.String("id"))

	if err = checkAdministratorLocalOtpAuth(d, created); err != nil {
		return diag.FromErr(err)
	}

	return resourceReadAdministrator(ctx, d, meta)
}

// Uses for importing
func resourceReadAdministrator(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	administrator, err := apiClient.GetAdministrator(d.Id())
	if err != nil {
		if errors.Is(err, pritunl.ErrAdministratorNotFound) {
			d.SetId("")

			return nil
		}

		return administratorDiagnostics(err)
	}

	d.Set("username", administrator.String("username"))
	d.Set("yubikey_id", administrator.String("yubikey_id"))
	d.Set("default", administrator.Bool("default"))

	for _, attribute := range administratorBoolAttributes {
		// an instance that predates local_otp_auth leaves the field out of its
		// responses, and false is what it means to it. A configuration asking
		// for the field on such an instance never reaches this point, both
		// write paths refuse it beforehand, see checkAdministratorLocalOtpAuth
		d.Set(attribute, administrator.Bool(attribute))
	}

	// the credentials only authenticate while the account has API access, a
	// request signed with the pair of an account without it is refused with a
	// 401, and Pritunl replaces the pair as soon as the access is turned back
	// on. Reporting them for an account that has no API access would hand out
	// credentials that neither work nor last.
	if administrator.Bool("auth_api") {
		d.Set("token", administrator.String("token"))
		d.Set("secret", administrator.String("secret"))
	} else {
		d.Set("token", "")
		d.Set("secret", "")
	}

	// the password is never read back: GET returns no password at all, not
	// even a hash, the same way pritunl_settings never reads server_key back

	return nil
}

func resourceUpdateAdministrator(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	// PUT /admin/<id> is a full replace: the account is read back immediately
	// before the write so that everything this resource does not manage is
	// handed over again untouched, the super user flag and the API access
	// included
	administrator, err := apiClient.GetAdministrator(d.Id())
	if err != nil {
		return administratorDiagnostics(err)
	}

	if err = overlayAdministrator(d, administrator); err != nil {
		return diag.FromErr(err)
	}

	if err = apiClient.UpdateAdministrator(d.Id(), administrator); err != nil {
		return administratorDiagnostics(err)
	}

	return resourceReadAdministrator(ctx, d, meta)
}

// overlayAdministrator writes the attributes managed by this resource on top
// of the administrator that was just read from the instance.
//
// Everything it does not touch is left in the object exactly as the API
// returned it, which is what the surrounding full object round trip hands
// back, the fields a newer Pritunl adds to the endpoint included. The write
// path never sends a token, a secret or an otp_secret, those are cleared by
// the client right before the request, see Administrator.normalize.
func overlayAdministrator(d *schema.ResourceData, administrator pritunl.Administrator) error {
	administrator["username"] = administratorUsername(d.Get("username").(string))

	// only a non-empty password is handed over: the attribute is write-only
	// and reads back as empty on an account this configuration never gave a
	// password to, and Pritunl leaves the password alone for an empty one
	// anyway
	if password := d.Get("password").(string); password != "" {
		administrator["password"] = password
	}

	// the empty id is the YubiKey taken off the account, a value of its own
	// that has to be told apart from an unmanaged attribute
	if yubikeyId, configured := administratorConfiguredString(d, "yubikey_id"); configured {
		administrator["yubikey_id"] = administratorYubikeyId(yubikeyId)
	}

	for _, attribute := range administratorBoolAttributes {
		if administratorConfigured(d, attribute) {
			administrator[attribute] = d.Get(attribute).(bool)
		}
	}

	return checkAdministratorLocalOtpAuth(d, administrator)
}

// checkAdministratorLocalOtpAuth turns a local_otp_auth the instance knows
// nothing about into an error instead of letting it pass unnoticed.
//
// Pritunl only grew the field at some point: an older instance leaves it out
// of its responses and drops it from a request without a word, so a
// configuration asking for it would apply cleanly and never take effect. The
// account read back from the API is what decides, an account of an instance
// that supports the field always carries it.
func checkAdministratorLocalOtpAuth(d *schema.ResourceData, administrator pritunl.Administrator) error {
	if administrator.Has("local_otp_auth") {
		return nil
	}

	if !administratorConfigured(d, "local_otp_auth") || !d.Get("local_otp_auth").(bool) {
		return nil
	}

	return errors.New("this pritunl instance does not support local_otp_auth, it reports no such field for its " +
		"administrators, and it silently ignores the one a request carries. Use otp_auth for a two-step " +
		"authentication code from an authenticator application, or upgrade the instance")
}

func resourceDeleteAdministrator(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(pritunl.Client)

	if err := apiClient.DeleteAdministrator(d.Id()); err != nil {
		// an account that is already gone has nothing left to delete, and
		// Pritunl answers a delete of one by failing while building the
		// response rather than with a 404 of its own, so the account is looked
		// up before the failure is reported
		if _, readErr := apiClient.GetAdministrator(d.Id()); errors.Is(readErr, pritunl.ErrAdministratorNotFound) {
			d.SetId("")

			return nil
		}

		return administratorDiagnostics(err)
	}

	d.SetId("")

	return nil
}

// administratorDiagnostics turns the error of a refused request into a
// diagnostic that says what Pritunl was protecting.
//
// The guards of the endpoint are the reason this exists: a destroy that would
// leave the instance without a super user, or a write that would lock the last
// one out, is refused by Pritunl with an error code of its own, and reporting
// that as a bare non-200 leaves an operator with nothing to act on. The
// message the API sends along is kept as it is and the detail explains what
// caused it.
func administratorDiagnostics(err error) diag.Diagnostics {
	var apiError *pritunl.ApiError
	if !errors.As(err, &apiError) {
		return diag.FromErr(err)
	}

	summary := apiError.Message
	if summary == "" {
		summary = "The Pritunl API refused the administrator request"
	}

	var detail string

	switch apiError.Code {
	case "no_admins":
		detail = "Pritunl refuses to remove the last super user of the instance, which would leave nobody able to " +
			"administer it. Only the super users that are enabled are counted, so a super user carrying " +
			"`disabled = true` cannot be removed either while it is the only other one, even though removing it " +
			"could not lock anybody out. Enable it again, or create another super user administrator, before " +
			"destroying it, or take it out of the Terraform state with `terraform state rm` if it is meant to stay " +
			"on the instance unmanaged."
	case "no_super_users":
		detail = "Pritunl refuses to take the super user flag off the last super user of the instance, which would " +
			"leave nobody able to administer it. Grant `super_user = true` to another administrator first."
	case "no_admins_enabled":
		detail = "Pritunl refuses to disable the last enabled super user of the instance, which would lock everybody " +
			"out of the web console. Enable another super user administrator first."
	case "requires_super_user":
		detail = "Every request on `/admin` requires a super user. The `pritunl` provider is configured with the " +
			"credentials of an administrator that is not one, so no administrator can be read, created, updated or " +
			"deleted with them."
	case "admin_username_exists":
		detail = "Usernames are unique across the administrators of the instance. Pick another one, or import the " +
			"existing account with `terraform import pritunl_administrator.example <username>`."
	case "admin_invalid_otp":
		detail = "An administrator authenticates either with an authenticator application (`otp_auth`) or with a " +
			"code issued by the Pritunl instance (`local_otp_auth`), never with both."
	default:
		if apiError.StatusCode == 401 {
			detail = "The credentials the `pritunl` provider authenticates with were refused. When this happens " +
				"right after a write, the administrator managed here is most likely the very one those credentials " +
				"belong to, and the write revoked them: `auth_api = false` discards the API credentials of an " +
				"account and `disabled = true` locks it out entirely."
		}
	}

	if detail == "" {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{{
		Severity: diag.Error,
		Summary:  summary,
		Detail:   detail,
	}}
}
