// Copyright (c) 2016 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

var defaultCipherSuites = []uint16{
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
	// TLSv1.3
	tls.TLS_CHACHA20_POLY1305_SHA256,
	tls.TLS_AES_128_GCM_SHA256,
	tls.TLS_AES_256_GCM_SHA384,
	// TLSv1.2
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	// This cipher suite is included to enable http/2. For details, see
	// required to include for http/2: https://http2.github.io/http2-spec/#rfc.section.9.2.2
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
}

// TLSCertFromFiles returns a provider that returns a tls.Certificate by loading an X509 key pair from the files in the
// specified locations.
func TLSCertFromFiles(certFile, keyFile string) TLSCertProvider {
	return func() (tls.Certificate, error) {
		return tls.LoadX509KeyPair(certFile, keyFile)
	}
}

// CertPoolProvider is a type alias for providing a *x509.CertPool
type CertPoolProvider func() (*x509.CertPool, error)

// CertPoolOption is a type alias for a function that modifies a *x509.CertPool in-place.
type CertPoolOption func(certPool *x509.CertPool) error

func CertPoolFromCertPoolOptions(certPoolOptions []CertPoolOption) CertPoolProvider {
	return func() (*x509.CertPool, error) {
		certPool := x509.NewCertPool()
		for _, certPoolOption := range certPoolOptions {
			err := certPoolOption(certPool)
			if err != nil {
				return nil, err
			}
		}
		return certPool, nil
	}
}

func CertPoolFromCAFiles(caFiles ...string) CertPoolProvider {
	return CertPoolFromCertPoolOptions([]CertPoolOption{
		CertPoolOptionFromCAFiles(caFiles...),
	})
}

func CertPoolOptionFromCAFiles(caFiles ...string) CertPoolOption {
	return func(certPool *x509.CertPool) error {
		for _, caFile := range caFiles {
			cert, err := os.ReadFile(caFile)
			if err != nil {
				return fmt.Errorf("failed to load certificates from file %s: %v", caFile, err)
			}
			if ok := certPool.AppendCertsFromPEM(cert); !ok {
				return fmt.Errorf("no certificates detected in file %s", caFile)
			}
		}
		return nil
	}
}

func CertPoolFromCABytes(cert []byte) CertPoolProvider {
	return CertPoolFromCertPoolOptions([]CertPoolOption{
		CertPoolOptionCABytes(cert),
	})
}

func CertPoolOptionCABytes(cert []byte) CertPoolOption {
	return func(certPool *x509.CertPool) error {
		if ok := certPool.AppendCertsFromPEM(cert); !ok {
			return fmt.Errorf("no certificates detected in file")
		}
		return nil
	}
}

func CertPoolFromCerts(certs ...*x509.Certificate) CertPoolProvider {
	return CertPoolFromCertPoolOptions([]CertPoolOption{
		CertPoolOptionFromCerts(certs...),
	})
}

func CertPoolOptionFromCerts(certs ...*x509.Certificate) CertPoolOption {
	return func(certPool *x509.CertPool) error {
		for _, cert := range certs {
			certPool.AddCert(cert)
		}
		return nil
	}
}

type configurer func(*tls.Config) error

func getClientCertificateParam(provider TLSCertProvider) configurer {
	return func(cfg *tls.Config) error {
		cfg.GetClientCertificate = func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			cert, err := provider()
			return &cert, err
		}
		return nil
	}
}

func certificatesParam(provider TLSCertProvider) configurer {
	return func(cfg *tls.Config) error {
		cert, err := provider()
		if err != nil {
			return fmt.Errorf("failed to load TLS certificate: %v", err)
		}
		cfg.Certificates = []tls.Certificate{cert}
		return nil
	}
}

func cipherSuitesParam(cipherSuites ...uint16) configurer {
	return func(cfg *tls.Config) error {
		cfg.CipherSuites = cipherSuites
		return nil
	}
}

func configureTLSConfig(tlsCfg *tls.Config, cfgs ...configurer) (*tls.Config, error) {
	for _, currCfg := range cfgs {
		if err := currCfg(tlsCfg); err != nil {
			return nil, err
		}
	}
	tlsCfg.BuildNameToCertificate()
	return tlsCfg, nil
}
