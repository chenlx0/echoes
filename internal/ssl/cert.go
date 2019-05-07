package ssl

import (
	"crypto/tls"
)

// GetAllCertificates Read All certs files
func GetAllCertificates(domains []string, path string) ([]tls.Certificate, error) {
	certs := make([]tls.Certificate, 0)
	for _, d := range domains {
		pemPath := path + d + ".pem"
		keyPath := path + d + ".key"
		cert, err := tls.LoadX509KeyPair(pemPath, keyPath)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}
	return certs, nil
}

//
