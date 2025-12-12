// Copyright (c) 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlsconfig_test

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/palantir/pkg/tlsconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCertPoolFromCerts(t *testing.T) {
	// Load a certificate from file to use in the test
	certPEM, err := os.ReadFile(caCertFile)
	require.NoError(t, err)

	block, _ := pem.Decode(certPEM)
	require.NotNil(t, block, "failed to decode PEM block")

	cert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)

	provider := tlsconfig.CertPoolFromCerts(cert)
	pool, err := provider()
	require.NoError(t, err)
	assert.NotNil(t, pool)
}

func TestCertPoolFromCAFilesNonExistentFile(t *testing.T) {
	provider := tlsconfig.CertPoolFromCAFiles("testdata/nonexistent.pem")
	pool, err := provider()
	require.Error(t, err)
	assert.Nil(t, pool)
	assert.Contains(t, err.Error(), "failed to load certificates from file testdata/nonexistent.pem")
}
