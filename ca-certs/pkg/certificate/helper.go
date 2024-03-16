// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package certificate

import "crypto/x509"

// SignatureAlgorithm returns the x509 signature algorithm
func SignatureAlgorithm(algorithm string) x509.SignatureAlgorithm {
	switch algorithm {
	case "MD5WithRSA":
		return x509.MD5WithRSA
	case "SHA1WithRSA":
		return x509.SHA1WithRSA
	case "SHA256WithRSA":
		return x509.SHA256WithRSA
	case "SHA384WithRSA":
		return x509.SHA384WithRSA
	case "SHA512WithRSA":
		return x509.SHA512WithRSA
	case "ECDSAWithSHA1":
		return x509.ECDSAWithSHA1
	case "ECDSAWithSHA256":
		return x509.ECDSAWithSHA256
	case "ECDSAWithSHA384":
		return x509.ECDSAWithSHA384
	case "ECDSAWithSHA512":
		return x509.ECDSAWithSHA512
	case "SHA256WithRSAPSS":
		return x509.SHA256WithRSAPSS
	case "SHA384WithRSAPSS":
		return x509.SHA384WithRSAPSS
	case "SHA512WithRSAPSS":
		return x509.SHA512WithRSAPSS
	case "PureEd25519":
		return x509.PureEd25519
	default:
		return x509.UnknownSignatureAlgorithm
	}
}

// PublicKeyAlgorithm returns the x509 public key algorithm
func PublicKeyAlgorithm(algorithm string) x509.PublicKeyAlgorithm {
	switch algorithm {
	case "RSA":
		return x509.RSA
	case "DSA":
		return x509.DSA
	case "ECDSA":
		return x509.ECDSA
	case "Ed25519":
		return x509.Ed25519
	default:
		return x509.UnknownPublicKeyAlgorithm
	}
}
