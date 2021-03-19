// Copyright (c) 2021 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tcpjson_test

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/palantir/pkg/tcpjson"
	"github.com/stretchr/testify/require"
)

func TestNewTCPConnProvider(t *testing.T) {
	for _, tc := range []struct {
		name        string
		uris        []string
		tlsCfg      *tls.Config
		expectedErr string
	}{
		{
			name:        "no uris or TLS config",
			expectedErr: tcpjson.ErrNoURIs,
		},
		{
			name: "uris with no TLS config",
			uris: []string{"tcp://127.0.0.1:8765"},
		},
		{
			name:        "no scheme on URIs",
			uris:        []string{"127.0.0.1:8765"},
			expectedErr: tcpjson.ErrFailedParsingURI,
		},
		{
			name: "valid URIs",
			uris: []string{
				"tcp://127.0.0.1:8765",
				"tcp://127.0.0.2:8765",
			},
		},
		{
			name: "valid URIs with TLS config",
			uris: []string{
				"tcp://127.0.0.1:8765",
				"tcp://127.0.0.2:8765",
			},
			tlsCfg: &tls.Config{InsecureSkipVerify: true},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			connProvider, err := tcpjson.NewTCPConnProvider(tc.uris, tc.tlsCfg)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, connProvider)
			}
		})
	}
}

// TestGetConn verifies that GetConn can successfully dial a TCP server and a valid net.Conn is returned for use.
func TestGetConn(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.TLS = &tls.Config{InsecureSkipVerify: true}
	server.StartTLS()
	defer server.Close()

	provider, err := tcpjson.NewTCPConnProvider([]string{server.URL}, &tls.Config{InsecureSkipVerify: true})
	require.NoError(t, err)

	conn, err := provider.GetConn()
	require.NoError(t, err)
	err = conn.Close()
	require.NoError(t, err)
}
