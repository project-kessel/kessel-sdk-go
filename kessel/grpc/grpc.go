package grpc

import (
	"context"
	"fmt"
	"github.com/project-kessel/kessel-sdk-go/kessel/auth"
)
import "google.golang.org/grpc/credentials"

type callCredentials struct {
	credentials *auth.OAuth2ClientCredentials
}

func OAuth2CallCredentials(auth *auth.OAuth2ClientCredentials) credentials.PerRPCCredentials {
	return callCredentials{credentials: auth}
}

func (o callCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	token, err := o.credentials.GetToken(auth.GetTokenOptions{})
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"authorization": fmt.Sprintf("Bearer %s", token.AccessToken),
	}, nil
}

func (o callCredentials) RequireTransportSecurity() bool {
	return true
}
