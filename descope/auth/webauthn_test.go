package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/descope/go-sdk/descope/api"
	"github.com/descope/go-sdk/descope/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignUpStart(t *testing.T) {
	expectedResponse := WebAuthnTransactionResponse{TransactionID: "a"}
	a, err := newTestAuth(nil, DoOkWithBody(func(r *http.Request) {
		req := authenticationWebAuthnSignUpRequestBody{}
		err := readBody(r, &req)
		require.NoError(t, err)
		assert.EqualValues(t, "test@test.com", req.User.ExternalID)
		assert.EqualValues(t, "test2@test.com", req.User.Name)
	}, expectedResponse))
	require.NoError(t, err)
	res, err := a.WebAuthn().SignUpStart("test@test.com", &User{Name: "test2@test.com"}, "https://example.com")
	require.NoError(t, err)
	assert.EqualValues(t, expectedResponse.TransactionID, res.TransactionID)
}

func TestSignUpStartNilUser(t *testing.T) {
	expectedResponse := WebAuthnTransactionResponse{TransactionID: "a"}
	a, err := newTestAuth(nil, DoOkWithBody(func(r *http.Request) {
		req := authenticationWebAuthnSignUpRequestBody{}
		err := readBody(r, &req)
		require.NoError(t, err)
		assert.EqualValues(t, "test@test.com", req.User.ExternalID)
	}, expectedResponse))
	require.NoError(t, err)
	res, err := a.WebAuthn().SignUpStart("test@test.com", nil, "https://example.com")
	require.NoError(t, err)
	assert.EqualValues(t, expectedResponse.TransactionID, res.TransactionID)
}

func TestSignInFinish(t *testing.T) {
	expectedResponse := &WebAuthnFinishRequest{TransactionID: "a"}
	a, err := newTestAuth(nil, DoOk(nil))
	require.NoError(t, err)
	w := httptest.NewRecorder()
	res, err := a.WebAuthn().SignInFinish(expectedResponse, w)
	require.NoError(t, err)
	assert.EqualValues(t, jwtTokenValid, res.SessionToken.JWT)
	require.Len(t, w.Result().Cookies(), 1)
	assert.EqualValues(t, jwtTokenValid, w.Result().Cookies()[0].Value)
}

func TestSignInStart(t *testing.T) {
	expectedResponse := WebAuthnTransactionResponse{TransactionID: "a"}
	a, err := newTestAuth(nil, DoOkWithBody(func(r *http.Request) {
		req := authenticationRequestBody{}
		err := readBody(r, &req)
		require.NoError(t, err)
		assert.EqualValues(t, "a", req.ExternalID)
	}, expectedResponse))
	require.NoError(t, err)
	res, err := a.WebAuthn().SignInStart("a", "https://example.com", nil, nil)
	require.NoError(t, err)
	assert.EqualValues(t, expectedResponse.TransactionID, res.TransactionID)
}

func TestSignInStartStepup(t *testing.T) {
	expectedResponse := WebAuthnTransactionResponse{TransactionID: "a"}
	a, err := newTestAuth(nil, DoOkWithBody(func(r *http.Request) {
		req := authenticationRequestBody{}
		err := readBody(r, &req)
		require.NoError(t, err)
		assert.EqualValues(t, "a", req.ExternalID)
		body, err := readBodyMap(r)
		require.NoError(t, err)
		assert.EqualValues(t, map[string]interface{}{"stepup": true, "customClaims": map[string]interface{}{"k1": "v1"}}, body["loginOptions"])
		reqToken := r.Header.Get(api.AuthorizationHeaderName)
		splitToken := strings.Split(reqToken, api.BearerAuthorizationPrefix)
		require.Len(t, splitToken, 2)
		bearer := splitToken[1]
		bearers := strings.Split(bearer, ":")
		require.Len(t, bearers, 2)
		assert.EqualValues(t, "test", bearers[1])

	}, expectedResponse))
	require.NoError(t, err)
	res, err := a.WebAuthn().SignInStart("a", "https://example.com", &http.Request{Header: http.Header{"Cookie": []string{"DSR=test"}}}, &LoginOptions{Stepup: true, CustomClaims: map[string]interface{}{"k1": "v1"}})
	require.NoError(t, err)
	assert.EqualValues(t, expectedResponse.TransactionID, res.TransactionID)
}

func TestSignInWebAuthnStartEmpty(t *testing.T) {
	a, err := newTestAuth(nil, nil)
	require.NoError(t, err)
	res, err := a.WebAuthn().SignInStart("", "https://example.com", nil, nil)
	require.Error(t, err)
	assert.Empty(t, res)
	assert.EqualValues(t, errors.BadRequestErrorCode, err.(*errors.WebError).Code)
}

func TestSignInWebAuthnStepupNoJWT(t *testing.T) {
	a, err := newTestAuth(nil, nil)
	require.NoError(t, err)
	res, err := a.WebAuthn().SignInStart("a", "https://example.com", nil, &LoginOptions{Stepup: true})
	require.Error(t, err)
	assert.Empty(t, res)
	assert.ErrorIs(t, err, errors.InvalidStepupJwtError)
}

func TestSignUpFinish(t *testing.T) {
	expectedResponse := &WebAuthnFinishRequest{TransactionID: "a"}
	a, err := newTestAuth(nil, DoOk(nil))
	require.NoError(t, err)
	w := httptest.NewRecorder()
	res, err := a.WebAuthn().SignUpFinish(expectedResponse, w)
	require.NoError(t, err)
	assert.EqualValues(t, jwtTokenValid, res.SessionToken.JWT)
	require.Len(t, w.Result().Cookies(), 1)
	assert.EqualValues(t, jwtTokenValid, w.Result().Cookies()[0].Value)
}

func TestSignUpOrInStart(t *testing.T) {
	expectedResponse := WebAuthnTransactionResponse{TransactionID: "a", Create: true}
	a, err := newTestAuth(nil, DoOkWithBody(func(r *http.Request) {
		req := authenticationWebAuthnSignUpOrInRequestBody{}
		err := readBody(r, &req)
		require.NoError(t, err)
		assert.EqualValues(t, "test@test.com", req.ExternalID)
		assert.EqualValues(t, "https://example.com", req.Origin)
	}, expectedResponse))
	require.NoError(t, err)
	res, err := a.WebAuthn().SignUpOrInStart("test@test.com", "https://example.com")
	require.NoError(t, err)
	assert.EqualValues(t, expectedResponse.TransactionID, res.TransactionID)
	assert.EqualValues(t, expectedResponse.Create, res.Create)
}

func TestSignUpOrInStartEmpty(t *testing.T) {
	a, err := newTestAuth(nil, nil)
	require.NoError(t, err)
	res, err := a.WebAuthn().SignUpOrInStart("", "https://example.com")
	require.Error(t, err)
	assert.Nil(t, res)
	assert.EqualValues(t, errors.BadRequestErrorCode, err.(*errors.WebError).Code)
}

func TestWebAuthnUpdateUserDeviceStart(t *testing.T) {
	expectedResponse := WebAuthnTransactionResponse{TransactionID: "a"}
	a, err := newTestAuth(nil, DoOkWithBody(func(r *http.Request) {
		req := authenticationRequestBody{}
		err := readBody(r, &req)
		require.NoError(t, err)
		assert.EqualValues(t, "test@test.com", req.ExternalID)
	}, expectedResponse))
	require.NoError(t, err)
	r := &http.Request{Header: http.Header{}}
	r.AddCookie(&http.Cookie{Name: RefreshCookieName, Value: jwtTokenValid})
	res, err := a.WebAuthn().UpdateUserDeviceStart("test@test.com", "https://example.com", r)
	require.NoError(t, err)
	assert.EqualValues(t, expectedResponse.TransactionID, res.TransactionID)
}

func TestWebAuthnUpdateUserDeviceStartEmpty(t *testing.T) {
	a, err := newTestAuth(nil, nil)
	require.NoError(t, err)
	r := &http.Request{Header: http.Header{}}
	r.AddCookie(&http.Cookie{Name: RefreshCookieName, Value: jwtTokenValid})
	res, err := a.WebAuthn().UpdateUserDeviceStart("", "https://example.com", r)
	require.Error(t, err)
	assert.Empty(t, res)
	assert.EqualValues(t, errors.BadRequestErrorCode, err.(*errors.WebError).Code)
}

func TestWebAuthnUpdateUserDeviceFinish(t *testing.T) {
	expectedResponse := &WebAuthnFinishRequest{TransactionID: "a"}
	a, err := newTestAuth(nil, DoOk(nil))
	require.NoError(t, err)
	err = a.WebAuthn().UpdateUserDeviceFinish(expectedResponse)
	require.NoError(t, err)
}
