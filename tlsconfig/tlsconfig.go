// Copyright 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

var defaultCipherSuites = []uint16{
	// This cipher suite is included to enable http/2. For details, see
	// https://blog.bracelab.com/achieving-perfect-ssl-labs-score-with-go
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_RSA_WITH_AES_256_CBC_SHA,
}

// NewClientConfig returns a tls.Config that is suitable to use for a client connection in 2-way TLS connections. The
// configuration uses the provided key/value pair as the certificate it presents and uses the provided caFiles as the
// set of root certificate authorities used to validate certificates. If caFiles is nil or empty, the host's root CA set
// is used. If cipherSuites is nil or empty, defaultCipherSuites is used.
func NewClientConfig(certFile, keyFile string, caFiles []string, cipherSuites []uint16) (*tls.Config, error) {
	return create(cipherSuites,
		withAuthKeyPair(certFile, keyFile),
		withRootCAs(caFiles),
	)
}

// NewServerConfig returns a tls.Config that is suitable to use for a server in 2-way TLS connections. The configuration
// uses the provided clientCAFiles as the set of certificate authorities used to validate client certificates and
// authType as the policy the server follows for client TLS authentication. If cipherSuites is nil or empty,
// defaultCipherSuites is used. The configuration does not store the key and certificate used by the server.
func NewServerConfig(clientCAFiles []string, authType tls.ClientAuthType, cipherSuites []uint16) (*tls.Config, error) {
	return create(cipherSuites,
		withClientConfig(clientCAFiles, authType),
	)
}

// BuildCACertPool returns a certificate pool that contains the certificates read from the provided files. Returns an
// error if any of the provided files cannot be opened or if any of the files contain no certificates.
func BuildCACertPool(caFiles ...string) (*x509.CertPool, error) {
	caCertPool := x509.NewCertPool()
	for _, caFile := range caFiles {
		cert, err := ioutil.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load certificates from file %s: %v", caFile, err)
		}
		if ok := caCertPool.AppendCertsFromPEM(cert); !ok {
			return nil, fmt.Errorf("no certificates detected in file %s", caFile)
		}
	}
	return caCertPool, nil
}

type configurer func(*tls.Config) error

func withAuthKeyPair(certFile, keyFile string) configurer {
	return func(tlsConfig *tls.Config) error {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return fmt.Errorf("failed to load certificate from cert file %s and key file %s: %v", certFile, keyFile, err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		return nil
	}
}

func withClientConfig(clientCAFiles []string, authType tls.ClientAuthType) configurer {
	return func(tlsConfig *tls.Config) error {
		caCertPool, err := BuildCACertPool(clientCAFiles...)
		if err != nil {
			return fmt.Errorf("failed to load client CA certificates: %v", err)
		}
		tlsConfig.ClientCAs = caCertPool
		tlsConfig.ClientAuth = authType
		return nil
	}
}

func withRootCAs(caFiles []string) configurer {
	return func(tlsConfig *tls.Config) error {
		if len(caFiles) == 0 {
			return nil
		}
		caCertPool, err := BuildCACertPool(caFiles...)
		if err != nil {
			return fmt.Errorf("failed to load root CA certificates: %v", err)
		}
		tlsConfig.RootCAs = caCertPool
		return nil
	}
}

func create(cipherSuites []uint16, configs ...configurer) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		CipherSuites:             defaultCipherSuites,
	}
	if len(cipherSuites) != 0 {
		tlsConfig.CipherSuites = cipherSuites
	}

	for _, currCfg := range configs {
		if currCfg == nil {
			continue
		}
		if err := currCfg(tlsConfig); err != nil {
			return nil, err
		}
	}

	tlsConfig.BuildNameToCertificate()
	return tlsConfig, nil
}
