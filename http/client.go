// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/http2"
)

func NewHTTPClient(timeout time.Duration, tlsConf *tls.Config) *http.Client {
	return &http.Client{
		Transport: newTransport(timeout, tlsConf),
		Timeout:   timeout,
	}
}

func NewHTTP2Client(timeout time.Duration, tlsConf *tls.Config) (*http.Client, error) {
	tr := newTransport(timeout, tlsConf)
	if err := http2.ConfigureTransport(tr); err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}, nil
}

func newTransport(timeout time.Duration, tlsConf *tls.Config) *http.Transport {
	return &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		TLSClientConfig:     tlsConf,
		MaxIdleConnsPerHost: 32,
		MaxIdleConns:        32,
		IdleConnTimeout:     timeout,
		TLSHandshakeTimeout: timeout,
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: timeout,
		}).DialContext,
	}
}
