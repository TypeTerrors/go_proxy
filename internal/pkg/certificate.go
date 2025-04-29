package pkg

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"prx/internal/models"
	"time"
)

func GenerateSelfSignedCertificate(fqdn string) (models.Cert, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return models.Cert{}, fmt.Errorf("error generatting private rsa key %w", err)

	}

	serialNumber, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		return models.Cert{}, fmt.Errorf("error generatting serial number tls template: %w", err)
	}

	tmpl := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   fqdn,
			Organization: []string{"nated corp"},
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{fqdn},
	}

	derCert, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		return models.Cert{}, fmt.Errorf("error creating certificate: %w", err)
	}

	certBuf := &bytes.Buffer{}
	if err := pem.Encode(certBuf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derCert,
	}); err != nil {
		return models.Cert{}, fmt.Errorf("error encoding certificate PEM: %w", err)
	}

	keyBuf := &bytes.Buffer{}
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	if err := pem.Encode(keyBuf, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	}); err != nil {
		return models.Cert{}, fmt.Errorf("error encoding key PEM: %w", err)
	}

	return models.Cert{
		Key: certBuf.Bytes(),
		Crt: keyBuf.Bytes(),
	}, nil
}
