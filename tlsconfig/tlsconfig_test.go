// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlsconfig_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palantir/pkg/tlsconfig"
)

const (
	cert1File         = "testdata/cert-1.pem"
	key1File          = "testdata/key-1.pem"
	cert2File         = "testdata/cert-2.pem"
	combinedCertsFile = "testdata/combined-certs.pem"
	certWithKeyFile   = "testdata/cert-with-key.pem"
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
				cert2File,
			},
		},
		{
			name: "cipherSuites specified",
			cipherSuites: []uint16{
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		},
	} {
		cfg, err := tlsconfig.NewClientConfig(cert1File, key1File, currCase.caFiles, currCase.cipherSuites)
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
			keyFile:   key1File,
			wantError: "^failed to load certificate from cert file  and key file .+/key-1.pem: open : no such file or directory$",
		},
		{
			name:      "missing key file",
			certFile:  cert1File,
			keyFile:   "",
			wantError: "^failed to load certificate from cert file .+/cert-1.pem and key file : open : no such file or directory$",
		},
		{
			name:     "invalid CA file",
			certFile: cert1File,
			keyFile:  key1File,
			caFiles: []string{
				key1File,
			},
			wantError: "^failed to load root CA certificates: no certificates detected in file .+/key-1.pem$",
		},
	} {
		cfg, err := tlsconfig.NewClientConfig(currCase.certFile, currCase.keyFile, currCase.caFiles, nil)
		require.Error(t, err, fmt.Sprintf("Case %d: %s", currCaseNum, currCase.name))
		assert.Regexp(t, regexp.MustCompile(currCase.wantError), err.Error(), "Case %d: %s", currCaseNum, currCase.name)
		assert.Nil(t, cfg, "Case %d: %s", currCaseNum, currCase.name)
	}
}

func TestNewServerConfig(t *testing.T) {
	for currCaseNum, currCase := range []struct {
		name          string
		clientCAFiles []string
		authType      tls.ClientAuthType
		cipherSuites  []uint16
	}{
		{
			name: "defaults",
		},
		{
			name: "caFiles specified",
			clientCAFiles: []string{
				cert1File,
			},
		},
		{
			name:     "authType specified",
			authType: tls.NoClientCert,
		},
		{
			name: "cipherSuites specified",
			cipherSuites: []uint16{
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		},
	} {
		cfg, err := tlsconfig.NewServerConfig(currCase.clientCAFiles, currCase.authType, currCase.cipherSuites)
		require.NoError(t, err)
		assert.NotNil(t, cfg, "Case %d: %s", currCaseNum, currCase.name)
	}
}

func TestNewServerConfigErrors(t *testing.T) {
	for currCaseNum, currCase := range []struct {
		name          string
		clientCAFiles []string
		wantError     string
	}{
		{
			name: "invalid CA file",
			clientCAFiles: []string{
				key1File,
			},
			wantError: "^failed to load client CA certificates: no certificates detected in file .+/key-1.pem$",
		},
	} {
		cfg, err := tlsconfig.NewServerConfig(currCase.clientCAFiles, tls.NoClientCert, nil)
		require.Error(t, err, fmt.Sprintf("Case %d: %s", currCaseNum, currCase.name))
		assert.Regexp(t, regexp.MustCompile(currCase.wantError), err.Error(), "Case %d: %s", currCaseNum, currCase.name)
		assert.Nil(t, cfg, "Case %d: %s", currCaseNum, currCase.name)
	}
}

func TestBuildCACertPool(t *testing.T) {
	for currCaseNum, currCase := range []struct {
		name         string
		inputFiles   []string
		wantNumCerts int
	}{
		{
			name:         "no files",
			wantNumCerts: 0,
		},
		{
			name: "file with single certificate",
			inputFiles: []string{
				cert1File,
			},
			wantNumCerts: 1,
		},
		{
			name: "multiple files with single certificate",
			inputFiles: []string{
				cert1File,
				cert2File,
			},
			wantNumCerts: 2,
		},
		{
			name: "single file with multiple certificates",
			inputFiles: []string{
				combinedCertsFile,
			},
			wantNumCerts: 2,
		},
		{
			name: "single file with certificate and key",
			inputFiles: []string{
				certWithKeyFile,
			},
			wantNumCerts: 1,
		},
	} {
		certPool, err := tlsconfig.BuildCACertPool(currCase.inputFiles...)
		require.NoError(t, err, "Case %d: %s", currCaseNum, currCase.name)
		assert.Equal(t, currCase.wantNumCerts, len(certPool.Subjects()), "Case %d: %s", currCaseNum, currCase.name)
	}
}

func TestBuildCACertPoolErrors(t *testing.T) {
	for currCaseNum, currCase := range []struct {
		name       string
		inputFiles []string
		wantError  string
	}{
		{
			name: "nonexistent file",
			inputFiles: []string{
				"nonexistent-file.txt",
			},
			wantError: `^failed to load certificates from file nonexistent-file.txt: open nonexistent-file.txt: no such file or directory$`,
		},
		{
			name: "file with no certificates",
			inputFiles: []string{
				key1File,
			},
			wantError: `^no certificates detected in file .+/key-1.pem$`,
		},
		{
			name: "multiple files where one has no certificates",
			inputFiles: []string{
				cert1File,
				key1File,
			},
			wantError: `^no certificates detected in file .+/key-1.pem$`,
		},
	} {
		_, err := tlsconfig.BuildCACertPool(currCase.inputFiles...)
		require.Error(t, err, fmt.Sprintf("Case %d: %s", currCaseNum, currCase.name))
		assert.Regexp(t, regexp.MustCompile(currCase.wantError), err.Error(), "Case %d: %s", currCaseNum, currCase.name)
	}
}

// newTestKeyPair creates a new self-signed key/certificate pair.
func newTestKeyPair(serial int64, org, dnsName string) (string, string) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	certTemplate := x509.Certificate{
		SerialNumber: big.NewInt(serial),
		Subject:      pkix.Name{Organization: []string{org}},
		DNSNames:     []string{dnsName},
	}

	certDERBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &privateKey.PublicKey, privateKey)
	if err != nil {
		panic(err)
	}

	var certBuf bytes.Buffer
	err = pem.Encode(&certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: certDERBytes})
	if err != nil {
		panic(err)
	}
	var keyBuf bytes.Buffer
	err = pem.Encode(&keyBuf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	if err != nil {
		panic(err)
	}
	return certBuf.String(), keyBuf.String()
}
