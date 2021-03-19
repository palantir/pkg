// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tcpjson

import (
	"crypto/tls"
	"net"
	"net/url"
	"sync/atomic"
	"time"

	werror "github.com/palantir/witchcraft-go-error"
)

// ConnProvider defines the behavior to retrieve an established net.Conn.
type ConnProvider interface {
	// GetConn returns a net.Conn or an error if there is no connection established.
	// It is the caller's responsibility to close the returned net.Conn and
	// gracefully handle any closed connection errors if the net.Conn is shared across clients.
	GetConn() (net.Conn, error)
}

const (
	ErrNoURIs           = "no URIs provided to connect to"
	ErrFailedParsingURI = "failed to parse provided URI"
	ErrFailedDial       = "failed to dial the host:port"
)

var _ ConnProvider = (*tcpConnProvider)(nil)

// tcpConnProvider implements a ConnProvider that will round-robin TCP connections to a specific set of hosts.
type tcpConnProvider struct {
	// nextHostIdx contains the index of the next host to connect to from the hosts slice below.
	// The index will be reset to 0 when len(hosts) is reached to facilitate round-robin connections.
	nextHostIdx int32
	hosts       []string
	tlsConfig   *tls.Config
}

// NewTCPConnProvider returns a new ConnProvider that provides TCP connections.
// The provided uris must not be empty and must be able to be parsed as URIs. Refer to the documentation for url.Parse.
// If the provided TLS config is nil, then the default config will be used.
func NewTCPConnProvider(uris []string, tlsCfg *tls.Config) (ConnProvider, error) {
	if len(uris) < 1 {
		return nil, werror.Error(ErrNoURIs)
	}

	// Extract the host:port from each provided URI to simplify establishing a Dial connection
	var hosts []string
	for _, uri := range uris {
		u, err := url.Parse(uri)
		if err != nil {
			return nil, werror.Error(ErrFailedParsingURI, werror.SafeParam("uri", uri))
		}
		hosts = append(hosts, u.Host)
	}

	return &tcpConnProvider{
		nextHostIdx: 0,
		hosts:       hosts,
		tlsConfig:   tlsCfg,
	}, nil
}

var (
	defaultDialer = &net.Dialer{
		// Dial timeout is set to the http.DefaultTransport setting of 30 sec.
		Timeout: 30 * time.Second,
	}
)

func (s *tcpConnProvider) GetConn() (net.Conn, error) {
	hostIdx := atomic.LoadInt32(&s.nextHostIdx)
	nextHostIdx := int(hostIdx+1) % len(s.hosts)
	atomic.CompareAndSwapInt32(&s.nextHostIdx, hostIdx, int32(nextHostIdx))

	tlsConn, err := tls.DialWithDialer(defaultDialer, "tcp", s.hosts[hostIdx], s.tlsConfig)
	if err != nil {
		return nil, werror.Wrap(err, ErrFailedDial)
	}
	return tlsConn, nil
}
