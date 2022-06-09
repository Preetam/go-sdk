package auth

import (
	"net/http"
)

// Implementation in descope/auth/auth.go
type Authentication interface {
	// SignInOTP - Use to login a user based on the given identifier either email or a phone
	// and choose the selected delivery method for verification. (see auth/DeliveryMethod)
	// returns an error upon failure.
	SignInOTP(method DeliveryMethod, identifier string) error
	// SignUpOTP - Use to create a new user based on the given identifier either email or a phone.
	// choose the selected delivery method for verification. (see auth/DeliveryMethod)
	// optional to add user metadata for farther user details such as name and more.
	// returns an error upon failure.
	SignUpOTP(method DeliveryMethod, identifier string, user *User) error

	// VerifyCode - Use to verify a SignIn/SignUp based on the given identifier either an email or a phone
	// followed by the code used to verify and authenticate the user.
	// In case the request cookie can be renewed an automatic renewal is called and returns a new set of cookies to use.
	// Use the ResponseWriter (optional) to apply the cookies to the response automatically.
	// returns a list of cookies or an error upon failure.
	// This is a shortcut for VerifyCodeWithOptions(method, identifier, code, WithResponseOption(w))
	VerifyCode(method DeliveryMethod, identifier string, code string, w http.ResponseWriter) (*AuthenticationInfo, error)

	// VerifyCodeWithOptions - Use to verify a SignIn/SignUp based on the given identifier either an email or a phone
	// followed by the code used to verify and authenticate the user.
	// returns a list of cookies or an error upon failure.
	VerifyCodeWithOptions(method DeliveryMethod, identifier string, code string, options ...Option) (*AuthenticationInfo, error)

	// SignInMagicLink - Use to login a user based on a magic link that will be sent either email or a phone
	// and choose the selected delivery method for verification. (see auth/DeliveryMethod)
	// optional to add crossDevice - if true, request will return pending reference, that should be passed to GetPendingSession requests in order to get the authenticated session.
	// returns a pending reference if crossDevice is true, and an error upon failure.
	SignInMagicLink(method DeliveryMethod, identifier, URI string, crossDevice bool) error
	// SignUpMagicLink - Use to create a new user based on the given identifier either email or a phone.
	// choose the selected delivery method for verification. (see auth/DeliveryMethod)
	// optional to add user metadata for farther user details such as name and more.
	// optional to add crossDevice - if true, request will return pending reference, that should be passed to GetPendingSession requests in order to get the authenticated session.
	// returns a pending reference if crossDevice is true, and an error upon failure.
	SignUpMagicLink(method DeliveryMethod, identifier, URI string, user *User, crossDevice bool) error

	// VerifyMagicLink - Use to verify a SignInMagicLink/SignUpMagicLink request, based on the magic link token generated.
	// returns a list of cookies only if the magic link was created with crossDevice=false or an error upon failure.
	// This is a shortcut for VerifyMagicLinkWithOptions(method, code, WithResponseOption(w))
	VerifyMagicLink(code string, w http.ResponseWriter) (*AuthenticationInfo, error)
	// VerifyMagicLinkWithOptions - Use to verify a SignInMagicLink/SignUpMagicLink request, based on the magic link token generated.
	VerifyMagicLinkWithOptions(code string, options ...Option) (*AuthenticationInfo, error)

	// GetPendingSession - Use to get the pending session with the pending reference returned from a SignInMagicLink/SignUpMagicLink request with crossDevice=true.
	// returns a list of cookies or an error upon failure.
	// This is a shortcut for GetPendingSessionWithOptions(pendingRef, WithResponseOption(w))
	GetPendingSession(pendingRef string, w http.ResponseWriter) (*AuthenticationInfo, error)
	// GetPendingSessionWithOptions - Use to get the pending session with the pending reference returned from a SignInMagicLink/SignUpMagicLink request with crossDevice=true.
	// returns a list of cookies or an error upon failure.
	GetPendingSessionWithOptions(pendingRef string, options ...Option) (*AuthenticationInfo, error)

	// OAuthStart - use to start an OAuth authentication using the given OAuthProvider.
	// returns an error upon failure and a string represent the redirect URL upon success.
	// Uses the response writer to automatically redirect the client to the provider url for authentication.
	// A successful authentication will result in a callback to the url defined in the current project settings.
	// This is a shortcut for OAuthStartWithOptions(provider, WithResponseOption(w))
	OAuthStart(provider OAuthProvider, w http.ResponseWriter) (string, error)
	// OAuthStartWithOptions - use to start an OAuth authentication using the given OAuthProvider and options.
	OAuthStartWithOptions(provider OAuthProvider, options ...Option) (string, error)

	// ValidateSession - Use to validate a session of a given request.
	// Should be called before any private API call that requires authorization.
	// In case the request cookie can be renewed an automatic renewal is called and returns a new set of cookies to use.
	// Use the ResponseWriter (optional) to apply the cookies to the response automatically.
	// returns true upon success or false and an error upon failure.
	// This is a shortcut for ValidateSessionWithOptions(r, WithResponseOption(w))
	ValidateSession(request *http.Request, w http.ResponseWriter) (bool, *AuthenticationInfo, error)
	ValidateSessionWithOptions(request *http.Request, options ...Option) (bool, *AuthenticationInfo, error)

	// Logout - Use to perform logout from all active devices. This will revoke the given tokens
	// and if given options will also remove existing session on the given response sent to the client.
	// Use the ResponseWriter (optional) to apply the cookies to the response automatically.
	// This is a shortcut for LogoutWithOptions(r, WithResponseOption(w))
	Logout(request *http.Request, w http.ResponseWriter) error
	// LogoutWithOptions - Use to perform logout from all active devices. This will revoke the given tokens
	// and if given options will also remove existing session on the given response.
	LogoutWithOptions(request *http.Request, options ...Option) error
}
