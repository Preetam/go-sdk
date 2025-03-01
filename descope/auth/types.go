package auth

import (
	"regexp"

	"github.com/descope/go-sdk/descope/logger"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"golang.org/x/exp/maps"
)

// TOTPResponse - returns all relevant data to complete a TOTP registration
// One can select which method of registration to use for handshaking with an Authenticator app
type TOTPResponse struct {
	ProvisioningURL string `json:"provisioningURL,omitempty"`
	Image           string `json:"image,omitempty"`
	Key             string `json:"key,omitempty"`
}

type AuthenticationInfo struct {
	SessionToken *Token        `json:"token,omitempty"`
	User         *UserResponse `json:"user,omitempty"`
	FirstSeen    bool          `json:"firstSeen,omitempty"`
}

type WebAuthnTransactionResponse struct {
	TransactionID string `json:"transactionId,omitempty"`
	Options       string `json:"options,omitempty"`
	Create        bool   `json:"create,omitempty"`
}

type WebAuthnFinishRequest struct {
	TransactionID string `json:"transactionID,omitempty"`
	Response      string `json:"response,omitempty"`
}

type AuthFactor string

const AuthFactorUnknown AuthFactor = ""
const AuthFactorEmail AuthFactor = "email"
const AuthFactorPhone AuthFactor = "sms"
const AuthFactorSaml AuthFactor = "fed"
const AuthFactorOAuth AuthFactor = "oauth"
const AuthFactorWebauthn AuthFactor = "webauthn"
const AuthFactorTOTP AuthFactor = "totp"
const AuthFactorMFA AuthFactor = "mfa"

type Token struct {
	RefreshExpiration int64                  `json:"refreshExpiration,omitempty"`
	Expiration        int64                  `json:"expiration,omitempty"`
	JWT               string                 `json:"jwt,omitempty"`
	ID                string                 `json:"id,omitempty"`
	ProjectID         string                 `json:"projectId,omitempty"`
	Claims            map[string]interface{} `json:"claims,omitempty"`
}

func (to *Token) GetTenants() []string {
	tenants := to.getTenants()
	return maps.Keys(tenants)
}

func (to *Token) GetTenantValue(tenant, key string) any {
	tenants := to.getTenants()
	if info, ok := tenants[tenant].(map[string]any); ok {
		return info[key]
	}
	return nil
}

func (to *Token) getTenants() map[string]any {
	if to.Claims != nil {
		if tenants, ok := to.Claims[ClaimAuthorizedTenants].(map[string]any); ok {
			return tenants
		}
	}
	return make(map[string]any)
}

func (to *Token) CustomClaim(value string) interface{} {
	if to.Claims != nil {
		return to.Claims[value]
	}
	return nil
}

func (to *Token) AuthFactors() []AuthFactor {
	if to.Claims == nil {
		return nil
	}
	var afs []AuthFactor
	factors, ok := to.Claims["amr"]
	if ok {
		factorsArr, ok := factors.([]interface{})
		if ok {
			for i := range factorsArr {
				af, ok := factorsArr[i].(string)
				if ok {
					afs = append(afs, AuthFactor(af))
				} else {
					logger.LogInfo("Unkown authfactor type [%T]", factorsArr[i]) //notest
				}
			}
		} else {
			logger.LogInfo("Unknown amr value type [%T]", factors) //notest
		}
	}
	// cases of no factors are not interesting, so not going to log them
	return afs
}

func (to *Token) IsMFA() bool {
	return len(to.AuthFactors()) > 1
}

type LoginOptions struct {
	Stepup       bool                   `json:"stepup,omitempty"`
	MFA          bool                   `json:"mfa,omitempty"`
	CustomClaims map[string]interface{} `json:"customClaims,omitempty"`
}

func (lo *LoginOptions) IsJWTRequired() bool {
	return lo != nil && (lo.Stepup || lo.MFA)
}

type JWTResponse struct {
	SessionJwt       string        `json:"sessionJwt,omitempty"`
	RefreshJwt       string        `json:"refreshJwt,omitempty"`
	CookieDomain     string        `json:"cookieDomain,omitempty"`
	CookiePath       string        `json:"cookiePath,omitempty"`
	CookieMaxAge     int32         `json:"cookieMaxAge,omitempty"`
	CookieExpiration int32         `json:"cookieExpiration,omitempty"`
	User             *UserResponse `json:"user,omitempty"`
	FirstSeen        bool          `json:"firstSeen,omitempty"`
}

type MagicLinkResponse struct {
	PendingRef string `json:"pendingRef,omitempty"` // Pending referral code used to poll magic link authentication status
}

func NewAuthenticationInfo(jRes *JWTResponse, token *Token) *AuthenticationInfo {
	if jRes == nil {
		jRes = &JWTResponse{}
	}
	return &AuthenticationInfo{SessionToken: token, User: jRes.User, FirstSeen: jRes.FirstSeen}
}

func NewToken(JWT string, token jwt.Token) *Token {
	if token == nil {
		return nil
	}

	return &Token{
		JWT:        JWT,
		ID:         token.Subject(),
		ProjectID:  token.Issuer(),
		Expiration: token.Expiration().Unix(),
		Claims:     token.PrivateClaims(),
	}
}

type User struct {
	Name  string `json:"name,omitempty"`
	Phone string `json:"phone,omitempty"`
	Email string `json:"email,omitempty"`
}

type WebauthnUserRequest struct {
	ExternalID string `json:"externalId,omitempty"`
	Name       string `json:"name,omitempty"`
	Icon       string `json:"icon,omitempty"`
}

type UserResponse struct {
	User          `json:",inline"`
	UserID        string   `json:"userId,omitempty"`
	ExternalIDs   []string `json:"externalIds,omitempty"`
	VerifiedEmail bool     `json:"verifiedEmail,omitempty"`
	VerifiedPhone bool     `json:"verifiedPhone,omitempty"`
}

type authenticationRequestBody struct {
	ExternalID   string        `json:"externalId,omitempty"`
	LoginOptions *LoginOptions `json:"loginOptions,omitempty"`
}

type authenticationSignUpRequestBody struct {
	WhatsApp   string `json:"whatsapp,omitempty"`
	Phone      string `json:"phone,omitempty"`
	Email      string `json:"email,omitempty"`
	ExternalID string `json:"externalId,omitempty"`
	User       *User  `json:"user"`
}

type authenticationWebAuthnSignUpRequestBody struct {
	User   *WebauthnUserRequest `json:"user"`
	Origin string               `json:"origin"`
}

type authenticationWebAuthnSignInRequestBody struct {
	ExternalID   string        `json:"externalId,omitempty"`
	Origin       string        `json:"origin"`
	LoginOptions *LoginOptions `json:"loginOptions,omitempty"`
}

type authenticationWebAuthnSignUpOrInRequestBody struct {
	ExternalID string `json:"externalId,omitempty"`
	Origin     string `json:"origin"`
}

type authenticationWebAuthnAddDeviceRequestBody struct {
	ExternalID string `json:"externalId,omitempty"`
	Origin     string `json:"origin"`
}

type authenticationVerifyRequestBody struct {
	*authenticationRequestBody `json:",inline"`
	Code                       string `json:"code"`
}

type authenticationVerifyTOTPRequestBody struct {
	*authenticationVerifyRequestBody `json:",inline"`
	LoginOptions                     *LoginOptions `json:"loginOptions,omitempty"`
}

type totpSignUpRequestBody struct {
	ExternalID string `json:"externalId,omitempty"`
	User       *User  `json:"user,omitempty"`
}

type otpUpdateEmailRequestBody struct {
	ExternalID string `json:"externalId,omitempty"`
	Email      string `json:"email,omitempty"`
}

type otpUpdatePhoneRequestBody struct {
	ExternalID string `json:"externalId,omitempty"`
	Phone      string `json:"phone,omitempty"`
}

type magicLinkAuthenticationRequestBody struct {
	*authenticationRequestBody `json:",inline"`
	URI                        string        `json:"URI,omitempty"`
	CrossDevice                bool          `json:"crossDevice,omitempty"`
	LoginOptions               *LoginOptions `json:"loginOptions,omitempty"`
}

type magicLinkAuthenticationSignUpRequestBody struct {
	*authenticationSignUpRequestBody `json:",inline"`
	URI                              string `json:"URI,omitempty"`
	CrossDevice                      bool   `json:"crossDevice,omitempty"`
}

type magicLinkUpdateEmailRequestBody struct {
	Email       string `json:"email,inline"`
	ExternalID  string `json:"externalId,inline"`
	URI         string `json:"URI,omitempty"`
	CrossDevice bool   `json:"crossDevice,omitempty"`
}

type magicLinkUpdatePhoneRequestBody struct {
	Phone       string `json:"phone,inline"`
	ExternalID  string `json:"externalId,inline"`
	URI         string `json:"URI,omitempty"`
	CrossDevice bool   `json:"crossDevice,omitempty"`
}

type magicLinkAuthenticationVerifyRequestBody struct {
	Token string `json:"token"`
}

type authenticationGetMagicLinkSessionBody struct {
	PendingRef string `json:"pendingRef"`
}

type exchangeTokenBody struct {
	Code string `json:"code"`
}

func newSignInRequestBody(externalID string, loginOptions *LoginOptions) *authenticationRequestBody {
	return &authenticationRequestBody{ExternalID: externalID, LoginOptions: loginOptions}
}

func newSignUpRequestBody(method DeliveryMethod, user *User) *authenticationSignUpRequestBody {
	switch method {
	case MethodSMS:
		return &authenticationSignUpRequestBody{Phone: user.Phone}
	case MethodWhatsApp:
		return &authenticationSignUpRequestBody{WhatsApp: user.Phone}
	}

	return &authenticationSignUpRequestBody{Email: user.Email}
}

func newSignUPTOTPRequestBody(externalID string, user *User) *totpSignUpRequestBody {
	return &totpSignUpRequestBody{ExternalID: externalID, User: user}
}

func newOTPUpdateEmailRequestBody(externalID, email string) *otpUpdateEmailRequestBody {
	return &otpUpdateEmailRequestBody{ExternalID: externalID, Email: email}
}

func newOTPUpdatePhoneRequestBody(externalID, phone string) *otpUpdatePhoneRequestBody {
	return &otpUpdatePhoneRequestBody{ExternalID: externalID, Phone: phone}
}

func newMagicLinkAuthenticationRequestBody(value, URI string, crossDevice bool, loginOptions *LoginOptions) *magicLinkAuthenticationRequestBody {
	return &magicLinkAuthenticationRequestBody{authenticationRequestBody: newSignInRequestBody(value, loginOptions), URI: URI, CrossDevice: crossDevice, LoginOptions: loginOptions}
}

func newMagicLinkAuthenticationSignUpRequestBody(method DeliveryMethod, externalID, URI string, user *User, crossDevice bool) *magicLinkAuthenticationSignUpRequestBody {
	b := newSignUpRequestBody(method, user)
	b.User = user
	b.ExternalID = externalID
	return &magicLinkAuthenticationSignUpRequestBody{authenticationSignUpRequestBody: b, CrossDevice: crossDevice, URI: URI}
}

func newMagicLinkAuthenticationVerifyRequestBody(token string) *magicLinkAuthenticationVerifyRequestBody {
	return &magicLinkAuthenticationVerifyRequestBody{Token: token}
}

func newAuthenticationSignUpRequestBody(method DeliveryMethod, externalID string, user *User) *authenticationSignUpRequestBody {
	b := newSignUpRequestBody(method, user)
	b.User = user
	b.ExternalID = externalID
	return b
}

func newAuthenticationVerifyRequestBody(value string, code string) *authenticationVerifyRequestBody {
	return &authenticationVerifyRequestBody{authenticationRequestBody: newSignInRequestBody(value, nil), Code: code}
}

func newAuthenticationVerifyTOTPRequestBody(value string, code string, loginOptions *LoginOptions) *authenticationVerifyTOTPRequestBody {
	return &authenticationVerifyTOTPRequestBody{authenticationVerifyRequestBody: newAuthenticationVerifyRequestBody(value, code), LoginOptions: loginOptions}
}

func newMagicLinkUpdateEmailRequestBody(externalID, email string, URI string, crossDevice bool) *magicLinkUpdateEmailRequestBody {
	return &magicLinkUpdateEmailRequestBody{ExternalID: externalID, Email: email, URI: URI, CrossDevice: crossDevice}
}

func newMagicLinkUpdatePhoneRequestBody(externalID, phone string, URI string, crossDevice bool) *magicLinkUpdatePhoneRequestBody {
	return &magicLinkUpdatePhoneRequestBody{ExternalID: externalID, Phone: phone, URI: URI, CrossDevice: crossDevice}
}

func newAuthenticationGetMagicLinkSessionBody(pendingRef string) *authenticationGetMagicLinkSessionBody {
	return &authenticationGetMagicLinkSessionBody{PendingRef: pendingRef}
}

func newExchangeTokenBody(code string) *exchangeTokenBody {
	return &exchangeTokenBody{Code: code}
}

type DeliveryMethod string

type OAuthProvider string

type ContextKey string

const (
	MethodWhatsApp DeliveryMethod = "whatsapp"
	MethodSMS      DeliveryMethod = "sms"
	MethodEmail    DeliveryMethod = "email"

	OAuthFacebook  OAuthProvider = "facebook"
	OAuthGithub    OAuthProvider = "github"
	OAuthGoogle    OAuthProvider = "google"
	OAuthMicrosoft OAuthProvider = "microsoft"
	OAuthGitlab    OAuthProvider = "gitlab"
	OAuthApple     OAuthProvider = "apple"

	SessionCookieName = "DS"
	RefreshCookieName = "DSR"

	RedirectLocationCookieName = "Location"

	ContextUserIDProperty               = "DESCOPE_USER_ID"
	ContextUserIDPropertyKey ContextKey = ContextUserIDProperty
	ClaimAuthorizedTenants              = "tenants"

	claimAttributeName = "drn"
	claimPermissions   = "permissions"
	claimRoles         = "roles"
)

var (
	phoneRegex = regexp.MustCompile(`^(?:(?:\(?(?:00|\+)([1-4]\d\d|[1-9]\d?)\)?)?[\-\.\ \\\/]?)?((?:\(?\d{1,}\)?[\-\.\ \\\/]?){0,})(?:[\-\.\ \\\/]?(?:#|ext\.?|extension|x)[\-\.\ \\\/]?(\d+))?$`)
	emailRegex = regexp.MustCompile("^(((([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|((\\x22)((((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(([\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(\\([\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(\\x22)))@((([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$")
)
