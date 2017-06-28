// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlsconfig_test

import (
	"crypto/tls"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palantir/pkg/tlsconfig"
)

func TestNewClientConfig(t *testing.T) {
	for currCaseNum, currCase := range []struct {
		name         string
		caFiles      []string
		cipherSuites []uint16
	}{
		{
			name: "defaults",
		},
		{
			name: "caFiles specified",
			caFiles: []string{
				caCertFile,
			},
		},
		{
			name: "cipherSuites specified",
			cipherSuites: []uint16{
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		},
	} {
		cfg, err := tlsconfig.NewClientConfig(
			tlsconfig.TLSCertFromFiles(clientCertFile, clientKeyFile),
			tlsconfig.ClientRootCAs(tlsconfig.CertPoolFromCAFiles(currCase.caFiles...)),
			tlsconfig.ClientCipherSuites(currCase.cipherSuites...),
		)
		require.NoError(t, err)
		assert.NotNil(t, cfg, "Case %d: %s", currCaseNum, currCase.name)
	}
}

func TestNewClientConfigErrors(t *testing.T) {
	for currCaseNum, currCase := range []struct {
		name      string
		certFile  string
		keyFile   string
		caFiles   []string
		wantError string
	}{
		{
			name:      "missing certificate file",
			certFile:  "",
			keyFile:   clientKeyFile,
			wantError: "failed to load TLS certificate: open : no such file or directory",
		},
		{
			name:      "missing key file",
			certFile:  clientCertFile,
			keyFile:   "",
			wantError: "failed to load TLS certificate: open : no such file or directory",
		},
		{
			name:     "invalid CA file",
			certFile: clientCertFile,
			keyFile:  clientKeyFile,
			caFiles: []string{
				serverKeyFile,
			},
			wantError: "failed to create certificate pool: no certificates detected in file testdata/server-key.pem",
		},
	} {
		cfg, err := tlsconfig.NewClientConfig(
			tlsconfig.TLSCertFromFiles(currCase.certFile, currCase.keyFile),
			tlsconfig.ClientRootCAs(tlsconfig.CertPoolFromCAFiles(currCase.caFiles...)),
		)
		require.Error(t, err, fmt.Sprintf("Case %d: %s", currCaseNum, currCase.name))
		assert.EqualError(t, err, currCase.wantError, "Case %d: %s", currCaseNum, currCase.name)
		assert.Nil(t, cfg, "Case %d: %s", currCaseNum, currCase.name)
	}
}
