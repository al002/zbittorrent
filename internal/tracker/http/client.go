package httpTracker

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
)

type Client struct {
	hc   *http.Client
	url_ *url.URL
}

type (
	ProxyFunc       func(*http.Request) (*url.URL, error)
	DialContextFunc func(ctx context.Context, network, addr string) (net.Conn, error)
)

type NewClientOpts struct {
	Proxy             ProxyFunc
	DialContext       DialContextFunc
	ServerName        string
	DisableKeepAlives bool
}

func NewClient(url_ *url.URL, opts NewClientOpts) Client {
	return Client{
		url_: url_,
		hc: &http.Client{
			Transport: &http.Transport{
				DialContext: opts.DialContext,
				Proxy:       opts.Proxy,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
					ServerName:         opts.ServerName,
				},
				DisableKeepAlives: opts.DisableKeepAlives,
			},
		},
	}
}

func (cl Client) Close() error {
	cl.hc.CloseIdleConnections()
  return nil
}
