// +build ignore

// package main writes the crypto material used by unit tests. Run using "go generate".
package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
)

const (
	cert1File         = "cert-1.pem"
	key1File          = "key-1.pem"
	cert2File         = "cert-2.pem"
	combinedCertsFile = "combined-certs.pem"
	certWithKeyFile   = "cert-with-key.pem"
)

func main() {
	cert1, key1 := newTestKeyPair(1, "test", "localhost")
	err := ioutil.WriteFile(cert1File, []byte(cert1), 0644)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(key1File, []byte(key1), 0644)
	if err != nil {
		panic(err)
	}
	cert2, _ := newTestKeyPair(2, "test", "test.com")
	err = ioutil.WriteFile(cert2File, []byte(cert2), 0644)
	if err != nil {
		panic(err)
	}
	combinedCerts := fmt.Sprintf("%s%s", cert1, cert2)
	err = ioutil.WriteFile(combinedCertsFile, []byte(combinedCerts), 0644)
	if err != nil {
		panic(err)
	}
	certWithKey := fmt.Sprintf("%s%s", cert1, key1)
	err = ioutil.WriteFile(certWithKeyFile, []byte(certWithKey), 0644)
	if err != nil {
		panic(err)
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
