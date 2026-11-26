package certgen

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"
)

type Pair struct {
	Cert    *x509.Certificate
	CertPEM []byte
	KeyPEM  []byte
}

func (p Pair) TLSCertificate() (tls.Certificate, error) {
	return tls.X509KeyPair(p.CertPEM, p.KeyPEM)
}

func NewCA(name string, notBefore, notAfter time.Time) (Pair, *ecdsa.PrivateKey, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return Pair{}, nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber:          serial(),
		Subject:               pkix.Name{CommonName: name},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}
	pair, err := sign(tmpl, tmpl, key, key)
	return pair, key, err
}

// Leaf issues a server certificate for hosts. With a nil parent it is self-signed.
func Leaf(hosts []string, notBefore, notAfter time.Time, parent *x509.Certificate, parentKey *ecdsa.PrivateKey) (Pair, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return Pair{}, err
	}
	tmpl := &x509.Certificate{
		SerialNumber: serial(),
		Subject:      pkix.Name{CommonName: hosts[0]},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			tmpl.IPAddresses = append(tmpl.IPAddresses, ip)
		} else {
			tmpl.DNSNames = append(tmpl.DNSNames, h)
		}
	}
	if parent == nil {
		return sign(tmpl, tmpl, key, key)
	}
	return sign(tmpl, parent, key, parentKey)
}

func sign(tmpl, parent *x509.Certificate, key, parentKey *ecdsa.PrivateKey) (Pair, error) {
	der, err := x509.CreateCertificate(rand.Reader, tmpl, parent, &key.PublicKey, parentKey)
	if err != nil {
		return Pair{}, err
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return Pair{}, err
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return Pair{}, err
	}
	return Pair{
		Cert:    cert,
		CertPEM: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		KeyPEM:  pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}),
	}, nil
}

func serial() *big.Int {
	n, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 62))
	return n
}
