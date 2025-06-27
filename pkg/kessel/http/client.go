package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/errors"

	"github.com/go-kratos/kratos/v2/transport/http"
)

type ClientBuilder struct {
	endpoint     string
	timeout      time.Duration
	userAgent    string
	tlsConfig    *tls.Config
	interceptors []Interceptor
	headers      map[string]string
}

type Interceptor func(*http.Request) error

func NewClientBuilder(endpoint string) *ClientBuilder {
	return &ClientBuilder{
		endpoint:  endpoint,
		timeout:   10 * time.Second,
		userAgent: "kessel-go-sdk",
		headers:   make(map[string]string),
	}
}

func (b *ClientBuilder) WithTimeout(timeout time.Duration) *ClientBuilder {
	b.timeout = timeout
	return b
}

func (b *ClientBuilder) WithUserAgent(userAgent string) *ClientBuilder {
	b.userAgent = userAgent
	return b
}

func (b *ClientBuilder) WithTLSConfig(tlsConfig *tls.Config) *ClientBuilder {
	b.tlsConfig = tlsConfig
	return b
}

func (b *ClientBuilder) WithInsecure() *ClientBuilder {
	b.tlsConfig = &tls.Config{InsecureSkipVerify: true}
	return b
}

func (b *ClientBuilder) WithHeader(key, value string) *ClientBuilder {
	b.headers[key] = value
	return b
}

func (b *ClientBuilder) WithBearerToken(token string) *ClientBuilder {
	b.headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	return b
}

func (b *ClientBuilder) WithInterceptor(interceptor Interceptor) *ClientBuilder {
	b.interceptors = append(b.interceptors, interceptor)
	return b
}

func (b *ClientBuilder) Build(ctx context.Context) (*http.Client, error) {
	opts := []http.ClientOption{
		http.WithEndpoint(b.endpoint),
		http.WithTimeout(b.timeout),
	}

	cli, err := http.NewClient(ctx, opts...)
	if err != nil {
		return nil, errors.NewHTTPClientError(err, "failed to create HTTP client")
	}

	return cli, nil
}
