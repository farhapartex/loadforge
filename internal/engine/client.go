package engine

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/farhapartex/loadforge/internal/config"
	"golang.org/x/net/http2"
)

const (
	defaultTimeout      = 30 * time.Second
	defaultMaxIdleConns = 100
	defaultIdleConnTime = 90 * time.Second
)

func buildClient(opts *config.RequestOptions) *http.Client {
	timeout := defaultTimeout
	followRedirects := true
	tlsSkipVerify := false
	enableHTTP2 := true

	if opts != nil {
		if opts.Timeout != "" {
			if d, err := time.ParseDuration(opts.Timeout); err == nil {
				timeout = d
			}
		}

		if opts.FollowRedirects != nil {
			followRedirects = *opts.FollowRedirects
		}

		if opts.TLSSkipVerify {
			tlsSkipVerify = true
		}

		if opts.HTTP2 != nil {
			enableHTTP2 = *opts.HTTP2
		}
	}

	tlsCgf := &tls.Config{
		InsecureSkipVerify: tlsSkipVerify,
	}

	transport := &http.Transport{
		TLSClientConfig:    tlsCgf,
		MaxIdleConns:       defaultMaxIdleConns,
		IdleConnTimeout:    defaultIdleConnTime,
		DisableCompression: false,
		DisableKeepAlives:  false,
	}

	if enableHTTP2 {
		if err := http2.ConfigureTransport(transport); err != nil {
			_ = err
		}
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	// handle redirect policy
	if !followRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client
}
