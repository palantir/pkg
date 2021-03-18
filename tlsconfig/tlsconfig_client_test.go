// Copyright (c) 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlsconfig_test

import (
	"crypto/tls"
	"fmt"
	"testing"

	"github.com/palantir/pkg/tlsconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClientConfig(t *testing.T) {
	for currCaseNum, currCase := range []struct {
		name         string
		caFiles      []string
		cipherSuites []uint16
		certProvider tlsconfig.TLSCertProvider
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
		{
			name: "certProvider specified",
			certProvider: func() (tls.Certificate, error) {
				return tls.Certificate{}, nil
			},
		},
	} {
		cfg, err := tlsconfig.NewClientConfig(
			tlsconfig.ClientRootCAs(tlsconfig.CertPoolFromCAFiles(currCase.caFiles...)),
			tlsconfig.ClientCipherSuites(currCase.cipherSuites...),
			tlsconfig.ClientKeyPair(currCase.certProvider),
		)
		require.NoError(t, err)
		assert.NotNil(t, cfg, "Case %d: %s", currCaseNum, currCase.name)
	}
}

func TestNewClientConfigInsecureSkipVerify(t *testing.T) {
	cfg, err := tlsconfig.NewClientConfig(tlsconfig.ClientInsecureSkipVerify())
	assert.NoError(t, err)
	assert.True(t, cfg.InsecureSkipVerify)
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
			keyFile:   clientKeyFile,
			wantError: "failed to load TLS certificate: open : no such file or directory",
		},
		{
			name:      "missing key file",
			certFile:  clientCertFile,
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
			tlsconfig.ClientKeyPairFiles(currCase.certFile, currCase.keyFile),
			tlsconfig.ClientRootCAs(tlsconfig.CertPoolFromCAFiles(currCase.caFiles...)),
		)
		require.Error(t, err, fmt.Sprintf("Case %d: %s", currCaseNum, currCase.name))
		assert.EqualError(t, err, currCase.wantError, "Case %d: %s", currCaseNum, currCase.name)
		assert.Nil(t, cfg, "Case %d: %s", currCaseNum, currCase.name)
	}
}
