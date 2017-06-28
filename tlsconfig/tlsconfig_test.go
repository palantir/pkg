// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlsconfig_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palantir/pkg/tlsconfig"
)

const (
	caCertFile     = "testdata/ca-cert.pem"
	serverCertFile = "testdata/server-cert.pem"
	serverKeyFile  = "testdata/server-key.pem"
	clientCertFile = "testdata/client-cert.pem"
	clientKeyFile  = "testdata/client-key.pem"
)

func TestUseTLSConfigForConnection(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(rw, "OK: %s", req.URL.Path)
	}))
	serverCfg, err := tlsconfig.NewServerConfig(
		tlsconfig.TLSCertFromFiles(serverCertFile, serverKeyFile),
		tlsconfig.ServerClientCAs(tlsconfig.CertPoolFromCAFiles(caCertFile)),
	)
	require.NoError(t, err)
	server.TLS = serverCfg
	server.StartTLS()
	defer server.Close()

	clientCfg, err := tlsconfig.NewClientConfig(
		tlsconfig.TLSCertFromFiles(clientCertFile, clientKeyFile),
		tlsconfig.ClientRootCAs(tlsconfig.CertPoolFromCAFiles(caCertFile)),
	)
	require.NoError(t, err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: clientCfg,
		},
	}

	resp, err := client.Get(server.URL + "/hello")
	require.NoError(t, err)
	bytes, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, "OK: /hello", string(bytes))
}
