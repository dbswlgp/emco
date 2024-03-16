// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// CreateCertificateSigningRequest creates a certificate request
func CreateCertificateSigningRequest(info x509.CertificateRequest, pk interface{}) ([]byte, error) {
	csr, err := x509.CreateCertificateRequest(rand.Reader, &info, pk)
	if err != nil {
		return []byte{}, err
	}

	//Encode csr
	request := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csr,
	})

	return request, nil
}

// GeneratePrivateKey creates the rsa private key
func GeneratePrivateKey(keySize int) (*pem.Block, error) {
	key, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, errors.New("failed to generate the certitifcate signing key")
	}

	pemBlock, _ := pem.Decode([]byte(pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key)})))
	if pemBlock == nil {
		return nil, errors.New("failed to decode the private key")
	}

	return pemBlock, nil
}

// ParsePrivateKey parse the rsa private key
func ParsePrivateKey(der []byte) (*rsa.PrivateKey, error) {
	return x509.ParsePKCS1PrivateKey(der)
}
