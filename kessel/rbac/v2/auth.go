package v2

import (
	"context"
	"github.com/project-kessel/kessel-sdk-go/kessel/auth"
	"net/http"
)

type AuthRequest interface {
	ConfigureRequest(ctx context.Context, request *http.Request) error
}

type OAuth2AuthRequestOptions struct {
	HttpClient *http.Client
}

type oauth2Auth struct {
	credentials *auth.OAuth2ClientCredentials
	httpClient  *http.Client
}

func OAuth2AuthRequest(credentials *auth.OAuth2ClientCredentials, options OAuth2AuthRequestOptions) AuthRequest {
	return oauth2Auth{credentials: credentials, httpClient: options.HttpClient}
}

func (o oauth2Auth) ConfigureRequest(ctx context.Context, request *http.Request) error {
	token, err := o.credentials.GetToken(ctx, auth.GetTokenOptions{
		HttpClient: o.httpClient,
	})

	if err != nil {
		return err
	}

	request.Header.Set("authorization", "Bearer "+token.AccessToken)
	return nil
}
