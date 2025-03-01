package auth

import (
	"net/http"

	"github.com/descope/go-sdk/descope/api"
	"github.com/descope/go-sdk/descope/errors"
	"github.com/descope/go-sdk/descope/logger"
	"github.com/descope/go-sdk/descope/utils"
)

type oauth struct {
	authenticationsBase
}

type oauthStartResponse struct {
	URL string `json:"url"`
}

func (auth *oauth) Start(provider OAuthProvider, redirectURL string, r *http.Request, loginOptions *LoginOptions, w http.ResponseWriter) (url string, err error) {
	m := map[string]string{
		"provider": string(provider),
	}
	if len(redirectURL) > 0 {
		m["redirectURL"] = redirectURL
	}
	var pswd string
	if loginOptions.IsJWTRequired() {
		pswd, err = getValidRefreshToken(r)
		if err != nil {
			return "", errors.InvalidStepupJwtError
		}
	}

	httpResponse, err := auth.client.DoPostRequest(composeOAuthURL(), loginOptions, &api.HTTPRequest{QueryParams: m}, pswd)
	if err != nil {
		return
	}

	if httpResponse.Res != nil {
		res := &oauthStartResponse{}
		err = utils.Unmarshal([]byte(httpResponse.BodyStr), res)
		if err != nil {
			logger.LogError("failed to parse location from response for [%s]", err, provider)
			return "", err
		}
		url = res.URL
		redirectToURL(url, w)
	}

	return
}

func (auth *oauth) ExchangeToken(code string, w http.ResponseWriter) (*AuthenticationInfo, error) {
	return auth.exchangeToken(code, composeOAuthExchangeTokenURL(), w)
}
