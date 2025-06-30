package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/config"
	"github.com/project-kessel/kessel-sdk-go/pkg/kessel/errors"

	"github.com/go-kratos/kratos/v2/transport/http"
)

type ClientBuilder struct {
	endpoint        string
	timeout         time.Duration
	userAgent       string
	tlsConfig       *tls.Config
	insecure        bool
	maxIdleConns    int
	idleConnTimeout time.Duration
	interceptors    []Interceptor
	headers         map[string]string
}

type Interceptor func(*http.Request) error

func NewClientBuilder(endpoint string) *ClientBuilder {
	return &ClientBuilder{
		endpoint:        endpoint,
		timeout:         10 * time.Second,
		userAgent:       "kessel-go-sdk",
		insecure:        false,
		maxIdleConns:    100,
		idleConnTimeout: 90 * time.Second,
		headers:         make(map[string]string),
	}
}

// NewClientBuilderFromConfig creates a ClientBuilder from HTTPConfig
func NewClientBuilderFromConfig(cfg *config.HTTPConfig) *ClientBuilder {
	return &ClientBuilder{
		endpoint:        cfg.Endpoint,
		timeout:         cfg.Timeout,
		userAgent:       cfg.UserAgent,
		tlsConfig:       cfg.TLSConfig,
		insecure:        cfg.Insecure,
		maxIdleConns:    cfg.MaxIdleConns,
		idleConnTimeout: cfg.IdleConnTimeout,
		interceptors:    []Interceptor{},
		headers:         make(map[string]string),
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
	b.insecure = false
	b.tlsConfig = tlsConfig
	return b
}

func (b *ClientBuilder) WithInsecure() *ClientBuilder {
	b.insecure = true
	b.tlsConfig = &tls.Config{InsecureSkipVerify: true}
	return b
}

func (b *ClientBuilder) WithMaxIdleConns(maxIdleConns int) *ClientBuilder {
	b.maxIdleConns = maxIdleConns
	return b
}

func (b *ClientBuilder) WithIdleConnTimeout(timeout time.Duration) *ClientBuilder {
	b.idleConnTimeout = timeout
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

	// Add TLS configuration if provided
	if b.tlsConfig != nil {
		opts = append(opts, http.WithTLSConfig(b.tlsConfig))
	}

	cli, err := http.NewClient(ctx, opts...)
	if err != nil {
		return nil, errors.NewHTTPClientError(err, "failed to create HTTP client")
	}

	return cli, nil
}
