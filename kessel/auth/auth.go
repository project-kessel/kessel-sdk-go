package auth

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/zitadel/oidc/v3/pkg/client"
)

const expirationWindow = 300  // 5 minutes in second
const defaultExpiresIn = 3600 // 1 hour in seconds

type OIDCDiscoveryMetadata struct {
	TokenEndpoint string
}

type RefreshTokenResponse struct {
	AccessToken string
	ExpiresAt   time.Time
}

type OAuth2ClientCredentials struct {
	clientId      string
	clientSecret  string
	tokenEndpoint string
	cachedToken   RefreshTokenResponse
	tokenMutex    sync.Mutex
}

type FetchOIDCDiscoveryOptions struct {
	// Optionally specify an http.Client or use http.DefaultClient
	HttpClient *http.Client
}

type GetTokenOptions struct {
	// Whether the token should be refreshed regardless if it is expired or not
	ForceRefresh bool
	// Optionally specify an http.Client or use http.DefaultClient
	HttpClient *http.Client
}

type oauth2TokenEndpointCaller struct {
	tokenEndpoint string
	httpClient    *http.Client
}

type requestToken struct {
	ClientID     string `schema:"client_id,omitempty"`
	ClientSecret string `schema:"client_secret,omitempty"`
	GrantType    string `schema:"grant_type"`
}

func NewOAuth2ClientCredentials(clientId string, clientSecret string, tokenEndpoint string) OAuth2ClientCredentials {
	return OAuth2ClientCredentials{
		clientId:      clientId,
		clientSecret:  clientSecret,
		tokenEndpoint: tokenEndpoint,
		cachedToken:   RefreshTokenResponse{},
		tokenMutex:    sync.Mutex{},
	}
}

func FetchOIDCDiscovery(ctx context.Context, issuerUrl string, options FetchOIDCDiscoveryOptions) (OIDCDiscoveryMetadata, error) {
	httpClient := options.HttpClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	discoveryConfig, err := client.Discover(ctx, issuerUrl, httpClient)
	if err != nil {
		return OIDCDiscoveryMetadata{}, err
	}

	return OIDCDiscoveryMetadata{TokenEndpoint: discoveryConfig.TokenEndpoint}, nil
}

func (o *OAuth2ClientCredentials) GetToken(ctx context.Context, options GetTokenOptions) (RefreshTokenResponse, error) {
	httpClient := options.HttpClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	if !options.ForceRefresh && o.isTokenValid() {
		return o.cachedToken, nil
	}

	o.tokenMutex.Lock()
	defer o.tokenMutex.Unlock()

	if options.ForceRefresh {
		o.cachedToken = RefreshTokenResponse{}
	}

	var err error
	o.cachedToken, err = o.refreshToken(ctx, httpClient)

	return o.cachedToken, err
}

func (o *OAuth2ClientCredentials) refreshToken(ctx context.Context, httpClient *http.Client) (RefreshTokenResponse, error) {
	request := requestToken{
		ClientID:     o.clientId,
		ClientSecret: o.clientSecret,
		GrantType:    "client_credentials",
	}

	tokenEndpointCaller := oauth2TokenEndpointCaller{
		tokenEndpoint: o.tokenEndpoint,
		httpClient:    httpClient,
	}

	token, err := client.CallTokenEndpoint(ctx, request, tokenEndpointCaller)

	if err != nil {
		return RefreshTokenResponse{}, err
	}

	expiresIn := token.ExpiresIn
	if expiresIn == 0 {
		expiresIn = defaultExpiresIn
	}

	return RefreshTokenResponse{
		AccessToken: token.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(expiresIn) * time.Second),
	}, nil
}

func (o *OAuth2ClientCredentials) isTokenValid() bool {
	if o.cachedToken.AccessToken == "" {
		return false
	}

	return time.Now().Add(time.Duration(expirationWindow) * time.Second).Before(o.cachedToken.ExpiresAt)
}

func (o oauth2TokenEndpointCaller) TokenEndpoint() string {
	return o.tokenEndpoint
}

func (o oauth2TokenEndpointCaller) HttpClient() *http.Client {
	return o.httpClient
}
